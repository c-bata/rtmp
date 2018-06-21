package rtmp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/zhangpeihao/goamf"
)

const (
	WindowAcknowledgementSize = 2500000
	PeerBandWidth             = 2500000
)

type ConnectionState int

const (
	// StateInitialized means both the client and server are uninitialized.
	// The protocol version is sent during this stage.
	StateUninitialized ConnectionState = iota
	// StateVersionSent means both client and server are in the Version Sent state after the Uninitialized state.
	// The client is waiting for the packet S1 and the server is waiting for the packet C1.
	StateVersionSent
	// StateAckSent means the client and the server wait for S2 and C2 respectively.
	StateAckSent
	// StateHandshakeDone means the client and the server exchange messages.
	StateHandshakeDone
	// StateConnectResponseSent means that server has already sent the response for connect command message.
	StateConnectResponseSent
	// StateSentCreateStreamResponse means that server returns _result command message for createStream
	StateSentCreateStreamResponse
	// StatePublishingContent means that server is just receiving content.
	StatePublishingContent
)

type MessageType uint8

const (
	MessageSetChunkSize              MessageType = 1
	MessageAbort                                 = 2
	MessageAcknowledgement                       = 3
	MessageUserControl                           = 4
	MessageAcknowledgementWindowSize             = 5
	MessageSetPeerBandwidth                      = 6
	MessageAudio                                 = 8
	MessageVideo                                 = 9
	MessageDataAMF3                              = 15
	MessageSharedObjectAMF3                      = 16
	MessageCommandAMF3                           = 17
	MessageDataAMF0                              = 18
	MessageSharedObjectAMF0                      = 19
	MessageCommandAMF0                           = 20
	MessageAggregate                             = 22
)

// A Conn represents the RTMP connection and implements the RTMP protocol over net.Conn interface.
type conn struct {
	netconn    net.Conn
	server     *Server
	bufr       *bufio.Reader
	bufw       *bufio.Writer
	readbuf    []byte
	writebuf   []byte
	state      ConnectionState
	chunkSize  uint32
	streamName string
}

func (c *conn) serve() error {
	if err := c.handshake(); err != nil {
		c.server.logf("Handshaking Error: %s", err)
		c.netconn.Close()
		return err
	}

	for {
		if err := c.readChunk(); err == io.EOF {
			c.netconn.Close()
			return nil
		} else if err != nil {
			return err
		}
	}
}

//
// +-------------+                            +-------------+
// |    Client   |       TCP/IP Network       |    Server   |
// +-------------+             |              +-------------+
//        |                    |                     |
//  Uninitialized              |               Uninitialized
//        |         C0         |                     |
//        |------------------->|          C0         |
//        |                    |-------------------->|
//        |         C1         |                     |
//        |------------------->|          S0         |
//        |                    |<--------------------|
//        |                    |          S1         |
//  Version sent               |<--------------------|
//        |         S0         |                     |
//        |<-------------------|                     |
//        |         S1         |                     |
//        |<-------------------|               Version sent
//        |                    |          C1         |
//        |                    |-------------------->|
//        |         C2         |                     |
//        |------------------->|          S2         |
//        |                    |<--------------------|
//     Ack sent                |                  Ack Sent
//        |         S2         |                     |
//        |<-------------------|                     |
//        |                    |         C2          |
//        |                    |-------------------->|
//  Handshake Done             |               Handshake Done
//        |                    |                     |
//

func (c *conn) handshake() error {
	c.server.logf("Begin RTMP Handshake.")

	// << C0
	c0, err := readC0S0(c.bufr)
	if err != nil {
		c.server.logf("Read C0 error: %s", err)
		return err
	} else if c0.version > 3 {
		err = errors.New("unsupported rtmp version")
		c.server.logf("%s: %#v", err, c0.version)
		return err
	}
	c.server.logf("Receive a C0 chunk.")
	// >> S0
	s0 := newChunkC0S0()
	if _, err := c.bufw.Write(s0.Bytes()); err != nil {
		c.server.logf("Write S0 error: %s", err)
		return err
	}
	if err := c.bufw.Flush(); err != nil {
		c.server.logf("Flush S0 error: %s", err)
		return err
	}
	c.server.logf("Send a S0 chunk.")
	c.state = StateVersionSent

	// << C1
	c1, err := readC1S1(c.bufr)
	if err != nil {
		c.server.logf("Read C1 error: %s", err)
		return err
	}
	c.server.logf("Receive a C1 chunk.")
	// >> S1
	s1 := newChunkC1S1(0)
	if _, err := c.bufw.Write(s1.Bytes()); err != nil {
		c.server.logf("Write S1 error: %s", err)
		return err
	}
	if err := c.bufw.Flush(); err != nil {
		c.server.logf("Flush S1 error: %s", err)
		return err
	}
	c.server.logf("Send a S1 chunk.")
	c.state = StateAckSent

	// >> S2
	s2 := newChunkC2S2(c1)
	if _, err = c.bufw.Write(s2.Bytes()); err != nil {
		c.server.logf("Write S2 error: %s", err)
		return err
	}
	if err := c.bufw.Flush(); err != nil {
		c.server.logf("Flush S2 error: %s", err)
		return err
	}
	c.server.logf("Send a S2 chunk.")

	// << C2
	c2, err := readC2S2(c.bufr)
	if err != nil {
		c.server.logf("Read C2 error: %s", err)
		return err
	}
	c.server.logf("Receive a C2 chunk.")
	if bytes.Compare(c2.randomEcho, s1.randomBytes) != 0 {
		return errors.New("random echo doesn't match")
	}
	c.state = StateHandshakeDone
	return nil
}

