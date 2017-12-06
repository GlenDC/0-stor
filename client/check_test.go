package client

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	servers, serverClean := testGRPCServer(t, 4)
	defer serverClean()

	shards := make([]string, len(servers))
	for i, server := range servers {
		shards[i] = server.Address()
	}

	policy := Policy{
		Organization:           "testorg",
		Namespace:              "namespace1",
		DataShards:             shards,
		MetaShards:             []string{"test"},
		IYOAppID:               "",
		IYOSecret:              "",
		BlockSize:              1024,
		Compress:               true,
		Encrypt:                true,
		EncryptKey:             "cF0BFpIsljOS8UmaP8YRHRX0nBPVRVPw",
		ReplicationNr:          0,
		ReplicationMaxSize:     1, //force to use distribution over replication
		DistributionNr:         2,
		DistributionRedundancy: 1,
	}

	c, err := getTestClient(policy)
	require.NoError(t, err, "fail to create client")
	defer c.Close()

	data := make([]byte, 602*10)

	_, err = rand.Read(data)
	require.NoError(t, err, "fail to read random data")
	key := []byte("testkey")

	meta, err := c.Write(key, data, []string{})
	require.NoError(t, err, "fail write data")

	// Check status is ok after a write
	status, err := c.Check(meta.Key)
	require.NoError(t, err, "fail to check object")
	assert.Equal(t, CheckStatusOk, status)

	// corrupt file by removing blocks
	store, err := c.getStor(meta.Chunks[0].Shards[0])
	require.NoError(t, err)

	for i := 0; i < len(meta.Chunks); i += 4 {
		if i%4 == 0 {
			err = store.ObjectDelete(meta.Chunks[i].Key)
			require.NoError(t, err)
		}
	}

	// Check status is corrupted
	status, err = c.Check(meta.Key)
	require.NoError(t, err, "fail to check object")
	assert.Equal(t, CheckStatusMissing, status)
}
