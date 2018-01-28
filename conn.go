package rtmp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"time"
)

// A Conn represents the RTMP connection and implements the RTMP protocol over net.Conn interface.
type conn struct {
	netconn    net.Conn
	server     *Server
	bufr       *bufio.Reader
	bufw       *bufio.Writer
	readbuf    []byte
	writebuf   []byte
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
	return nil
}

func (c *conn) readChunk() error {
	time.Sleep(1)
	return nil
}
