package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewReplicatedStoragePanics(t *testing.T) {
	require.Panics(t, func() {
		NewReplicatedObjectStorage(nil, 1, -1)
	}, "no cluster given given")
	require.Panics(t, func() {
		NewReplicatedObjectStorage(dummyCluster{}, 0, -1)
	}, "no valid replicationNr given")
}

func TestReplicationStorageReadCheckWrite(t *testing.T) {
	t.Run("replicationNr=1,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(2)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedObjectStorage(cluster, 1, 0)
		require.NotNil(t, storage)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("replicationNr=2,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(4)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedObjectStorage(cluster, 2, 0)
		require.NotNil(t, storage)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("replicationNr=2,jobCount=1", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(4)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedObjectStorage(cluster, 2, 1)
		require.NotNil(t, storage)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("replicationNr=16,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(32)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedObjectStorage(cluster, 16, 0)
		require.NotNil(t, storage)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("replicationNr=16,jobCount=1", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(32)
		require.NoError(t, err)
		defer cleanup()

		storage := NewReplicatedObjectStorage(cluster, 16, 1)
		require.NotNil(t, storage)

		testStorageReadCheckWrite(t, storage)
	})
}

func TestReplicatedStorageCheckRepair(t *testing.T) {
	// TODO
}
