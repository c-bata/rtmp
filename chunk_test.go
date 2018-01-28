package rtmp

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

func TestReadBasicHeader(t *testing.T) {
	header := []byte{0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0xb8, 0x14, 0x00, 0x00, 0x00, 0x00}
	in := bufio.NewReader(bytes.NewBuffer(header))

	actual, err := readBasicHeader(in)
	if err != nil {
		t.Errorf("Should be nil, but got %s", err)
	}

	expected := &BasicHeader{
		FMT:           0,
		ChunkStreamID: 3,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Should be %#v, but got %#v", actual, expected)
	}
}

func TestGenerateBasicHeader(t *testing.T) {
	expected := []byte{0x03}
	bh := &BasicHeader{
		FMT:           0,
		ChunkStreamID: 3,
	}
	actual, err := genBasicHeader(bh)
	if err != nil {
		t.Errorf("Should be nil, but got %s", err)
	}
	if bytes.Compare(actual, expected) != 0 {
		t.Errorf("Should be %#v, but got %#v", actual, expected)
	}
}

func TestReadMessageHeader(t *testing.T) {
	header := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0xb8, 0x14, 0x00, 0x00, 0x00, 0x00}
	in := bufio.NewReader(bytes.NewBuffer(header))
	actual, err := readMessageHeader(in, 0)
	if err != nil {
		t.Errorf("Should be nil, but got %s", err)
	}
	expected := &MessageHeader{
		Timestamp:       0,
		MessageLength:   184,
		MessageTypeID:   20,
		MessageStreamID: 0,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Should be %#v, but got %#v", actual, expected)
	}
}

func TestGenerateMessageHeader(t *testing.T) {
	expected := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0xb8, 0x14, 0x00, 0x00, 0x00, 0x00}
	mh := &MessageHeader{
		Timestamp:       0,
		MessageLength:   184,
		MessageTypeID:   20,
		MessageStreamID: 0,
	}
	actual, err := genMessageHeader(mh, 0)
	if err != nil {
		t.Errorf("Should be nil, but got %s", err)
	}
	if bytes.Compare(actual, expected) != 0 {
		t.Errorf("Should be %#v, but got %#v", actual, expected)
	}
}

func TestReadChunkHeader(t *testing.T) {
	// FMT (2 bits)                = 00
	// chunk stream id (6 bits)    = 000011
	// timestamp (3bytes)          = 0000 0000 0000 0000 0000 0000
	// message length (3 bytes)    = 0000 0000 0000 0000 1011 1000
	// message type id (1 bytes)   = 0001 0100
	// message stream id (4 bytes) = 0000 0000 0000 0000 0000 0000 0000 0000
	header := []byte{0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0xb8, 0x14, 0x00, 0x00, 0x00, 0x00}
	in := bufio.NewReader(bytes.NewBuffer(header))
	actual, err := readChunkHeader(in)
	if err != nil {
		t.Errorf("Should be nil, but got %s", err)
	}
	expected := &ChunkHeader{
		BasicHeader: &BasicHeader{
			FMT:           0,
			ChunkStreamID: 3,
		},
		MessageHeader: &MessageHeader{
			Timestamp:       0,
			MessageLength:   184,
			MessageTypeID:   20,
			MessageStreamID: 0,
		},
		ExtendedTimestamp: 0,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Should be %#v, but got %#v", expected, actual)
	}
}

func TestGenerateChunkHeader(t *testing.T) {
	expected := []byte{0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0xb8, 0x14, 0x00, 0x00, 0x00, 0x00}
	ch := &ChunkHeader{
		BasicHeader: &BasicHeader{
			FMT:           0,
			ChunkStreamID: 3,
		},
		MessageHeader: &MessageHeader{
			Timestamp:       0,
			MessageLength:   184,
			MessageTypeID:   20,
			MessageStreamID: 0,
		},
	}
	actual, err := genChunkHeader(ch)
	if err != nil {
		t.Errorf("Should be nil, but got %s", err)
	}
	if bytes.Compare(actual, expected) != 0 {
		t.Errorf("Should be %#v, but got %#v", actual, expected)
	}
}
