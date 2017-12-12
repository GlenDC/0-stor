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

func TestRandomStorageReadWrite(t *testing.T) {
	cluster := memory.NewCluster([]string{"a", "b", "c"})
	require.NotNil(t, cluster)

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
