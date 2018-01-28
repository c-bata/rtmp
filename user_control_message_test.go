package rtmp

import (
	"bytes"
	"testing"
)

func TestGenerateUserStreamBegin(t *testing.T) {
	expected := []byte{
		0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	actual, err := GenerateUserStreamBegin(0)
	if err != nil {
		t.Errorf("should be nil, but got %s", err)
	}
	if bytes.Compare(expected, actual) != 0 {
		t.Errorf("should be %#v, but got %#v", expected, actual)
	}
}
