package storage

import (
	"crypto/rand"
	"fmt"
	mathRand "math/rand"
	"testing"

	"github.com/zero-os/0-stor/client/datastor"

	"github.com/stretchr/testify/require"
)

func testStorageReadWrite(t *testing.T, storage Storage, cluster datastor.Cluster) {
	require.NotNil(t, storage)

	t.Run("fixed test cases", func(t *testing.T) {
		require := require.New(t)

		objects := []datastor.Object{
			datastor.Object{
				Key:  []byte("a"),
				Data: []byte("b"),
			},
			datastor.Object{
				Key:  []byte("foo"),
				Data: []byte("bar"),
			},
			datastor.Object{
				Key:  []byte("大家好"),
				Data: []byte("大家好"),
			},
			datastor.Object{
				Key:           []byte("this-is-my-key"),
				Data:          []byte("Hello, World!"),
				ReferenceList: []string{"user1", "user2"},
			},
			datastor.Object{
				Key:           []byte("this-is-my-key"),
				Data:          []byte("Hello, World!"),
				ReferenceList: []string{"user1", "user2"},
			},
		}
		for _, inputObject := range objects {
			// write object & validate
			cfg, err := storage.Write(inputObject)
			require.NoError(err)
			require.Equal(inputObject.Key, cfg.Key)
			require.Equal(len(inputObject.Data), cfg.DataSize)

			// validate that all shards contain valid data
			for _, shardID := range cfg.Shards {
				shard, err := cluster.GetShard(shardID)
				require.NoError(err)
				require.Equal(shardID, shard.Identifier())

				status, err := shard.GetObjectStatus(cfg.Key)
				require.NoError(err)
				require.Equalf(datastor.ObjectStatusOK.String(), status.String(),
					"key %q NotFound/Corrupt", cfg.Key)
			}

			// read object & validate
			outputObject, err := storage.Read(cfg)
			require.NoError(err)
			require.Equal(inputObject, outputObject)
		}
	})

	t.Run("random test cases", func(t *testing.T) {
		require := require.New(t)

		for i := 0; i < 256; i++ {
			key := []byte(fmt.Sprintf("key#%d", i+1))
			data := make([]byte, mathRand.Int31n(4096)+1)
			rand.Read(data)

			refList := make([]string, mathRand.Int31n(16)+1)
			for i := range refList {
				id := make([]byte, mathRand.Int31n(128)+1)
				rand.Read(id)
				refList[i] = string(id)
			}

			inputObject := datastor.Object{
				Key:           key,
				Data:          data,
				ReferenceList: refList,
			}

			// write object & validate
			cfg, err := storage.Write(inputObject)
			require.NoError(err)
			require.Equal(inputObject.Key, cfg.Key)
			require.Equal(len(data), cfg.DataSize)

			// validate that all shards contain valid data
			for _, shardID := range cfg.Shards {
				shard, err := cluster.GetShard(shardID)
				require.NoError(err)
				require.Equal(shardID, shard.Identifier())

				status, err := shard.GetObjectStatus(cfg.Key)
				require.NoError(err)
				require.Equalf(datastor.ObjectStatusOK.String(), status.String(),
					"key %q NotFound/Corrupt", cfg.Key)
			}

			// read object & validate
			outputObject, err := storage.Read(cfg)
			require.NoError(err)
			require.Equal(inputObject, outputObject)
		}
	})
}
