package storage

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"

	log "github.com/Sirupsen/logrus"
	"github.com/templexxx/reedsolomon"
	"github.com/zero-os/0-stor/client/datastor"
)

// NewDistributedObjectStorage creates a new DistributedObjectStorage,
// using the given Cluster and default ReedSolomonEncoderDecoder as internal DistributedEncoderDecoder.
// See `DistributedObjectStorage` `DistributedEncoderDecoder` for more information.
func NewDistributedObjectStorage(cluster datastor.Cluster, k, m, jobCount int) (*DistributedObjectStorage, error) {
	dec, err := NewReedSolomonEncoderDecoder(k, m)
	if err != nil {
		return nil, err
	}
	return NewDistributedObjectStorageWithEncoderDecoder(cluster, dec, jobCount), nil
}

// NewDistributedObjectStorageWithEncoderDecoder creates a new DistributedObjectStorage,
// using the given Cluster and DistributedEncoderDecoder.
// See `DistributedObjectStorage` `DistributedEncoderDecoder` for more information.
func NewDistributedObjectStorageWithEncoderDecoder(cluster datastor.Cluster, dec DistributedEncoderDecoder, jobCount int) *DistributedObjectStorage {
	if cluster == nil {
		panic("no Cluster given")
	}
	if dec == nil {
		panic("no DistributedEncoderDecoder given")
	}

	if jobCount < 1 {
		jobCount = DefaultJobCount
	}

	return &DistributedObjectStorage{
		cluster:  cluster,
		dec:      dec,
		jobCount: jobCount,
	}
}

// DistributedObjectStorage defines a storage implementation,
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
// When creating a DistributedObjectStorage you can also pass in your
// own DistributedEncoderDecoder should you not be satisfied with the default implementation.
type DistributedObjectStorage struct {
	cluster  datastor.Cluster
	dec      DistributedEncoderDecoder
	jobCount int
}

