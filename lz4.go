package lz4

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/pierrec/lz4/v4"
)

// Decompress an lz4-java block. The data returned is only safe to use until the next operation
func Decompress(data io.Reader) ([]byte, error) {
	var header [21]byte
	_, err := data.Read(header[:])
	if err != nil {
		return nil, err
	}
	magicValue := string(header[:8])
	if magicValue != magic {
		return nil, fmt.Errorf("invalid magic value")
	}

	compressedLength := binary.LittleEndian.Uint32(header[9:13])
	decompressedLength := binary.LittleEndian.Uint32(header[13:17])

	token := header[8]
	compressionMethod := token & 0xf0
	switch compressionMethod {
	case methodLZ4:
		var compressed = make([]byte, compressedLength)

		var decompressed = decompressedBuffer.Get().([]byte)
		if len(decompressed) < int(decompressedLength) {
			decompressed = make([]byte, decompressedLength)
		}
		defer decompressedBuffer.Put(decompressed)

		if _, err := data.Read(compressed); err != nil {
			return nil, err
		}

		_, err = lz4.UncompressBlock(compressed, decompressed)

		return decompressed, err
	case methodUncompressed:
		var compressed = make([]byte, compressedLength)
		if _, err := data.Read(compressed); err != nil {
			return nil, err
		}

		return compressed, nil
	default:
		return nil, fmt.Errorf("unknown compression method %d", compressionMethod)
	}
}

var decompressedBuffer = sync.Pool{
	New: func() any { return make([]byte, 0) },
}

const magic = "LZ4Block"
const (
	methodUncompressed = 1 << (iota + 4)
	methodLZ4
)

