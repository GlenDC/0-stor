package storage

import (
	"github.com/lunny/log"
	"github.com/zero-os/0-stor/client/datastor"
)

// NewRandomStorage creates a new RandomStorage.
// See `RandomStorage` for more information.
func NewRandomStorage(cluster datastor.Cluster) *RandomStorage {
	if cluster == nil {
		panic("no cluster given")
	}
	return &RandomStorage{cluster: cluster}
}

// RandomStorage is the most simplest Storage implementation.
// For writing it only writes to one shard, randomly chosen.
// For reading it expects just, and only, one shard, to read from.
// Repairing is not supported for this storage for obvious reasons.
type RandomStorage struct {
	cluster datastor.Cluster
}

// Write implements storage.Storage.Write
func (rs *RandomStorage) Write(object datastor.Object) (StorageConfig, error) {
	var (
		err   error
		shard datastor.Shard
	)

	// go through all shards, in pseudo-random fashion,
	// until the object could be written to one of them.
	it := rs.cluster.GetRandomShardIterator(nil)
	for it.Next() {
		shard = it.Shard()
		err = shard.SetObject(object)
		if err == nil {
			return StorageConfig{
				Key:    object.Key,
				Shards: []string{shard.Identifier()},
			}, nil
		}
		log.Errorf("failed to write %q to random shard %q: %v",
			object.Key, shard.Identifier(), err)
	}
	return StorageConfig{}, ErrInsufficientShards
}

// Read implements storage.Storage.Read
func (rs *RandomStorage) Read(cfg StorageConfig) (datastor.Object, error) {
	if len(cfg.Shards) != 1 {
		return datastor.Object{}, ErrUnexpectedShardsCount
	}

	shard, err := rs.cluster.GetShard(cfg.Shards[0])
	if err != nil {
		return datastor.Object{}, err
	}

	object, err := shard.GetObject(cfg.Key)
	if err != nil {
		return datastor.Object{}, err
	}
	return *object, nil
}

// Repair implements storage.Storage.Repair
func (rs *RandomStorage) Repair(cfg StorageConfig) (StorageConfig, error) {
	return StorageConfig{}, ErrNotSupported
}

// Close implements storage.Storage.Close
func (rs *RandomStorage) Close() error {
	return rs.cluster.Close()
}

var (
	_ Storage = (*RandomStorage)(nil)
)
