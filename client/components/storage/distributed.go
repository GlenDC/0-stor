package storage

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"

	log "github.com/Sirupsen/logrus"
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
	parts, err := ds.dec.Encode(object.Data)
	if err != nil {
		return StorageConfig{}, err
	}

	group, ctx := errgroup.WithContext(context.Background())

	jobCount := ds.jobCount
	partsCount := len(parts)
	if jobCount > partsCount {
		jobCount = partsCount
	}

	// sends each part to an available worker goroutine,
	// which tries to store it in a random shard.
	inputCh := make(chan []byte, jobCount)
	group.Go(func() error {
		defer close(inputCh) // closes itself
		for _, part := range parts {
			select {
			case inputCh <- part:
			case <-ctx.Done():
				return nil
			}
		}
		return nil
	})

	// create a channel-based iterator, to fetch the shards,
	// randomly and thread-save
	shardCh := datastor.ShardIteratorChannel(ctx,
		ds.cluster.GetRandomShardIterator(nil), jobCount)

	// write all the different parts to their own separate shard,
	// and return the identifiers of the used shards over the resultCh,
	// which will be used to collect all the successfull shards' identifiers for the final output
	resultCh := make(chan string, jobCount)
	// create all the actual workers
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			var (
				part  []byte
				open  bool
				err   error
				shard datastor.Shard
			)
			for {
				// wait for a part to write
				select {
				case part, open = <-inputCh:
					if !open {
						// channel is closed -> return
						return nil
					}
				case <-ctx.Done():
					return nil
				}

				// loop here, until we either have an error,
				// or until we have written to a shard
			writeLoop:
				for {
					// fetch a random shard,
					// it's an error if this is not possible,
					// as a shard is expected to be still available at this stage
					select {
					case shard, open = <-shardCh:
						if !open {
							// not enough shards are available,
							// we know this because the iterator ch has already been closed
							return ErrInsufficientShards
						}
					case <-ctx.Done():
						return errors.New("context was unexpectedly cancelled, " +
							"while fetching shard for a distribute-write request")
					}

					// do the actual storage
					err = shard.SetObject(datastor.Object{
						Key:           object.Key,
						Data:          part,
						ReferenceList: object.ReferenceList,
					})
					if err == nil {
						select {
						case resultCh <- shard.Identifier():
							break writeLoop
						case <-ctx.Done():
							return errors.New("context was unexpectedly cancelled, " +
								"while returning the identifier of a shard for a distribute-write request")
						}
					}

					// casually log the shard-write error,
					// and continue trying with another shard...
					log.Errorf("failed to write %q to random shard %q: %v",
						object.Key, shard.Identifier(), err)
				}
			}
		})
	}

	// close the result channel,
	// when all grouped goroutines are finished, so it can be used as an iterator
	go func() {
		err := group.Wait()
		if err != nil {
			log.Errorf("duplicate-writing %q has failed due to an error: %v",
				object.Key, err)
		}
		close(resultCh)
	}()

	// collect the identifiers of all shards, we could write our object to
	shards := make([]string, 0, partsCount)
	// fetch all results
	for id := range resultCh {
		shards = append(shards, id)
	}

	cfg := StorageConfig{Key: object.Key, Shards: shards}
	// check if we have sufficient distributions
	if len(shards) < partsCount {
		return cfg, ErrInsufficientShards
	}
	return cfg, nil
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

	// DataShardCount returns the (minimum) amount of data shards,
	// required to fetch data from, in order to be able to decode the encoded data.
	//
	// The implementation can return `-1` in case this feature is not supported,
	// in which case it will fetch data from as much shards as there are given.
	DataShardCount() int
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
		return nil, errors.New("no data given to encode")
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
		return nil, errors.New("no parts given to decode")
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

// DataShardCount implements DistributedEncoderDecoder.DataShardCount
func (rs *ReedSolomonEncoderDecoder) DataShardCount() int {
	return rs.k
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
