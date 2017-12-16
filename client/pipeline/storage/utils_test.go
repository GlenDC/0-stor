package storage

import (
	"errors"
	"fmt"
	"math"

	"github.com/zero-os/0-stor/client/datastor"
	clientGRPC "github.com/zero-os/0-stor/client/datastor/grpc"
	serverGRPC "github.com/zero-os/0-stor/server/api/grpc"
	"github.com/zero-os/0-stor/server/db/memory"
)

func newGRPCServerCluster(count int) (*clientGRPC.Cluster, func(), error) {
	if count < 1 {
		return nil, nil, errors.New("invalid GRPC server-client count")
	}
	var (
		cleanupSlice []func()
		addressSlice []string
	)
	for i := 0; i < count; i++ {
		_, addr, cleanup, err := newGRPCServerClient()
		if err != nil {
			for _, cleanup := range cleanupSlice {
				cleanup()
			}
			return nil, nil, err
		}
		cleanupSlice = append(cleanupSlice, cleanup)
		addressSlice = append(addressSlice, addr)
	}

	cluster, err := clientGRPC.NewCluster(addressSlice, "myLabel", "")
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

func newGRPCServerClient() (*clientGRPC.Client, string, func(), error) {
	server, err := serverGRPC.New(memory.New(), nil, 0, 0)
	if err != nil {
		return nil, "", nil, err
	}
	go func() {
		err := server.Listen("localhost:0")
		if err != nil {
			panic(err)
		}
	}()

	client, err := clientGRPC.NewClient(server.Address(), "myLabel", "")
	if err != nil {
		server.Close()
		return nil, "", nil, err
	}

	clean := func() {
		fmt.Sprintln("clean called")
		client.Close()
		server.Close()
	}

	return client, server.Address(), clean, nil
}

type dummyCluster struct{}

func (dc dummyCluster) GetShard(id string) (datastor.Shard, error) { panic("dummy::GetShard") }
func (dc dummyCluster) GetRandomShard() (datastor.Shard, error)    { panic("dummy::GetRandomShard") }
func (dc dummyCluster) GetRandomShardIterator(exceptShards []string) datastor.ShardIterator {
	panic("dummy::GetRandomShardIterator")
}
func (dc dummyCluster) ListedShardCount() int {
	return int(math.MaxInt32)
}

var (
	_ datastor.Cluster = dummyCluster{}
)
