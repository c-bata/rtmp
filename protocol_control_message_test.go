package rtmp

import (
	"bytes"
	"testing"
)

func TestGenerateWindowAcknowledgementSizeChunk(t *testing.T) {
	expected := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0x25, 0xa0}
	actual, err := GenerateWindowAcknowledgementSizeChunk(2500000)
	if err != nil {
		t.Errorf("should be nil, but got %s", err)
	}
	if bytes.Compare(actual, expected) != 0 {
		t.Errorf("should be %#v, but got %#v", expected, actual)
	}
}

func TestGenerateSetPeerBandwidthChunk(t *testing.T) {
	expected := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0x25, 0xa0, 0x02}
	actual, err := GenerateSetPeerBandwidthChunk(2500000, 2)
	if err != nil {
		t.Errorf("should be nil, but got %s", err)
	}
	if bytes.Compare(expected, actual) != 0 {
		t.Errorf("should be %#v, but got %#v", expected, actual)
	}
}

func TestGenerateSetChunkSize(t *testing.T) {
	expected := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00}
	actual, err := GenerateSetChunkSize(4096)
	if err != nil {
		t.Errorf("should be nil, but got %s", err)
	}
	if bytes.Compare(expected, actual) != 0 {
		t.Errorf("should be %#v, but got %#v", expected, actual)
	}
}
