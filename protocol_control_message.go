package rtmp

import "encoding/binary"

type PeerBandwidthLimitType int

const (
	PeerBandwidthLimitTypeHard    PeerBandwidthLimitType = 0
	PeerBandwidthLimitTypeSoft                           = 1
	PeerBandwidthLimitTypeDynamic                        = 2
)

func generateProtocolControlMessageHeader(messageTypeID uint8, messageLength uint32) *ChunkHeader {
	return &ChunkHeader{
		BasicHeader: &BasicHeader{
			FMT:           0,
			ChunkStreamID: 2,
		},
		MessageHeader: &MessageHeader{
			MessageLength: messageLength,
			MessageTypeID: messageTypeID,
		},
	}
}

func GenerateSetChunkSize(chunkSize uint32) ([]byte, error) {
	ch := generateProtocolControlMessageHeader(1, 4)
	x, err := genChunkHeader(ch)
	if err != nil {
		return []byte{}, err
	}

	y := make([]byte, 4)
	binary.BigEndian.PutUint32(y, chunkSize)
	y[0] = y[0] & 0x7f
	return append(x, y...), nil
}

func GenerateWindowAcknowledgementSizeChunk(size uint32) ([]byte, error) {
	ch := generateProtocolControlMessageHeader(5, 4)
	x, err := genChunkHeader(ch)
	if err != nil {
		return []byte{}, err
	}

	y := make([]byte, 4)
	binary.BigEndian.PutUint32(y, size)
	return append(x, y...), nil
}

func GenerateSetPeerBandwidthChunk(size uint32, limitType uint8) ([]byte, error) {
	ch := generateProtocolControlMessageHeader(6, 5)
	x, err := genChunkHeader(ch)
	if err != nil {
		return []byte{}, err
	}

	y := make([]byte, 5)
	binary.BigEndian.PutUint32(y[:4], size)
	y[4] = byte(limitType)
	return append(x, y...), nil
}
