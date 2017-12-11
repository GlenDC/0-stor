package storage

import (
	"github.com/templexxx/reedsolomon"
	"github.com/zero-os/0-stor/client/datastor"
)

// NewDistributedStorage creates a new DistributedStorage,
// using the given Cluster and default ReedSolomonEncoderDecoder as internal DistributedEncoderDecoder.
// See `DistributedStorage` `DistributedEncoderDecoder` for more information.
func NewDistributedStorage(cluster datastor.Cluster, k, m, jobCount int) (*DistributedStorage, error) {
	dec, err := NewReedSolomonEncoderDecoder(k, m)
	if err != nil {
		return nil, err
	}
	return NewDistributedStorageWithEncoderDecoder(cluster, dec, jobCount), nil
}

// NewDistributedStorageWithEncoderDecoder creates a new DistributedStorage,
// using the given Cluster and DistributedEncoderDecoder.
// See `DistributedStorage` `DistributedEncoderDecoder` for more information.
func NewDistributedStorageWithEncoderDecoder(cluster datastor.Cluster, dec DistributedEncoderDecoder, jobCount int) *DistributedStorage {
	if cluster == nil {
		panic("no Cluster given")
	}
	if dec == nil {
		panic("no DistributedEncoderDecoder given")
	}

	if jobCount < 1 {
		jobCount = DefaultJobCount
	}

	return &DistributedStorage{
		cluster:  cluster,
		dec:      dec,
		jobCount: jobCount,
	}
}

// DistributedStorage defines a storage implementation,
// which splits and distributes data over a secure amount of shards,
// rather than just writing it to a single shard as it is.
// This to provide protection against data loss when one of the used shards drops.
//
// By default the erasure code algorithms as implemented in
// the github.com/templexxx/reedsolomon library are used,
// and wrapped by the default ReedSolomonEncoderDecoder type.
// When using this default distributed encoder-decoder,
// you need to provide at least 2 shards (1 data- and 1 parity- shard).
//
// When creating a DistributedStorage you can also pass in your
// own DistributedEncoderDecoder should you not be satisfied with the default implementation.
type DistributedStorage struct {
	cluster  datastor.Cluster
	dec      DistributedEncoderDecoder
	jobCount int
}

// Write implements storage.Storage.Write
func (ds *DistributedStorage) Write(object datastor.Object) (StorageConfig, error) {
	panic("TODO")
}

// Read implements storage.Storage.Read
func (ds *DistributedStorage) Read(cfg StorageConfig) (datastor.Object, error) {
	panic("TODO")
}

// Repair implements storage.Storage.Repair
func (ds *DistributedStorage) Repair(cfg StorageConfig) (StorageConfig, error) {
	panic("TODO")
}

// Close implements storage.Storage.Close
func (ds *DistributedStorage) Close() error {
	return ds.cluster.Close()
}

// DistributedEncoderDecoder is the type used internally to
// read and write the data of objects, read and written using the DistributedStorage.
type DistributedEncoderDecoder interface {
	// Encode object data into multiple (distributed) parts,
	// such that those parts can be reconstructed when the data has to be read again.
	Encode(data []byte) (parts [][]byte, err error)
	// Decode the different parts back into the original data slice,
	// as it was given in the original Encode call.
	Decode(parts [][]byte, dataSize int) (data []byte, err error)
}

// NewReedSolomonEncoderDecoder creates a new ReedSolomonEncoderDecoder.
// See `ReedSolomonEncoderDecoder` for more information.
func NewReedSolomonEncoderDecoder(k, m int) (*ReedSolomonEncoderDecoder, error) {
	if k < 1 {
		panic("k (data shard count) has to be at least 1")
	}
	if m < 1 {
		panic("m (parity shard count) has to be at least 1")
	}

	er, err := reedsolomon.New(k, m)
	if err != nil {
		return nil, err
	}
	return &ReedSolomonEncoderDecoder{
		k:  k,
		m:  m,
		er: er,
	}, nil
}

// ReedSolomonEncoderDecoder implements the DistributedEncoderDecoder,
// using the erasure encoding library github.com/templexxx/reedsolomon.
//
// This implementation is also used as the default DistributedEncoderDecoder
// for the DistributedStorage storage type.
type ReedSolomonEncoderDecoder struct {
	k, m int                         // data and parity count
	er   reedsolomon.EncodeReconster // encoder  & decoder
}

// Encode implements DistributedEncoderDecoder.Encode
func (rs *ReedSolomonEncoderDecoder) Encode(data []byte) ([][]byte, error) {
	if len(data) == 0 {
		panic("no data given to encode")
	}

	parts := rs.splitData(data)
	parities := reedsolomon.NewMatrix(rs.m, len(parts[0]))
	parts = append(parts, parities...)
	err := rs.er.Encode(parts)
	return parts, err
}

// Decode implements DistributedEncoderDecoder.Decode
func (rs *ReedSolomonEncoderDecoder) Decode(parts [][]byte, dataSize int) ([]byte, error) {
	if len(parts) == 0 {
		panic("no parts given to decode")
	}

	for _, part := range parts {
		if len(part) == 0 {
			panic("cannot decode a part which is empty")
		}
	}

	if err := rs.er.ReconstructData(parts); err != nil {
		return nil, err
	}

	var (
		data   = make([]byte, dataSize)
		offset int
	)
	for i := 0; i < rs.k; i++ {
		copy(data[offset:], parts[i])
		offset += len(parts[i])
		if offset >= dataSize {
			break
		}
	}
	return data, nil
}

func (rs *ReedSolomonEncoderDecoder) splitData(data []byte) [][]byte {
	data = rs.padIfNeeded(data)
	chunkSize := len(data) / rs.k
	chunks := make([][]byte, rs.k)

	for i := 0; i < rs.k; i++ {
		chunks[i] = data[i*chunkSize : (i+1)*chunkSize]
	}
	return chunks
}

func (rs *ReedSolomonEncoderDecoder) padIfNeeded(data []byte) []byte {
	padLen := rs.getPadLen(len(data))
	if padLen == 0 {
		return data
	}

	pad := make([]byte, padLen)
	return append(data, pad...)
}

func (rs *ReedSolomonEncoderDecoder) getPadLen(dataLen int) int {
	const padFactor = 256
	maxPadLen := rs.k * padFactor
	mod := dataLen % maxPadLen
	if mod == 0 {
		return 0
	}
	return maxPadLen - mod
}

var (
	_ Storage = (*DistributedStorage)(nil)

	_ DistributedEncoderDecoder = (*ReedSolomonEncoderDecoder)(nil)
)