func (c *conn) readChunk() error {
	header, err := readChunkHeader(c.bufr)
	if err == io.EOF {
		c.server.logf("Got EOF")
		return nil
	} else if err != nil {
		c.server.logf("Error while readChunkHeader: %s", err)
		return err
	}

	switch MessageType(header.MessageHeader.MessageTypeID) {
	case MessageSetChunkSize:
		//  0                   1                   2                   3
		//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |0|                   chunk size (31 bits)                      |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		if header.MessageHeader.MessageLength != 4 {
			return errors.New("the payload length of Set Chunk Size command should be 4")
		}
		payload := make([]byte, 4)
		_, err := io.ReadAtLeast(c.bufr, payload, int(4))
		if err != nil {
			return err
		}
		c.chunkSize = binary.BigEndian.Uint32(payload)
		c.server.logf("Set Chunk Size: %d", c.chunkSize)
		return nil
	case MessageAbort:
		//  0                   1                   2                   3
		//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |                   chunk stream id (32 bits)                   |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		csid := binary.BigEndian.Uint32(payload)
		c.server.logf("Abort Message: %d", csid)
		return nil
	case MessageAcknowledgement:
		//  0                   1                   2                   3
		//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |                    sequence number (4 bytes)                  |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		sequenceNumber := binary.BigEndian.Uint32(payload)
		c.server.logf("Acknowledgement Message: %d", sequenceNumber)
		return nil
	case MessageUserControl:
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		c.server.logf("User Control Message\n")
		c.server.logf("  payload : %#v\n", payload)
	case MessageAcknowledgementWindowSize:
		//  0                   1                   2                   3
		//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |              Acknowledgement Window size (4 bytes)            |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		ackWindowSize := binary.BigEndian.Uint32(payload)
		c.server.logf("WindowAcknowledgementSize Message: %d", ackWindowSize)
		return nil
	case MessageSetPeerBandwidth:
		//  0                   1                   2                   3
		//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |                   Acknowledgement Window size                 |
		// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		// |  Limit Type   |
		// +-+-+-+-+-+-+-+-+
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		ackWindowSize := binary.BigEndian.Uint32(payload[:4])
		limitType := payload[4]
		c.server.logf("SetPeerBandWidth Message: %d, %d", ackWindowSize, limitType)
		return nil
	case MessageAudio:
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		c.server.logf("Catch audio message")
	case MessageVideo:
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		c.server.logf("Catch video message")
	case MessageDataAMF3:
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		c.server.logf("Catch DataMessage(AMF3)")
	case MessageCommandAMF3:
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		c.server.logf("Catch AMF3 Command Message")
	case MessageSharedObjectAMF3:
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		c.server.logf("Catch SharedObjectMessage(AMF0)")
	case MessageDataAMF0:
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		c.server.logf("Catch DataMessage(AMF0)")
	case MessageCommandAMF0:
		c.server.logf("Catch AMF0 Command Message")
		err = c.handleCommandMessageAMF0(header)
		if err != nil {
			return err
		}
	case MessageSharedObjectAMF0:
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		c.server.logf("Catch SharedObjectMessage(AMF0)")
	case MessageAggregate:
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		c.server.logf("Catch AggregateMessage")
	default:
		c.server.logf("Catch unknown message type id: %d", header.MessageHeader.MessageTypeID)
		c.server.logf("%#v", header.BasicHeader)
		c.server.logf("%#v", header.MessageHeader)
		payload := make([]byte, header.MessageHeader.MessageLength)
		_, err := io.ReadAtLeast(c.bufr, payload, int(header.MessageHeader.MessageLength))
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (c *conn) handleCommandMessageAMF0(header *ChunkHeader) error {
	ml := header.MessageHeader.MessageLength
	payload := make([]byte, ml)
	_, err := io.ReadAtLeast(c.bufr, payload, int(ml))
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(payload)
	commandName, err := amf.ReadString(buf)
	if err != nil {
		return err
	}
	transactionID, err := amf.ReadDouble(buf)
	if err != nil {
		return err
	}

	switch commandName {
	case "connect":
		// Maybe this is bug of ffmpeg(librtmp). 0xc3 is in tcp byte stream.
		// Cause of this strange byte, the payload length does not equals with message length in Message Header.
		for i := range payload {
			if payload[i] == 0xc3 {
				if _, err := c.bufr.ReadByte(); err != nil {
					return err
				}
				break
			}
		}

		c.server.logf("Receive connect command message (transactionID: %f).", transactionID)
		// Send window acknowledgement
		was, err := GenerateWindowAcknowledgementSizeChunk(WindowAcknowledgementSize)
		if err != nil {
			return err
		}
		_, err = c.bufw.Write(was)
		if err != nil {
			return err
		}

		// Send peer bandwidth
		pbw, err := GenerateSetPeerBandwidthChunk(PeerBandWidth, PeerBandwidthLimitTypeDynamic)
		if err != nil {
			return err
		}
		_, err = c.bufw.Write(pbw)
		if err != nil {
			return err
		}

		// Send User Control Message Events - StreamBegin
		usb, err := GenerateUserStreamBegin(0)
		if err != nil {
			return err
		}
		_, err = c.bufw.Write(usb)
		if err != nil {
			return err
		}

		// Set Chunk Size (size = 4096)
		scs, err := GenerateSetChunkSize(4096)
		if err != nil {
			return err
		}
		_, err = c.bufw.Write(scs)
		if err != nil {
			return err
		}

		// Command Message: _result (connect)
		x, err := GenerateConnectResult(transactionID)
		if err != nil {
			return err
		}
		_, err = c.bufw.Write(x)
		if err != nil {
			return err
		}
		err = c.bufw.Flush()
		if err != nil {
			return err
		}
		c.state = StateConnectResponseSent
	case "releaseStream":
		c.server.logf("Receive a releaseStream command (transactionID: %f).", transactionID)
		if c.state < StateConnectResponseSent {
			return errors.New("connect response should be sent before receiving a releaseStream command")
		}
		//_, err := amf.ReadValue(buf) // Returns null-type
		//if err != nil {
		//	return err
		//}
		//streamName, err := amf.ReadString(buf) // Should return streamName (string)
		//if err != nil {
		//	return err
		//}
	case "FCPublish":
		_, err := amf.ReadValue(buf) // Returns null-type
		if err != nil {
			return err
		}
		streamName, err := amf.ReadString(buf) // Should return streamName(string)
		if err != nil {
			return err
		}
		c.server.logf("Receive FCPublish command message (transactionID: %f, streamName: %s).", transactionID, streamName)

		msg, err := GenerateOnFCPublishMessage(transactionID, streamName)
		if err != nil {
			return err
		}
		_, err = c.bufw.Write(msg)
		if err != nil {
			return err
		}
		err = c.bufw.Flush()
		if err != nil {
			return err
		}
		c.streamName = streamName
		return nil
	case "createStream":
		c.server.logf("Catch createStream command message - (transactionID: %f)", transactionID)
		msg, err := CreateStreamResponseMessage(transactionID)
		if err != nil {
			return err
		}
		_, err = c.bufw.Write(msg)
		if err != nil {
			return err
		}
		err = c.bufw.Flush()
		if err != nil {
			return err
		}
		c.state = StateSentCreateStreamResponse
		return nil
	case "publish":
		c.server.logf("Catch publish command message - (transactionID: %f)", transactionID)
		if c.state < StateSentCreateStreamResponse {
			return errors.New("connect response should be sent before receiving a releaseStream command")
		} else if c.state == StatePublishingContent {
			c.server.logf("Catch publish command message in StateSentCreateStreamResponse")
			return nil
		}
		// returns user control message(stream begin)
		var msg []byte
		msg, err = GenerateUserStreamBegin(1)
		if err != nil {
			return err
		}
		_, err = c.bufw.Write(msg)
		if err != nil {
			return err
		}
		msg, err = CreateOnStatusPublishStartMessage(transactionID, c.streamName)
		if err != nil {
			return err
		}
		_, err = c.bufw.Write(msg)
		if err != nil {
			return err
		}

		err = c.bufw.Flush()
		if err != nil {
			return err
		}
		c.state = StatePublishingContent
	}
	return nil
}
