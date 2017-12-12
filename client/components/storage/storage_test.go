package storage

import (
	"crypto/rand"
	mathRand "math/rand"
	"testing"

	"github.com/zero-os/0-stor/client/datastor"

	"github.com/stretchr/testify/require"
)

func testStorageReadWrite(t *testing.T, storage Storage) {
	require := require.New(t)
	require.NotNil(storage)

	t.Run("fixed test cases", func(t *testing.T) {

	})

	t.Run("random test cases", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			key := make([]byte, mathRand.Int31n(64)+1)
			rand.Read(key)
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

			cfg, err := storage.Write(inputObject)
			require.NoError(err)
			require.Equal(inputObject.Key, cfg.Key)
			require.Equal(len(data), cfg.DataSize)

			outputObject, err := storage.Read(cfg)
			require.NoError(err)
			require.Equal(inputObject, outputObject)
		}
	})
}
