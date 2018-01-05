package grpc

import (
	"context"
	"io/ioutil"
	"net"
	"testing"

	"github.com/zero-os/0-stor/daemon/api/grpc/rpctypes"
	pb "github.com/zero-os/0-stor/daemon/api/grpc/schema"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestDataAPI_Client_Key_ReadWriteDeleteCheck(t *testing.T) {
	require := require.New(t)

	daemon := newTestDaemon(t)
	require.NotNil(daemon)
	defer daemon.Close()

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)
	go func() {
		err := daemon.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(err, "can't connect to the server")

	client := pb.NewDataServiceClient(conn)
	require.NotNil(client)

	ctx := context.Background()

	writeResp, err := client.Write(ctx, &pb.DataWriteRequest{Data: []byte("bar")})
	require.NoError(err)
	require.NotNil(writeResp)
	chunks := writeResp.GetChunks()
	require.Len(chunks, 1)
	require.Len(chunks[0].GetObjects(), 1)

	readResp, err := client.Read(ctx, &pb.DataReadRequest{Chunks: chunks})
	require.NoError(err)
	require.NotNil(readResp)
	data := readResp.GetData()
	require.Equal([]byte("bar"), data)

	checkResp, err := client.Check(ctx, &pb.DataCheckRequest{Chunks: chunks})
	require.NoError(err)
	require.NotNil(checkResp)
	status := checkResp.GetStatus()
	require.Equal(pb.CheckStatusOptimal, status)

	deleteResp, err := client.Delete(ctx, &pb.DataDeleteRequest{Chunks: chunks})
	require.NoError(err)
	require.NotNil(deleteResp)

	checkResp, err = client.Check(ctx, &pb.DataCheckRequest{Chunks: chunks})
	require.NoError(err)
	require.NotNil(checkResp)
	status = checkResp.GetStatus()
	require.Equal(pb.CheckStatusInvalid, status)

	_, err = client.Read(ctx, &pb.DataReadRequest{Chunks: chunks})
	require.Equal(rpctypes.ErrGRPCKeyNotFound, err)
}

func TestDataAPI_Client_Key_ReadFileWriteFileCheck(t *testing.T) {
	require := require.New(t)

	daemon := newTestDaemon(t)
	require.NotNil(daemon)
	defer daemon.Close()

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)
	go func() {
		err := daemon.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(err, "can't connect to the server")

	client := pb.NewDataServiceClient(conn)
	require.NotNil(client)

	ctx := context.Background()

	inputFile := newInMemoryFile()
	defer inputFile.Close()
	_, err = inputFile.Write([]byte("bar"))
	require.NoError(err)

	writeResp, err := client.WriteFile(ctx, &pb.DataWriteFileRequest{FilePath: inputFile.Name()})
	require.NoError(err)
	require.NotNil(writeResp)
	chunks := writeResp.GetChunks()
	require.Len(chunks, 1)
	require.Len(chunks[0].GetObjects(), 1)

	outputFile := newInMemoryFile()
	defer outputFile.Close()

	readResp, err := client.ReadFile(ctx,
		&pb.DataReadFileRequest{Chunks: chunks,
			FilePath: outputFile.Name(), SynchronousIO: true})
	require.NoError(err)
	require.NotNil(readResp)

	data, err := ioutil.ReadAll(outputFile)
	require.NoError(err)
	require.Equal([]byte("bar"), data)

	checkResp, err := client.Check(ctx, &pb.DataCheckRequest{Chunks: chunks})
	require.NoError(err)
	require.NotNil(checkResp)
	status := checkResp.GetStatus()
	require.Equal(pb.CheckStatusOptimal, status)
}
