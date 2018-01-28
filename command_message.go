package rtmp

import (
	"bytes"
	"fmt"
	"math"

	"github.com/zhangpeihao/goamf"
)

type CommandCode string

const (
	CodeNetConnectAppShutdown   CommandCode = "NetConnection.Connect.AppShutdown"
	CodeNetConnectClosed                    = "NetConnection.Connect.Closed"
	CodeNetConnectFailed                    = "NetConnection.Connect.Failed"
	CodeNetConnectIdleTimeout               = "NetConnection.Connect.IdleTimeout"
	CodeNetConnectInvalidApp                = "NetConnection.Connect.InvalidApp"
	CodeNetConnectNetworkChange             = "NetConnection.Connect.NetworkChange"
	CodeNetConnectRejected                  = "NetConnection.Connect.Rejected"
	CodeNetConnectSuccess                   = "NetConnection.Connect.Success"
)

type CommandLevel string

const (
	CommandLevelStatus CommandLevel = "status"
	CommandLevelError               = "error"
)

type ResultCommand struct {
	Name          string
	TransactionID float64
	Properties    map[string]interface{}
	Information   map[string]interface{}
}

func (rc *ResultCommand) Bytes() []byte {
	buf := new(bytes.Buffer)
	amf.WriteValue(buf, rc.Name)
	amf.WriteValue(buf, rc.TransactionID)
	amf.WriteValue(buf, rc.Properties)
	amf.WriteValue(buf, rc.Information)
	return buf.Bytes()
}

func GenerateConnectResult(transactionID float64) ([]byte, error) {
	cmd := &ResultCommand{
		Name:          "_result",
		TransactionID: transactionID,
		Properties: map[string]interface{}{
			"fmsVer":       "FMS/3,5,7,7009",
			"capabilities": 31,
			"mode":         1,
		},
		Information: map[string]interface{}{
			"code":        CodeNetConnectSuccess,
			"description": "Connection succeeded.",
			"data": map[string]interface{}{
				"version": "3,5,7,7009",
			},
			"objectEncoding": 0,
			"level":          CommandLevelStatus,
		},
	}
	payload := cmd.Bytes()

	ch := &ChunkHeader{
		BasicHeader: &BasicHeader{
			FMT:           0,
			ChunkStreamID: 3,
		},
		MessageHeader: &MessageHeader{
			Timestamp:       0,
			MessageLength:   uint32(len(payload)),
			MessageTypeID:   20,
			MessageStreamID: 0,
		},
	}
	header, err := genChunkHeader(ch)
	if err != nil {
		return []byte{}, err
	}
	x := append(header, payload...)
	return x, nil
}

func GenerateOnFCPublishMessage(transactionID float64, streamName string) ([]byte, error) {
	buf := new(bytes.Buffer)
	amf.WriteValue(buf, "onFCPublish")
	amf.WriteValue(buf, transactionID)
	amf.WriteValue(buf, nil)
	amf.WriteValue(buf, 1)
	amf.WriteValue(buf, map[string]interface{}{
		"level":       "status",
		"code":        "NetStream.Publish.Start",
		"description": fmt.Sprintf("FCPublish to stream %s.", streamName),
	})
	payload := buf.Bytes()

	ch := &ChunkHeader{
		BasicHeader: &BasicHeader{
			FMT:           0,
			ChunkStreamID: 3,
		},
		MessageHeader: &MessageHeader{
			Timestamp:       0,
			MessageLength:   uint32(len(payload)),
			MessageTypeID:   20,
			MessageStreamID: 0,
		},
	}
	header, err := genChunkHeader(ch)
	if err != nil {
		return []byte{}, err
	}
	x := append(header, payload...)
	return x, nil
}

type CreateStreamCommand struct {
	Name          string
	TransactionID float64
	Properties    map[string]interface{}
	Message       map[string]interface{}
}

func (c *CreateStreamCommand) Bytes() []byte {
	buf := new(bytes.Buffer)
	amf.WriteValue(buf, c.Name)
	amf.WriteValue(buf, c.TransactionID)
	amf.WriteValue(buf, nil)
	amf.WriteValue(buf, 1)
	return buf.Bytes()
}

func CreateStreamResponseMessage(transactionID float64) ([]byte, error) {
	cmd := &CreateStreamCommand{
		Name:          "_result",
		TransactionID: transactionID,
	}
	payload := cmd.Bytes()
	ch := &ChunkHeader{
		BasicHeader: &BasicHeader{
			FMT:           0,
			ChunkStreamID: 3,
		},
		MessageHeader: &MessageHeader{
			Timestamp:       0,
			MessageLength:   uint32(len(payload)),
			MessageTypeID:   20,
			MessageStreamID: 0,
		},
	}
	header, err := genChunkHeader(ch)
	if err != nil {
		return []byte{}, err
	}
	x := append(header, payload...)
	return x, nil
}

type NetStreamStatusMessage struct {
	Name          string
	TransactionID float64
	InfoObject    map[string]interface{}
}

func (m *NetStreamStatusMessage) Bytes() []byte {
	buf := new(bytes.Buffer)
	amf.WriteValue(buf, m.Name)
	amf.WriteValue(buf, m.TransactionID)
	amf.WriteValue(buf, nil)
	amf.WriteValue(buf, m.InfoObject)
	return buf.Bytes()
}

func CreateOnStatusPublishStartMessage(transactionID float64, streamName string) ([]byte, error) {
	cmd := &NetStreamStatusMessage{
		Name:          "onStatus",
		TransactionID: transactionID,
		InfoObject: map[string]interface{}{
			"code":        "NetStream.Publish.Start",
			"description": fmt.Sprintf("Publishing %s.", streamName),
			"level":       "status",
		},
	}
	payload := cmd.Bytes()

	ch := &ChunkHeader{
		BasicHeader: &BasicHeader{
			FMT:           0,
			ChunkStreamID: 3,
		},
		MessageHeader: &MessageHeader{
			Timestamp:       0,
			MessageLength:   uint32(len(payload)),
			MessageTypeID:   20,
			MessageStreamID: uint32(math.Pow(2, 4*6)),
		},
	}
	header, err := genChunkHeader(ch)
	if err != nil {
		return []byte{}, err
	}
	x := append(header, payload...)
	return x, nil
}
