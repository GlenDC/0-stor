package storage

import (
	"crypto/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDistributedStoragePanics(t *testing.T) {
	require.Panics(t, func() {
		NewDistributedObjectStorage(nil, 1, 1, -1)
	}, "no cluster given given")
	require.Panics(t, func() {
		NewDistributedObjectStorage(dummyCluster{}, 0, 1, -1)
	}, "no valid k (data shard count) given")
	require.Panics(t, func() {
		NewDistributedObjectStorage(dummyCluster{}, 1, 0, -1)
	}, "no valid m (parity shard count) given")
}

func TestDistributedStorageReadCheckWrite(t *testing.T) {
	t.Run("k=1,m=1,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(4)
		require.NoError(t, err)
		defer cleanup()

		storage, err := NewDistributedObjectStorage(cluster, 1, 1, 0)
		require.NoError(t, err)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("k=2,m=1,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(6)
		require.NoError(t, err)
		defer cleanup()

		storage, err := NewDistributedObjectStorage(cluster, 2, 1, 0)
		require.NoError(t, err)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("k=1,m=2,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(6)
		require.NoError(t, err)
		defer cleanup()

		storage, err := NewDistributedObjectStorage(cluster, 1, 2, 0)
		require.NoError(t, err)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("k=2,m=2,jobCount=1", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(8)
		require.NoError(t, err)
		defer cleanup()

		storage, err := NewDistributedObjectStorage(cluster, 2, 2, 1)
		require.NoError(t, err)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("k=8,m=8,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(16)
		require.NoError(t, err)
		defer cleanup()

		storage, err := NewDistributedObjectStorage(cluster, 8, 8, 0)
		require.NoError(t, err)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("k=4,m=8,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(16)
		require.NoError(t, err)
		defer cleanup()

		storage, err := NewDistributedObjectStorage(cluster, 4, 8, 0)
		require.NoError(t, err)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("k=8,m=4,jobCount=D", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(16)
		require.NoError(t, err)
		defer cleanup()

		storage, err := NewDistributedObjectStorage(cluster, 8, 4, 0)
		require.NoError(t, err)

		testStorageReadCheckWrite(t, storage)
	})

	t.Run("k=8,m=8,jobCount=1", func(t *testing.T) {
		cluster, cleanup, err := newGRPCServerCluster(16)
		require.NoError(t, err)
		defer cleanup()

		storage, err := NewDistributedObjectStorage(cluster, 8, 8, 1)
		require.NoError(t, err)

		testStorageReadCheckWrite(t, storage)
	})
}

func TestDistributedStorageRepair(t *testing.T) {
	// TODO
}

func TestReedSolomonEncoderDecoderPanics(t *testing.T) {
	require := require.New(t)

	require.Panics(func() {
		NewReedSolomonEncoderDecoder(0, 1)
	}, "k is too small")
	require.Panics(func() {
		NewReedSolomonEncoderDecoder(1, 0)
	}, "m is too small")
}

func TestReedSolomonEncoderDecoderErrors(t *testing.T) {
	require := require.New(t)

	require.Error(func() error {
		ed, err := NewReedSolomonEncoderDecoder(1, 1)
		require.NoError(err)
		_, err = ed.Encode(nil)
		return err
	}(), "cannot encode nil-data")
	require.Error(func() error {
		ed, err := NewReedSolomonEncoderDecoder(1, 1)
		require.NoError(err)
		_, err = ed.Decode(nil, 1)
		return err
	}(), "cannot decode 0 parts")
}

func TestReedSolomonEncoderDecoder(t *testing.T) {
	t.Run("k=1, m=1", func(t *testing.T) {
		testReedSolomonEncoderDecoder(t, 1, 1)
	})
	t.Run("k=1, m=4", func(t *testing.T) {
		testReedSolomonEncoderDecoder(t, 1, 4)
	})
	t.Run("k=4, m=1", func(t *testing.T) {
		testReedSolomonEncoderDecoder(t, 4, 1)
	})
	t.Run("k=4, m=4", func(t *testing.T) {
		testReedSolomonEncoderDecoder(t, 4, 4)
	})
	t.Run("k=16, m=1", func(t *testing.T) {
		testReedSolomonEncoderDecoder(t, 16, 1)
	})
	t.Run("k=1, m=16", func(t *testing.T) {
		testReedSolomonEncoderDecoder(t, 1, 16)
	})
	t.Run("k=16, m=16", func(t *testing.T) {
		testReedSolomonEncoderDecoder(t, 16, 16)
	})
}

func TestReedSolomonEncoderDecoderAsyncUsage(t *testing.T) {
	t.Run("k=1, m=1, jc=2", func(t *testing.T) {
		testReedSolomonEncoderDecoderAsyncUsage(t, 1, 1, 2)
	})
	t.Run("k=1, m=1, jc=16", func(t *testing.T) {
		testReedSolomonEncoderDecoderAsyncUsage(t, 1, 1, 16)
	})
	t.Run("k=4, m=4, jc=2", func(t *testing.T) {
		testReedSolomonEncoderDecoderAsyncUsage(t, 4, 4, 2)
	})
	t.Run("k=4, m=4, jc=16", func(t *testing.T) {
		testReedSolomonEncoderDecoderAsyncUsage(t, 4, 4, 16)
	})
	t.Run("k=16, m=16, jc=2", func(t *testing.T) {
		testReedSolomonEncoderDecoderAsyncUsage(t, 16, 16, 2)
	})
	t.Run("k=16, m=16, jc=16", func(t *testing.T) {
		testReedSolomonEncoderDecoderAsyncUsage(t, 16, 16, 16)
	})
	t.Run("k=16, m=16, jc=16", func(t *testing.T) {
		testReedSolomonEncoderDecoderAsyncUsage(t, 16, 16, 128)
	})
}

func testReedSolomonEncoderDecoderAsyncUsage(t *testing.T, k, m, jobCount int) {
	assert := assert.New(t)

	ed, err := NewReedSolomonEncoderDecoder(k, m)
	require.NoError(t, err)
	require.NotNil(t, ed)

	var wg sync.WaitGroup
	wg.Add(jobCount)

	for i := 0; i < jobCount; i++ {
		go func() {
			defer wg.Done()

			input := make([]byte, 4096)
			_, err := rand.Read(input)
			assert.NoError(err)

			parts, err := ed.Encode(input)
			assert.NoError(err)
			assert.NotEmpty(parts)

			output, err := ed.Decode(parts, len(input))
			assert.NoError(err)
			assert.Equal(input, output)
		}()
	}

	wg.Wait()
}

func testReedSolomonEncoderDecoder(t *testing.T, k, m int) {
	require := require.New(t)

	ed, err := NewReedSolomonEncoderDecoder(k, m)
	require.NoError(err)
	require.NotNil(ed)

	testCases := []string{
		"a",
		"Hello, World!",
		func() string {
			b := make([]byte, 4096)
			_, err := rand.Read(b)
			require.NoError(err)
			return string(b)
		}(),
		"大家好",
	}

	for _, testCase := range testCases {
		parts, err := ed.Encode([]byte(testCase))
		require.NoError(err)
		require.NotEmpty(parts)

		data, err := ed.Decode(parts, len(testCase))
		require.NoError(err)
		require.Equal(testCase, string(data))
	}
}
