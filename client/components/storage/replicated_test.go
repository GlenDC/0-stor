package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/client/datastor/memory"
)

func TestNewReplicatedStoragePanics(t *testing.T) {
	require.Panics(t, func() {
		NewReplicatedStorage(nil, 1, -1)
	}, "no cluster given given")
	require.Panics(t, func() {
		NewReplicatedStorage(memory.NewCluster(nil), 0, -1)
	}, "no valid replicationNr given")
}

func TestReplicationStorageReadWrite_InMemory(t *testing.T) {
	t.Run("replicationNr=1,jobCount=D", func(t *testing.T) {
		cluster := memory.NewCluster([]string{"a", "b"})
		require.NotNil(t, cluster)

		storage := NewReplicatedStorage(cluster, 1, 0)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})

	t.Run("replicationNr=2,jobCount=D", func(t *testing.T) {
		cluster := memory.NewCluster([]string{"a", "b", "c", "d"})
		require.NotNil(t, cluster)

		storage := NewReplicatedStorage(cluster, 2, 0)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})

	t.Run("replicationNr=2,jobCount=1", func(t *testing.T) {
		cluster := memory.NewCluster([]string{"a", "b", "c", "d"})
		require.NotNil(t, cluster)

		storage := NewReplicatedStorage(cluster, 2, 1)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})

	t.Run("replicationNr=16,jobCount=D", func(t *testing.T) {
		cluster := memory.NewCluster([]string{
			"a", "b", "c", "d", "e", "f", "g", "h",
			"i", "j", "k", "l", "m", "n", "i", "o",
			"aa", "ab", "ac", "ad", "ae", "af", "ag", "ah",
			"ai", "aj", "ak", "al", "am", "an", "ai", "ao",
		})
		require.NotNil(t, cluster)

		storage := NewReplicatedStorage(cluster, 16, 0)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})

	t.Run("replicationNr=16,jobCount=1", func(t *testing.T) {
		cluster := memory.NewCluster([]string{
			"a", "b", "c", "d", "e", "f", "g", "h",
			"i", "j", "k", "l", "m", "n", "i", "o",
			"aa", "ab", "ac", "ad", "ae", "af", "ag", "ah",
			"ai", "aj", "ak", "al", "am", "an", "ai", "ao",
		})
		require.NotNil(t, cluster)

		storage := NewReplicatedStorage(cluster, 16, 1)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})
}

func TestReplicationStorageReadWrite_GRPC(t *testing.T) {
	t.Run("replicationNr=1,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(2)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedStorage(cluster, 1, 0)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})

	t.Run("replicationNr=2,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(4)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedStorage(cluster, 2, 0)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})

	t.Run("replicationNr=2,jobCount=1", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(4)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedStorage(cluster, 2, 1)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})

	t.Run("replicationNr=16,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(32)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedStorage(cluster, 16, 0)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})

	t.Run("replicationNr=16,jobCount=1", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(32)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedStorage(cluster, 16, 1)
		require.NotNil(t, storage)

		testStorageReadWrite(t, storage, cluster)
	})
}

func TestReplicatedStorageRepair(t *testing.T) {
	// TODO
}