// Write implements storage.ObjectStorage.Write
func (ds *DistributedObjectStorage) Write(object datastor.Object) (ObjectConfig, error) {
	parts, err := ds.dec.Encode(object.Data)
	if err != nil {
		return ObjectConfig{}, err
	}

	group, ctx := errgroup.WithContext(context.Background())

	jobCount := ds.jobCount
	partsCount := len(parts)
	if jobCount > partsCount {
		jobCount = partsCount
	}

	// sends each part to an available worker goroutine,
	// which tries to store it in a random shard.
	// however make sure that we store the shard list,
	// in the same order as how we received the different parts,
	// otherwise we might not be able to decode it once again.
	type indexedPart struct {
		Index int
		Data  []byte
	}
	inputCh := make(chan indexedPart, jobCount)
	group.Go(func() error {
		defer close(inputCh) // closes itself
		for index, part := range parts {
			select {
			case inputCh <- indexedPart{index, part}:
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
	type indexedShard struct {
		Index      int
		Identifier string
	}
	resultCh := make(chan indexedShard, jobCount)
	// create all the actual workers
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			var (
				part  indexedPart
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
							return ErrShardsUnavailable
						}
					case <-ctx.Done():
						return errors.New("context was unexpectedly cancelled, " +
							"while fetching shard for a distribute-write request")
					}

					// do the actual storage
					err = shard.SetObject(datastor.Object{
						Key:           object.Key,
						Data:          part.Data,
						ReferenceList: object.ReferenceList,
					})
					if err == nil {
						select {
						case resultCh <- indexedShard{part.Index, shard.Identifier()}:
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

	// collect the identifiers of all shards we could write our object to,
	// and store+send them in the same order as how we received the parts
	var (
		resultCount int
		shards      = make([]string, partsCount)
	)
	// fetch all results
	for result := range resultCh {
		shards[result.Index] = result.Identifier
		resultCount++
	}

	cfg := ObjectConfig{Key: object.Key, Shards: shards, DataSize: len(object.Data)}
	// check if we have sufficient distributions
	if resultCount < partsCount {
		return cfg, ErrShardsUnavailable
	}
	return cfg, nil
}

// Read implements storage.ObjectStorage.Read
func (ds *DistributedObjectStorage) Read(cfg ObjectConfig) (datastor.Object, error) {
	// validate the input shard count
	shardCount := len(cfg.Shards)
	if shardCount < 2 { // min shard count (k+m, where both values are 1)
		return datastor.Object{}, ErrUnexpectedShardsCount
	}

	// define the jobCount
	jobCount := ds.jobCount
	if jobCount > shardCount {
		jobCount = shardCount
	}

	// create our sync-purpose variables
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)

	// create a channel-based iterator, to fetch the shards,
	// in sequence as given, and thread-save,
	// also attach the index to each shard, such that
	// we can deliver the parts in the correct order
	type indexedShard struct {
		Index int
		Shard datastor.Shard
	}
	shardCh := make(chan indexedShard, jobCount)
	go func() {
		defer close(shardCh)

		var (
			index int
			it    = datastor.NewLazyShardIterator(ds.cluster, cfg.Shards)
		)
		for it.Next() {
			select {
			case shardCh <- indexedShard{index, it.Shard()}:
				index++
			case <-ctx.Done():
				return
			}
		}
	}()

	type readResult struct {
		Index         int
		Data          []byte
		ReferenceList []string
	}

	// read all the needed parts,
	// from the available datashards
	resultCh := make(chan readResult, jobCount)
	// create all the actual workers
	for i := 0; i < jobCount; i++ {
		group.Go(func() error {
			var (
				open   bool
				object *datastor.Object
				err    error
				shard  indexedShard
			)
			for {
				// fetch a random shard,
				// it's an error if this is not possible,
				// as a shard is expected to be still available at this stage
				select {
				case shard, open = <-shardCh:
					if !open {
						return nil
					}
				case <-ctx.Done():
					return nil
				}

				// fetch the data part
				object, err = shard.Shard.GetObject(cfg.Key)
				if err != nil {
					// casually log the shard-read error,
					// and continue trying with another shard...
					log.Errorf("failed to read %q from given shard %q: %v",
						cfg.Key, shard.Shard.Identifier(), err)
					continue // try another shard
				}
				result := readResult{
					Index:         shard.Index,
					Data:          object.Data,
					ReferenceList: object.ReferenceList,
				}
				select {
				case resultCh <- result:
				case <-ctx.Done():
					return errors.New("context was unexpectedly cancelled, " +
						"while returning the data part, freshly fetched from a shard for a distribute-read request")
				}
			}
		})
	}

	// close the result channel,
	// when all grouped goroutines are finished, so it can be used as an iterator
	go func() {
		err := group.Wait()
		if err != nil {
			log.Errorf("distribute-read %q has failed due to an error: %v",
				cfg.Key, err)
		}
		close(resultCh)
	}()

	// collect all the different distributed parts
	var (
		referenceList []string
		resultCount   int

		parts = make([][]byte, shardCount)
	)

	for result := range resultCh {
		// put the part in the correct slot
		parts[result.Index] = result.Data
		resultCount++

		// if the referenceList wasn't set yet, do so now
		if referenceList == nil {
			referenceList = result.ReferenceList
			continue
		}
		// TODO: Validate ReferenceList somehow?! Store ReferenceList better?!
	}

	// ensure that we have received all the different parts
	if resultCount < shardCount {
		return datastor.Object{}, ErrShardsUnavailable
	}

	// decode the distributed data
	data, err := ds.dec.Decode(parts, cfg.DataSize)
	if err != nil {
		return datastor.Object{}, err
	}
	if len(data) != cfg.DataSize {
		return datastor.Object{}, ErrInvalidDataSize
	}

	// return decoded object
	return datastor.Object{
		Key:           cfg.Key,
		Data:          data,
		ReferenceList: referenceList,
	}, nil
}

// Check implements storage.ObjectStorage.Check
func (ds *DistributedObjectStorage) Check(cfg ObjectConfig, fast bool) ObjectCheckStatus {
	return ObjectCheckStatusOptimal // TODO
}

// Repair implements storage.ObjectStorage.Repair
func (ds *DistributedObjectStorage) Repair(cfg ObjectConfig) (ObjectConfig, error) {
	panic("TODO")
}

// Close implements storage.ObjectStorage.Close
func (ds *DistributedObjectStorage) Close() error {
	return ds.cluster.Close()
}

// DistributedEncoderDecoder is the type used internally to
// read and write the data of objects, read and written using the DistributedObjectStorage.
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
// for the DistributedObjectStorage storage type.
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
	_ ObjectStorage = (*DistributedObjectStorage)(nil)

	_ DistributedEncoderDecoder = (*ReedSolomonEncoderDecoder)(nil)
)
