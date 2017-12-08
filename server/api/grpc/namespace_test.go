package grpc

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/server"
	pb "github.com/zero-os/0-stor/server/api/grpc/schema"
	"github.com/zero-os/0-stor/server/db"
	"github.com/zero-os/0-stor/server/db/badger"
	"github.com/zero-os/0-stor/server/encoding"
	"golang.org/x/net/context"
)

// tot 11:00 gewerkt

func TestGetNamespace(t *testing.T) {
	api, clean := getTestNamespaceAPI(t)
	defer clean()

	data, err := encoding.EncodeNamespace(server.Namespace{Label: []byte(label)})
	require.NoError(t, err)
	err = api.db.Set(db.NamespaceKey([]byte(label)), data)
	require.NoError(t, err)

	req := &pb.GetNamespaceRequest{}
	resp, err := api.GetNamespace(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, label, resp.GetLabel())
	assert.EqualValues(t, 0, resp.GetNrObjects())
}

func getTestNamespaceAPI(t *testing.T) (*NamespaceAPI, func()) {
	require := require.New(t)

	tmpDir, err := ioutil.TempDir("", "0stortest")
	require.NoError(err)

	db, err := badger.New(path.Join(tmpDir, "data"), path.Join(tmpDir, "meta"))
	if err != nil {
		require.NoError(err)
	}

	clean := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	require.NoError(err)
	api := NewNamespaceAPI(db)

	return api, clean
}
