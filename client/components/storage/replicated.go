package storage

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"

	log "github.com/Sirupsen/logrus"
	"github.com/zero-os/0-stor/client/datastor"
)

// NewReplicatedObjectStorage creates a new ReplicatedObjectStorage.
// See `ReplicatedObjectStorage` for more information.
//
// jobCount is optional and can be `<= 0` in order to use DefaultJobCount.
func NewReplicatedObjectStorage(cluster datastor.Cluster, replicationNr, jobCount int) *ReplicatedObjectStorage {
	if cluster == nil {
		panic("no cluster given")
	}
	if replicationNr < 1 {
		panic("replicationNr has to be at least 1")
	}

	if jobCount < 1 {
		jobCount = DefaultJobCount
	}
	writeJobCount := jobCount
	if writeJobCount < replicationNr {
		writeJobCount = replicationNr
	}

	return &ReplicatedObjectStorage{
		cluster:       cluster,
		replicationNr: replicationNr,
		jobCount:      jobCount,
		writeJobCount: writeJobCount,
	}
}

// ReplicatedObjectStorage defines a storage implementation,
// which writes an object to multiple shards at once,
// the amount of shards which is defined by the used replicationNr.
//
// For reading it will try to a multitude of the possible shards at once,
// and return the object that it received first. As it is expected that all
// shards return the same object for this key, when making use of this storage,
// there is no need to read from all shards and wait for all of those results as well.
//
// Repairing is done by first assembling a list of corrupt, OK and dead shards.
// Once that's done, the corrupt shards will be simply tried to be written to again,
// while the dead shards will be attempted to be replaced, if possible.
type ReplicatedObjectStorage struct {
	cluster                 datastor.Cluster
	replicationNr           int
	jobCount, writeJobCount int
}

