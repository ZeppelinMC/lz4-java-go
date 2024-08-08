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
		var buffer = buffers.Get().([]byte)
		var bufferMaxLen = int(max(compressedLength, decompressedLength))

		if len(buffer) < bufferMaxLen {
			buffer = make([]byte, bufferMaxLen)
		}
		defer buffers.Put(buffer)

		if _, err := data.Read(buffer[:compressedLength]); err != nil {
			return nil, err
		}

		_, err = lz4.UncompressBlock(buffer, buffer)

		return buffer, err
	case methodUncompressed:
		var buffer = buffers.Get().([]byte)

		if len(buffer) < compressedLength {
			buffer = make([]byte, compressedLength)
		}
		defer buffers.Put(buffer)
		
		if _, err := data.Read(buffer[:compressedLength]); err != nil {
			return nil, err
		}

		return buffer[:compressedLength], nil
	default:
		return nil, fmt.Errorf("unknown compression method %d", compressionMethod)
	}
}

var buffers = sync.Pool{
	New: func() any { return make([]byte, 0) },
}

const magic = "LZ4Block"
const (
	methodUncompressed = 1 << (iota + 4)
	methodLZ4
)


