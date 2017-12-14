package processing

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"github.com/golang/snappy"
	"github.com/pierrec/lz4"
)

type CompressionMode uint8

const (
	CompressionModeDefault CompressionMode = iota
	CompressionModeBestSpeed
	CompressionModeBestCompression
)

type CompressionType uint8

const (
	CompressionTypeSnappy CompressionType = iota
	CompressionTypeLZ4
	CompressionTypeGZip

	MaxStandardCompressionType = CompressionTypeGZip
)

type SnappyCompressorDecompressor struct{}

func (cd *SnappyCompressorDecompressor) WriteProcess(data []byte) ([]byte, error) {
	return snappy.Decode(nil, data)
}

func (cd *SnappyCompressorDecompressor) ReadProcess(data []byte) ([]byte, error) {
	return snappy.Encode(nil, data), nil
}

type LZ4CompressorDecompressor struct {
	writeBuffer bytes.Buffer
}

func (cd *LZ4CompressorDecompressor) WriteProcess(data []byte) ([]byte, error) {
	cd.writeBuffer.Reset()
	w := lz4.NewWriter(&cd.writeBuffer)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}

	b := cd.writeBuffer.Bytes()
	data = make([]byte, len(b))
	copy(data, b)

	return b, nil
}

func (cd *LZ4CompressorDecompressor) ReadProcess(data []byte) ([]byte, error) {
	r := lz4.NewReader(bytes.NewBuffer(data))
	return ioutil.ReadAll(r)
}

type GZipCompressorDecompressor struct {
	writeBuffer bytes.Buffer
}

func (cd *GZipCompressorDecompressor) WriteProcess(data []byte) ([]byte, error) {
	cd.writeBuffer.Reset()
	w := gzip.NewWriter(&cd.writeBuffer)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}

	b := cd.writeBuffer.Bytes()
	data = make([]byte, len(b))
	copy(data, b)

	return b, nil
}

func (cd *GZipCompressorDecompressor) ReadProcess(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

var (
	_ Processor = (*SnappyCompressorDecompressor)(nil)
	_ Processor = (*LZ4CompressorDecompressor)(nil)
	_ Processor = (*GZipCompressorDecompressor)(nil)
)
