package memory

import (
	"testing"

	"github.com/zero-os/0-stor/client/datastor"

	"github.com/stretchr/testify/require"
)

func TestShard(t *testing.T) {
	shard := new(Shard)
	require.Empty(t, shard.Identifier())
	shard.id = "foo"
	require.Equal(t, "foo", shard.Identifier())
}

func TestGetShard(t *testing.T) {
	require := require.New(t)

	cluster := NewCluster(nil)
	require.NotNil(cluster)
	defer func() {
		err := cluster.Close()
		require.NoError(err)
	}()

	shard, err := cluster.GetRandomShard()
	require.Equal(datastor.ErrNoShardsAvailable, err)
	require.Nil(shard)

	shard, err = cluster.GetShard("42")
	require.NoError(err)
	require.NotNil(shard)
	require.Equal("42", shard.Identifier())

	shard, err = cluster.GetShard("42")
	require.NoError(err)
	require.NotNil(shard)
	require.Equal("42", shard.Identifier())

	cluster = NewCluster([]string{"ab"})
	require.NotNil(cluster)

	shard, err = cluster.GetShard("ab")
	require.NoError(err)
	require.NotNil(shard)
	require.Equal("ab", shard.Identifier())
}

func TestGetRandomShards_NoListedShards(t *testing.T) {
	require := require.New(t)

	cluster := NewCluster(nil)
	require.NotNil(cluster)
	defer func() {
		err := cluster.Close()
		require.NoError(err)
	}()

	shard, err := cluster.GetRandomShard()
	require.Equal(datastor.ErrNoShardsAvailable, err)
	require.Nil(shard)

	it := cluster.GetRandomShardIterator(nil)
	require.NotNil(it)
	require.False(it.Next())
	require.Panics(func() {
		it.Shard()
	}, "invalid iterator")

	it = cluster.GetRandomShardIterator([]string{"foo"})
	require.NotNil(it)
	require.False(it.Next())
	require.Panics(func() {
		it.Shard()
	}, "invalid iterator")
}

func TestGetRandomShards_WithListedShards(t *testing.T) {
	require := require.New(t)

	cluster := NewCluster([]string{"a", "b"})
	require.NotNil(cluster)
	defer func() {
		err := cluster.Close()
		require.NoError(err)
	}()

	shard, err := cluster.GetRandomShard()
	require.NoError(err)
	require.NotNil(shard)
	id := shard.Identifier()
	require.NotEmpty(id)
	require.True(id == "a" || id == "b")

	it := cluster.GetRandomShardIterator(nil)
	require.NotNil(it)

	require.Panics(func() {
		it.Shard()
	}, "invalid iterator, need to call Next First")

	keys := map[string]struct{}{
		"a": struct{}{},
		"b": struct{}{},
	}
	for it.Next() {
		shard := it.Shard()
		require.NotNil(shard)

		id := shard.Identifier()
		require.NotEmpty(id)
		_, ok := keys[id]
		require.True(ok)
		delete(keys, id)
	}
	require.Empty(keys)

	it = cluster.GetRandomShardIterator([]string{"a"})
	require.NotNil(it)

	keys = map[string]struct{}{
		"b": struct{}{},
	}
	for it.Next() {
		shard := it.Shard()
		require.NotNil(shard)

		id := shard.Identifier()
		require.NotEmpty(id)
		_, ok := keys[id]
		require.True(ok)
		delete(keys, id)
	}
	require.Empty(keys)
}
