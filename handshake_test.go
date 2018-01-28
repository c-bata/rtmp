package rtmp

import (
	"bufio"
	"bytes"
	"testing"
)

func TestGenS0(t *testing.T) {
	s0 := newChunkC0S0()
	if bytes.Compare(s0.Bytes(), []byte{0x03}) != 0 {
		t.Errorf("server should respond with 3, but got %#v", s0.Bytes())
	}
}

func TestReadS0(t *testing.T) {
	in := bufio.NewReader(bytes.NewBuffer([]byte{0x3}))
	s1, err := readC0S0(in)
	if err != nil {
		t.Errorf("err should be nil, but got %s", err)
	}
	if s1.version != 0x03 {
		t.Errorf("should respond with 3, but got %d", s1.version)
	}
}

func TestS1C1(t *testing.T) {
	chunk1 := newChunkC1S1(0)
	expected := make([]byte, 8)
	if bytes.HasSuffix(chunk1.Bytes(), expected) {
		t.Errorf("S1 chunk should starts with 8 byte of 0, but got %#v", chunk1.Bytes()[:8])
	}

	in := bufio.NewReader(bytes.NewBuffer(chunk1.Bytes()))
	chunk2, err := readC1S1(in)
	if err != nil {
		t.Errorf("err should be nil, but got %s", err)
	}
	if 0 != chunk2.time {
		t.Errorf("Should be 0, but got %d", chunk2.time)
	}
	if bytes.Compare(chunk1.randomBytes, chunk2.randomBytes) != 0 {
		t.Errorf("Should be %#v, but got %#v", chunk1.randomBytes, chunk2.randomBytes)
	}
}
