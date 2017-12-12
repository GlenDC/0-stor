package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/client/datastor/memory"
)

func TestNewRandomStoragePanics(t *testing.T) {
	require.Panics(t, func() {
		NewRandomStorage(nil)
	}, "no cluster given")
}

func TestRandomStorageReadWrite_InMemory(t *testing.T) {
	cluster := memory.NewCluster([]string{"a", "b", "c"})
	require.NotNil(t, cluster)

	storage := NewRandomStorage(cluster)
	require.NotNil(t, storage)

	testStorageReadWrite(t, storage)
}

func TestRandomStorageReadWrite_GRPC(t *testing.T) {
	cluster, cleanup, err := newGRPCServerCluster(3)
	require.NoError(t, err)
	defer cleanup()

	storage := NewRandomStorage(cluster)
	require.NotNil(t, storage)

	testStorageReadWrite(t, storage)
}

func TestRandomStorageRepair(t *testing.T) {
	require := require.New(t)

	cluster := memory.NewCluster(nil)
	require.NotNil(cluster)

	storage := NewRandomStorage(cluster)
	require.NotNil(storage)

	defer func() {
		err := storage.Close()
		require.NoError(err)
	}()

	cfg, err := storage.Repair(StorageConfig{})
	require.Equal(ErrNotSupported, err)
	require.Equal(StorageConfig{}, cfg)
}