// Write implements storage.ObjectStorage.Write
func (rs *ReplicatedObjectStorage) Write(object datastor.Object) (ObjectConfig, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// request the worker goroutines,
	// to get exactly replicationNr amount of replications.
	requestCh := make(chan struct{}, rs.writeJobCount)
	go func() {
		defer close(requestCh) // closes itself
		for i := rs.replicationNr; i > 0; i-- {
			select {
			case requestCh <- struct{}{}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// create a channel-based iterator, to fetch the shards,
	// randomly and thread-save
	shardCh := datastor.ShardIteratorChannel(ctx,
		rs.cluster.GetRandomShardIterator(nil), rs.writeJobCount)

	group, ctx := errgroup.WithContext(ctx)

	// write to replicationNr amount of shards,
	// and return their identifiers over the resultCh,
	// collection all the successfull shards' identifiers for the final output
	resultCh := make(chan string, rs.writeJobCount)
	// create all the actual workers
	for i := 0; i < rs.writeJobCount; i++ {
		group.Go(func() error {
			var (
				open  bool
				err   error
				shard datastor.Shard
			)
			for {
				// wait for a request
				select {
				case _, open = <-requestCh:
					if !open {
						// fake request: channel is closed -> return
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
							"while fetching shard for a replicate-write request")
					}

					// do the actual storage
					err = shard.SetObject(object)
					if err == nil {
						select {
						case resultCh <- shard.Identifier():
							break writeLoop
						case <-ctx.Done():
							return errors.New("context was unexpectedly cancelled, " +
								"while returning the identifier of a shard for a replicate-write request")
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
			log.Errorf("replicate-writing %q has failed due to an error: %v",
				object.Key, err)
		}
		close(resultCh)
	}()

	// collect the identifiers of all shards, we could write our object to
	shards := make([]string, 0, rs.replicationNr)
	// fetch all results
	for id := range resultCh {
		shards = append(shards, id)
	}

	cfg := ObjectConfig{Key: object.Key, Shards: shards, DataSize: len(object.Data)}
	// check if we have sufficient replications
	if len(shards) < rs.replicationNr {
		return cfg, ErrShardsUnavailable
	}
	return cfg, nil
}

// Read implements storage.ObjectStorage.Read
func (rs *ReplicatedObjectStorage) Read(cfg ObjectConfig) (datastor.Object, error) {
	// ensure that at least 1 shard is given
	if len(cfg.Shards) == 0 {
		return datastor.Object{}, nil
	}

	var (
		err    error
		object *datastor.Object
		shard  datastor.Shard

		it = datastor.NewLazyShardIterator(rs.cluster, cfg.Shards)
	)
	// simply try to read sequentially until one could be read,
	// as we should in most scenarios only ever have to read from 1 (and 2 or 3 in bad situations),
	// it would be bad for performance to try to read from multiple goroutines and shards for all calls.
	for it.Next() {
		shard = it.Shard()
		object, err = shard.GetObject(cfg.Key)
		if err == nil {
			if len(object.Data) == cfg.DataSize {
				return *object, nil
			}
			log.Errorf("failed to read %q from replicated shard %q: invalid data size",
				cfg.Key, shard.Identifier())
		} else {
			log.Errorf("failed to read %q from replicated shard %q: %v",
				cfg.Key, shard.Identifier(), err)
		}
	}

	// sadly, no shard was available
	log.Errorf("%q couldn't be replicate-read from any of the configured shards", cfg.Key)
	return datastor.Object{}, ErrShardsUnavailable
}

// Check implements storage.ObjectStorage.Check
func (rs *ReplicatedObjectStorage) Check(cfg ObjectConfig, fast bool) (ObjectCheckStatus, error) {
	shardCount := len(cfg.Shards)
	if shardCount == 0 {
		return ObjectCheckStatusInvalid, ErrUnexpectedShardsCount
	}

	// define the jobCount
	jobCount := rs.jobCount
	if jobCount > shardCount {
		jobCount = shardCount
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create a channel-based iterator, to fetch the shards,
	// randomly and thread-save
	shardCh := datastor.ShardIteratorChannel(ctx,
		datastor.NewLazyShardIterator(rs.cluster, cfg.Shards), jobCount)

	// each worker will help us get through all shards,
	// until we found the desired amount of valid shards,
	// the maximum which is helped guarantee by the requestCh iterator,
	// while the minimum is defined by that same channel or by exhausting the shardCh.
	resultCh := make(chan struct{}, jobCount)

	// create our goroutine,
	// to close our resultCh in case we have exhausted our worker goroutines
	var wg sync.WaitGroup
	wg.Add(jobCount)
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// create all the actual workers
	for i := 0; i < jobCount; i++ {
		go func() {
			var (
				open   bool
				err    error
				status datastor.ObjectStatus
				shard  datastor.Shard
			)

			for {
				// fetch a random shard,
				// it's an error if this is not possible,
				// as a shard is expected to be still available at this stage
				select {
				case shard, open = <-shardCh:
					if !open {
						return
					}
				case <-ctx.Done():
					return
				}

				// validate if the object's status for this shard is OK
				status, err = shard.GetObjectStatus(cfg.Key)
				if err != nil {
					log.Errorf("error while validating %q stored on shard %q: %v",
						cfg.Key, shard.Identifier(), err)
					continue
				}
				if status != datastor.ObjectStatusOK {
					log.Debugf("object %q stored on shard %q is not valid: %s",
						cfg.Key, shard.Identifier(), status)
					continue
				}

				// shard is valid for this object,
				// notify the result collector about it
				select {
				case resultCh <- struct{}{}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// if we want a fast result,
	// we simply want to know that at least one is available
	if fast {
		select {
		case <-resultCh:
			return ObjectCheckStatusValid, nil
		case <-ctx.Done():
			return ObjectCheckStatusInvalid, nil
		}
	}

	// otherwise we'll go through all of them,
	// until we have a max of nrReplication results
	var validShardCount int
	for range resultCh {
		validShardCount++
		if validShardCount == rs.replicationNr {
			return ObjectCheckStatusOptimal, nil
		}
	}

	if validShardCount > 0 {
		return ObjectCheckStatusValid, nil
	}
	return ObjectCheckStatusInvalid, nil
}

// Repair implements storage.ObjectStorage.Repair
func (rs *ReplicatedObjectStorage) Repair(cfg ObjectConfig) (ObjectConfig, error) {
	panic("TODO")
}

// Close implements storage.ObjectStorage.Close
func (rs *ReplicatedObjectStorage) Close() error {
	return rs.cluster.Close()
}

var (
	_ ObjectStorage = (*ReplicatedObjectStorage)(nil)
)
