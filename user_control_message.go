package rtmp

import "encoding/binary"

func generateUserControlMessageHeader(messageLength uint32) *ChunkHeader {
	return &ChunkHeader{
		BasicHeader: &BasicHeader{
			FMT:           0,
			ChunkStreamID: 2,
		},
		MessageHeader: &MessageHeader{
			MessageLength: messageLength,
			MessageTypeID: 4,
		},
	}
}

func GenerateUserStreamBegin(streamID uint32) ([]byte, error) {
	var (
		eventType  uint16 = 0
		messageLen uint32 = 6
	)

	ch := generateUserControlMessageHeader(messageLen)
	x, err := genChunkHeader(ch)
	if err != nil {
		return []byte{}, err
	}

	y := make([]byte, messageLen)
	binary.BigEndian.PutUint16(y[:2], eventType)
	binary.BigEndian.PutUint32(y[2:], streamID)
	return append(x, y...), nil
}
