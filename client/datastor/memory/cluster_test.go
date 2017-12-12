package memory

import (
	"context"
	"crypto/rand"
	"fmt"
	mathRand "math/rand"
	"testing"

	"golang.org/x/sync/errgroup"

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

	cluster := NewCluster([]string{"a", "b", "c"})
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
	require.True(id == "a" || id == "b" || id == "c")

	it := cluster.GetRandomShardIterator(nil)
	require.NotNil(it)

	require.Panics(func() {
		it.Shard()
	}, "invalid iterator, need to call Next First")

	keys := map[string]struct{}{
		"a": struct{}{},
		"b": struct{}{},
		"c": struct{}{},
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

	it = cluster.GetRandomShardIterator([]string{"b"})
	require.NotNil(it)

	keys = map[string]struct{}{
		"a": struct{}{},
		"c": struct{}{},
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

func TestGetRandomShardAsync(t *testing.T) {
	require := require.New(t)

	const jobs = 4096

	var shards []string
	for i := 0; i < jobs; i++ {
		shards = append(shards, fmt.Sprintf("shard#%d", i+1))
	}
	cluster := NewCluster(shards)
	require.NotNil(cluster)
	defer func() {
		err := cluster.Close()
		require.NoError(err)
	}()

	group, ctx := errgroup.WithContext(context.Background())

	type writeResult struct {
		shardID string
		object  datastor.Object
	}
	ch := make(chan writeResult, jobs)

	for i := 0; i < jobs; i++ {
		i := i
		group.Go(func() error {
			key := []byte(fmt.Sprintf("key#%d", i+1))
			data := make([]byte, mathRand.Int31n(4096)+1)
			rand.Read(data)

			refList := make([]string, mathRand.Int31n(16)+1)
			for i := range refList {
				id := make([]byte, mathRand.Int31n(128)+1)
				rand.Read(id)
				refList[i] = string(id)
			}

			object := datastor.Object{
				Key:           key,
				Data:          data,
				ReferenceList: refList,
			}

			shard, err := cluster.GetRandomShard()
			if err != nil {
				return fmt.Errorf("get rand shard for key %q: %v", object.Key, err)
			}

			err = shard.SetObject(object)
			if err != nil {
				return fmt.Errorf("set error for key %q in shard %q: %v",
					object.Key, shard.Identifier(), err)
			}

			result := writeResult{
				shardID: shard.Identifier(),
				object:  object,
			}

			select {
			case ch <- result:
			case <-ctx.Done():
			}

			return nil
		})

		group.Go(func() error {
			var result writeResult
			select {
			case result = <-ch:
			case <-ctx.Done():
				return nil
			}

			shard, err := cluster.GetShard(result.shardID)
			if err != nil {
				return fmt.Errorf("get shard %q for key %q: %v",
					result.shardID, result.object.Key, err)
			}
			require.Equal(result.shardID, shard.Identifier())

			outputObject, err := shard.GetObject(result.object.Key)
			if err != nil {
				return fmt.Errorf("get error for key %q in shard %q: %v",
					result.object.Key, result.shardID, err)
			}

			require.NotNil(outputObject)
			object := result.object

			require.Equal(object.Key, outputObject.Key)
			require.Len(outputObject.Data, len(object.Data))
			require.Equal(outputObject.Data, object.Data)
			require.Len(outputObject.ReferenceList, len(object.ReferenceList))
			require.Equal(outputObject.ReferenceList, object.ReferenceList)

			return nil
		})
	}

	err := group.Wait()
	require.NoError(err)
}

func TestGetRandomShardIteratorAsync(t *testing.T) {
	require := require.New(t)

	const jobs = 4096

	var shards []string
	for i := 0; i < jobs; i++ {
		shards = append(shards, fmt.Sprintf("shard#%d", i+1))
	}
	cluster := NewCluster(shards)
	require.NotNil(cluster)
	defer func() {
		err := cluster.Close()
		require.NoError(err)
	}()

	group, ctx := errgroup.WithContext(context.Background())

	type writeResult struct {
		shardID string
		object  datastor.Object
	}
	ch := make(chan writeResult, jobs)

	shardCh := datastor.ShardIteratorChannel(ctx, cluster.GetRandomShardIterator(nil), jobs)
	require.NotNil(shardCh)

	for i := 0; i < jobs; i++ {
		i := i
		group.Go(func() error {
			key := []byte(fmt.Sprintf("key#%d", i+1))
			data := make([]byte, mathRand.Int31n(4096)+1)
			rand.Read(data)

			refList := make([]string, mathRand.Int31n(16)+1)
			for i := range refList {
				id := make([]byte, mathRand.Int31n(128)+1)
				rand.Read(id)
				refList[i] = string(id)
			}

			object := datastor.Object{
				Key:           key,
				Data:          data,
				ReferenceList: refList,
			}

			var shard datastor.Shard
			select {
			case shard = <-shardCh:
			case <-ctx.Done():
				return nil
			}

			err := shard.SetObject(object)
			if err != nil {
				return fmt.Errorf("set error for key %q in shard %q: %v",
					object.Key, shard.Identifier(), err)
			}

			result := writeResult{
				shardID: shard.Identifier(),
				object:  object,
			}

			select {
			case ch <- result:
			case <-ctx.Done():
			}

			return nil
		})

		group.Go(func() error {
			var result writeResult
			select {
			case result = <-ch:
			case <-ctx.Done():
				return nil
			}

			shard, err := cluster.GetShard(result.shardID)
			if err != nil {
				return fmt.Errorf("get shard %q for key %q: %v",
					result.shardID, result.object.Key, err)
			}
			require.Equal(result.shardID, shard.Identifier())

			outputObject, err := shard.GetObject(result.object.Key)
			if err != nil {
				return fmt.Errorf("get error for key %q in shard %q: %v",
					result.object.Key, result.shardID, err)
			}

			require.NotNil(outputObject)
			object := result.object

			require.Equal(object.Key, outputObject.Key)
			require.Len(outputObject.Data, len(object.Data))
			require.Equal(outputObject.Data, object.Data)
			require.Len(outputObject.ReferenceList, len(object.ReferenceList))
			require.Equal(outputObject.ReferenceList, object.ReferenceList)

			return nil
		})
	}

	err := group.Wait()
	require.NoError(err)
}
