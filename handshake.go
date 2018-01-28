package rtmp

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"io"
	"time"
)

// The C0 and S0 packets are a single octet, treated as a single 8-bit integer field:
//
//  0 1 2 3 4 5 6 7
// +-+-+-+-+-+-+-+-+
// |    version    |
// +-+-+-+-+-+-+-+-+
//

// In RTMP 1.0, A server that does not recognize the client's requested version SHOULD respond with 3.
func newChunkC0S0() *chunkC0S0 {
	return &chunkC0S0{version: 3}
}

type chunkC0S0 struct {
	version uint8
}

func (c *chunkC0S0) Bytes() []byte {
	return []byte{c.version}
}

func readC0S0(br *bufio.Reader) (*chunkC0S0, error) {
	ver, err := br.ReadByte()
	if err != nil {
		return nil, err
	}
	return &chunkC0S0{
		version: ver,
	}, nil
}

// The C1 and S1 packets are 1536 (=32*48) octets long, consisting of the following fields:
//
//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         time (4 bytes)                        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         zero (4 bytes)                        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         random bytes                          |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         random bytes                          |
// |                            (cont)                             |
// |                             ....                              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//

func newChunkC1S1(time1 uint32) *chunkC1S1 {
	randomBytes := make([]byte, 1528)
	rand.Read(randomBytes)
	return &chunkC1S1{
		time:        time1,
		randomBytes: randomBytes,
	}
}

type chunkC1S1 struct {
	time        uint32
	randomBytes []byte // 1528 bytes
}

func (c *chunkC1S1) Bytes() []byte {
	chunk := make([]byte, 1536)
	binary.BigEndian.PutUint32(chunk[:4], c.time)
	copy(chunk[8:], c.randomBytes)
	return chunk
}

func readC1S1(br *bufio.Reader) (*chunkC1S1, error) {
	chunk := make([]byte, 1536)
	_, err := io.ReadAtLeast(br, chunk, 1536)
	if err != nil {
		return nil, err
	}

	return &chunkC1S1{
		time:        binary.BigEndian.Uint32(chunk[:4]),
		randomBytes: chunk[8:],
	}, nil
}

// The C2 and S2 packets are 1536 (=32*48) octets long, and nearly an echo of S1 and C1 (respectively), consisting of the following fields:
//
//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         time (4 bytes)                        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                        time2 (4 bytes)                        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         random echo                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         random echo                           |
// |                            (cont)                             |
// |                             ....                              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//

func newChunkC2S2(c *chunkC1S1) *chunkC2S2 {
	now := uint32(time.Now().UnixNano() / int64(time.Millisecond))
	return &chunkC2S2{
		time:       c.time,
		time2:      now,
		randomEcho: c.randomBytes,
	}
}

type chunkC2S2 struct {
	time       uint32
	time2      uint32
	randomEcho []byte // 1528 bytes
}

func (c *chunkC2S2) Bytes() []byte {
	chunk := make([]byte, 1536)
	binary.BigEndian.PutUint32(chunk[:4], c.time)
	binary.BigEndian.PutUint32(chunk[4:8], c.time2)
	copy(chunk[8:], c.randomEcho)
	return chunk
}

func readC2S2(br *bufio.Reader) (*chunkC2S2, error) {
	chunk := make([]byte, 1536)
	_, err := io.ReadAtLeast(br, chunk, 1536)
	if err != nil {
		return nil, err
	}
	time1 := chunk[:4]
	time2 := chunk[4:8]
	return &chunkC2S2{
		time:       binary.BigEndian.Uint32(time1),
		time2:      binary.BigEndian.Uint32(time2),
		randomEcho: chunk[8:],
	}, nil
}
