package grpc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	clientGRPC "github.com/zero-os/0-stor/client/datastor/grpc"
	metastor "github.com/zero-os/0-stor/client/metastor/badger"
	"github.com/zero-os/0-stor/client/pipeline"
	"github.com/zero-os/0-stor/client/pipeline/storage"
	"github.com/zero-os/0-stor/server/api/grpc"
	"github.com/zero-os/0-stor/server/db/memory"
)

const testLabel = "testLabel"

func newTestDaemon(t *testing.T) *Daemon {
	dataCluster, cleanup, err := newServerCluster(1)
	require.NoError(t, err)

	chunkStorage, err := storage.NewRandomChunkStorage(dataCluster)
	if err != nil {
		cleanup()
		t.Fatal(err)
	}
	dataPipeline := pipeline.NewSingleObjectPipeline(chunkStorage,
		pipeline.DefaultProcessorConstructor, pipeline.DefaultHasherConstructor)

	tmpDir, err := ioutil.TempDir("", "0-stor-test-daemon")
	require.NoError(t, err)
	dataDir := path.Join(tmpDir, "data")
	metaDir := path.Join(tmpDir, "meta")
	metaClient, err := metastor.NewClient(dataDir, metaDir, nil)
	if err != nil {
		cleanup()
		t.Fatal(err)
	}
	cleanupB := func() {
		err := metaClient.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(tmpDir)
		if err != nil {
			t.Fatal(err)
		}
		cleanup()
	}

	daemon, err := New(Config{
		Pipeline:   dataPipeline,
		MetaClient: metaClient,
		IYOClient:  nil})
	if err != nil {
		cleanupB()
		t.Fatal(err)
	}
	daemon.closer = &testCloser{func() error {
		cleanupB()
		return nil
	}}
	return daemon
}

type testCloser struct {
	cb func() error
}

func (tc *testCloser) Close() error {
	return tc.cb()
}

func newServerCluster(count int) (*clientGRPC.Cluster, func(), error) {
	if count < 1 {
		return nil, nil, errors.New("invalid GRPC server-client count")
	}
	var (
		cleanupSlice []func()
		addressSlice []string
	)
	for i := 0; i < count; i++ {
		_, addr, cleanup, err := newServerClient()
		if err != nil {
			for _, cleanup := range cleanupSlice {
				cleanup()
			}
			return nil, nil, err
		}
		cleanupSlice = append(cleanupSlice, cleanup)
		addressSlice = append(addressSlice, addr)
	}

	cluster, err := clientGRPC.NewCluster(addressSlice, testLabel, nil)
	if err != nil {
		for _, cleanup := range cleanupSlice {
			cleanup()
		}
		return nil, nil, err
	}

	cleanup := func() {
		cluster.Close()
		for _, cleanup := range cleanupSlice {
			cleanup()
		}
	}
	return cluster, cleanup, nil
}

func newServer() (string, func(), error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, err
	}

	server, err := grpc.New(memory.New(), nil, 0, 0)
	if err != nil {
		return "", nil, err
	}

	go func() {
		err := server.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()
	cleanup := func() {
		err := server.Close()
		if err != nil {
			panic(err)
		}
	}
	return listener.Addr().String(), cleanup, nil
}

func newServerClient() (*clientGRPC.Client, string, func(), error) {
	addr, cleanup, err := newServer()
	if err != nil {
		return nil, "", nil, err
	}

	client, err := clientGRPC.NewClient(addr, testLabel, nil)
	if err != nil {
		cleanup()
		return nil, "", nil, err
	}

	clean := func() {
		fmt.Sprintln("clean called")
		client.Close()
		cleanup()
	}

	return client, addr, clean, nil
}
