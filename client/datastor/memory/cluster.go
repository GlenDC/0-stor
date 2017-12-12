package memory

import (
	"errors"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/zero-os/0-stor/client/datastor"
)

// NewCluster creates a cluster of listed and unlisted in-memory shards.
// It is legal for an in-memory cluster to have no listed shards.
// The given ids (identifiers) will be used to create the list of known shards.
// See Cluster for more information.
func NewCluster(ids []string) *Cluster {
	var (
		shardSlice   []*Shard
		shardMapping = make(map[string]*Shard)
	)
	for _, id := range ids {
		shard := &Shard{
			Client: NewClient(),
			id:     id,
		}
		shardSlice = append(shardSlice, shard)
		shardMapping[id] = shard
	}
	return &Cluster{
		listedShards:      shardMapping,
		listedShardsSlice: shardSlice,
		listedCount:       int64(len(shardSlice)),
		unlistedShards:    make(map[string]*Shard),
	}
}

// Cluster implements the datastor.Cluster interface,
// as to collect a group of listed and unlisted inn-memory datastor Clients.
//
// To be used for development and testing purposes only.
type Cluster struct {
	listedShards      map[string]*Shard
	listedShardsSlice []*Shard
	listedCount       int64

	unlistedShards map[string]*Shard

	mux sync.Mutex
}

// GetShard implements datastor.Cluster.GetShard
func (cluster *Cluster) GetShard(id string) (datastor.Shard, error) {
	shard, ok := cluster.listedShards[id]
	if ok {
		return shard, nil
	}

	cluster.mux.Lock()
	defer cluster.mux.Unlock()

	shard, ok = cluster.unlistedShards[id]
	if ok {
		return shard, nil
	}

	shard = &Shard{
		Client: NewClient(),
		id:     id,
	}
	cluster.unlistedShards[id] = shard
	return shard, nil
}

// GetRandomShard implements datastor.Cluster.GetRandomShard
func (cluster *Cluster) GetRandomShard() (datastor.Shard, error) {
	if cluster.listedCount == 0 {
		return nil, datastor.ErrNoShardsAvailable
	}
	index := datastor.RandShardIndex(cluster.listedCount)
	return cluster.listedShardsSlice[index], nil
}

// GetRandomShardIterator implements datastor.Cluster.GetRandomShardIterator
func (cluster *Cluster) GetRandomShardIterator(exceptShards []string) datastor.ShardIterator {
	slice := cluster.filteredSlice(exceptShards)
	return datastor.NewRandomShardIterator(slice)
}

// Close implements datastor.Cluster.Close
func (cluster *Cluster) Close() error {
	cluster.mux.Lock()
	defer cluster.mux.Unlock()

	var (
		err      error
		errCount int
	)

	for _, shard := range cluster.unlistedShards {
		err = shard.Close()
		if err != nil {
			errCount++
			log.Errorf("error while closing unlisted shard %q: %v",
				shard.Identifier(), err)
		}
	}

	for _, shard := range cluster.listedShards {
		err = shard.Close()
		if err != nil {
			errCount++
			log.Errorf("error while closing listed shard %q: %v",
				shard.Identifier(), err)
		}
	}

	if errCount > 0 {
		return errors.New("error while closing at least one shard")
	}
	return nil
}

func (cluster *Cluster) filteredSlice(exceptShards []string) []datastor.Shard {
	if len(exceptShards) == 0 {
		slice := make([]datastor.Shard, cluster.listedCount)
		for i := range slice {
			slice[i] = cluster.listedShardsSlice[i]
		}
		return slice
	}

	fm := make(map[string]struct{}, len(exceptShards))
	for _, id := range exceptShards {
		fm[id] = struct{}{}
	}

	var (
		ok       bool
		filtered = make([]datastor.Shard, 0, cluster.listedCount)
	)
	for _, shard := range cluster.listedShardsSlice {
		if _, ok = fm[shard.Identifier()]; !ok {
			filtered = append(filtered, shard)
		}
	}
	return filtered
}

// Shard implements the datastor.Shard interface,
// for an in-memory datastor Client,
// such that it can be used as part of an in-memory Cluster.
//
// To be used for development and testing purposes only.
type Shard struct {
	*Client

	id string
}

// Identifier implements datastor.Shard.Identifier
func (shard *Shard) Identifier() string {
	return shard.id
}
