package grpc

import (
	"context"
	"errors"
	"io"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/client/datastor"
	"github.com/zero-os/0-stor/server"

	"github.com/zero-os/0-stor/server/api/grpc/rpctypes"
	pb "github.com/zero-os/0-stor/server/api/grpc/schema"
	"google.golang.org/grpc"
)

func TestNewClientPanics(t *testing.T) {
	require := require.New(t)

	require.Panics(func() {
		NewClient("", "", "")
	}, "no address given")
	require.Panics(func() {
		NewClient("foo", "", "")
	}, "no label given")
}

func TestClientSetObject(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	err := client.SetObject(datastor.Object{})
	require.NoError(err, "server returns no error -> no error")

	os.err = rpctypes.ErrGRPCNilLabel
	err = client.SetObject(datastor.Object{})
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")

	errFoo := errors.New("fooErr")
	os.err = errFoo
	err = client.SetObject(datastor.Object{})
	require.Equal(errFoo, err, "any other error -> client error")
}

func TestClientGetObject(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	obj, err := client.GetObject(nil)
	require.Equal(datastor.ErrMissingData, err)
	require.Nil(obj)

	os.data = []byte("foo")
	obj, err = client.GetObject(nil)
	require.NoError(err)
	require.NotNil(obj)
	require.Nil(obj.Key)
	require.Equal([]byte("foo"), obj.Data)
	require.Empty(obj.ReferenceList)

	os.err = rpctypes.ErrGRPCNilLabel
	obj, err = client.GetObject(nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")
	require.Nil(obj)

	errFoo := errors.New("fooErr")
	os.err = errFoo
	obj, err = client.GetObject(nil)
	require.Equal(errFoo, err, "any other error -> client error")
	require.Nil(obj)
}

func TestClientDeleteObject(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	err := client.DeleteObject(nil)
	require.NoError(err, "server returns no error -> no error")

	os.err = rpctypes.ErrGRPCNilLabel
	err = client.DeleteObject(nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")

	errFoo := errors.New("fooErr")
	os.err = errFoo
	err = client.DeleteObject(nil)
	require.Equal(errFoo, err, "any other error -> client error")
}

func TestClientGetObjectStatus(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	status, err := client.GetObjectStatus(nil)
	require.NoError(err)
	require.Equal(datastor.ObjectStatusMissing, status)

	os.status = math.MaxInt32
	status, err = client.GetObjectStatus(nil)
	require.Equal(datastor.ErrInvalidStatus, err)
	require.Equal(datastor.ObjectStatus(0), status)

	os.status = pb.ObjectStatusCorrupted
	status, err = client.GetObjectStatus(nil)
	require.NoError(err)
	require.Equal(datastor.ObjectStatusCorrupted, status)

	os.err = rpctypes.ErrGRPCNilLabel
	status, err = client.GetObjectStatus(nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")
	require.Equal(datastor.ObjectStatus(0), status)

	errFoo := errors.New("fooErr")
	os.err = errFoo
	status, err = client.GetObjectStatus(nil)
	require.Equal(errFoo, err, "any other error -> client error")
	require.Equal(datastor.ObjectStatus(0), status)
}

func TestClientExistObject(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	exists, err := client.ExistObject(nil)
	require.NoError(err)
	require.False(exists)

	os.status = math.MaxInt32
	exists, err = client.ExistObject(nil)
	require.Equal(datastor.ErrInvalidStatus, err)
	require.False(exists)

	os.status = pb.ObjectStatusCorrupted
	exists, err = client.ExistObject(nil)
	require.Equal(datastor.ErrObjectCorrupted, err)
	require.False(exists)

	os.status = pb.ObjectStatusOK
	exists, err = client.ExistObject(nil)
	require.NoError(err)
	require.True(exists)

	os.status = pb.ObjectStatusMissing
	exists, err = client.ExistObject(nil)
	require.NoError(err)
	require.False(exists)

	os.err = rpctypes.ErrGRPCNilLabel
	exists, err = client.ExistObject(nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")
	require.False(exists)

	errFoo := errors.New("fooErr")
	os.err = errFoo
	exists, err = client.ExistObject(nil)
	require.Equal(errFoo, err, "any other error -> client error")
	require.False(exists)
}

func TestClientGetNamespace(t *testing.T) {
	require := require.New(t)

	var nss stubNamespaceService
	client := Client{namespaceService: &nss, label: "myLabel"}

	ns, err := client.GetNamespace()
	require.Equal(datastor.ErrInvalidLabel, err, "returned label should be equal to client's label")
	require.Nil(ns)

	nss.label = "myLabel"
	ns, err = client.GetNamespace()
	require.NoError(err, "returned label should be equal to client's label")
	require.NotNil(ns)
	require.Equal("myLabel", ns.Label)
	require.Equal(int64(0), ns.ReadRequestPerHour)
	require.Equal(int64(0), ns.WriteRequestPerHour)
	require.Equal(int64(0), ns.NrObjects)

	nss.err = rpctypes.ErrGRPCNilLabel
	ns, err = client.GetNamespace()
	require.Equal(rpctypes.ErrNilLabel, err, "server errors should be caught by client")
	require.Nil(ns)

	nss.nrOfObjects = 42
	nss.readRPH = 1023
	nss.writeRPH = math.MaxInt64
	nss.err = nil
	ns, err = client.GetNamespace()
	require.NoError(err)
	require.NotNil(ns)
	require.Equal("myLabel", ns.Label)
	require.Equal(int64(1023), ns.ReadRequestPerHour)
	require.Equal(int64(math.MaxInt64), ns.WriteRequestPerHour)
	require.Equal(int64(42), ns.NrObjects)

	client.jwtTokenDefined = true
	ns, err = client.GetNamespace()
	require.NoError(err)
	require.NotNil(ns)
	require.Equal("myLabel", ns.Label)
	require.Equal(int64(1023), ns.ReadRequestPerHour)
	require.Equal(int64(math.MaxInt64), ns.WriteRequestPerHour)
	require.Equal(int64(42), ns.NrObjects)

	client.label = "foo"
	ns, err = client.GetNamespace()
	require.Equal(datastor.ErrInvalidLabel, err, "returned label should be equal to client's label")
	require.Nil(ns)
}

func TestClientListObjectKeys(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	require.Panics(func() {
		client.ListObjectKeyIterator(nil)
	}, "no context given")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := client.ListObjectKeyIterator(ctx)
	require.NoError(err, "server returns no error -> no error")
	require.NotNil(ch)
	select {
	case resp, open := <-ch:
		require.True(open)
		require.Equal(datastor.ErrMissingKey, resp.Error)
		require.Nil(resp.Key)
	case <-time.After(time.Millisecond * 500):
	}

	os.err = rpctypes.ErrGRPCNilLabel
	ch, err = client.ListObjectKeyIterator(context.Background())
	require.Equal(rpctypes.ErrNilLabel, err)
	require.Nil(ch)

	os.err = nil
	os.streamErr = datastor.ErrMissingKey
	ch, err = client.ListObjectKeyIterator(context.Background())
	require.NoError(err, "channel should be created")
	require.NotNil(ch)
	select {
	case resp, open := <-ch:
		require.True(open)
		require.Equal(datastor.ErrMissingKey, resp.Error)
		require.Nil(resp.Key)
	case <-time.After(time.Millisecond * 500):
		t.Fatal("nothing received from channel while it was expected")
	}
	select {
	case _, open := <-ch:
		require.False(open)
	case <-time.After(time.Millisecond * 500):
	}

	os.streamErr = nil
	os.key = []byte("foo")
	ch, err = client.ListObjectKeyIterator(context.Background())
	require.NoError(err, "channel should be created")
	require.NotNil(ch)
	select {
	case resp, open := <-ch:
		require.True(open)
		require.NoError(resp.Error)
		require.Equal([]byte("foo"), resp.Key)
	case <-time.After(time.Millisecond * 500):
		t.Fatal("nothing received from channel while it was expected")
	}
	select {
	case _, open := <-ch:
		require.False(open)
	case <-time.After(time.Millisecond * 500):
	}
}

func TestClientSetReferenceList(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	err := client.SetReferenceList(nil, nil)
	require.NoError(err, "server returns no error -> no error")

	os.err = rpctypes.ErrGRPCNilLabel
	err = client.SetReferenceList(nil, nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")

	errFoo := errors.New("fooErr")
	os.err = errFoo
	err = client.SetReferenceList(nil, nil)
	require.Equal(errFoo, err, "any other error -> client error")
}

func TestClientGetReferenceList(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	refList, err := client.GetReferenceList(nil)
	require.Equal(datastor.ErrMissingRefList, err)
	require.Nil(refList)

	os.refList = []string{"user1"}
	refList, err = client.GetReferenceList(nil)
	require.NoError(err, "server returns no error -> no error")
	require.Equal([]string{"user1"}, refList)

	os.err = rpctypes.ErrGRPCNilLabel
	refList, err = client.GetReferenceList(nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")
	require.Nil(refList)

	errFoo := errors.New("fooErr")
	os.err = errFoo
	refList, err = client.GetReferenceList(nil)
	require.Equal(errFoo, err, "any other error -> client error")
	require.Nil(refList)
}

func TestClientGetReferenceCount(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	count, err := client.GetReferenceCount(nil)
	require.NoError(err, "server returns no error -> no error")
	require.Equal(int64(0), count)

	os.refList = []string{"user1", "user2"}
	count, err = client.GetReferenceCount(nil)
	require.NoError(err, "server returns no error -> no error")
	require.Equal(int64(2), count)

	os.err = rpctypes.ErrGRPCNilLabel
	count, err = client.GetReferenceCount(nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")
	require.Equal(int64(0), count)

	errFoo := errors.New("fooErr")
	os.err = errFoo
	count, err = client.GetReferenceCount(nil)
	require.Equal(errFoo, err, "any other error -> client error")
	require.Equal(int64(0), count)
}

func TestClientAppendToReferenceList(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	err := client.AppendToReferenceList(nil, nil)
	require.NoError(err, "server returns no error -> no error")

	os.err = rpctypes.ErrGRPCNilLabel
	err = client.AppendToReferenceList(nil, nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")

	errFoo := errors.New("fooErr")
	os.err = errFoo
	err = client.AppendToReferenceList(nil, nil)
	require.Equal(errFoo, err, "any other error -> client error")
}

func TestClientDeleteFromReferenceList(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	count, err := client.DeleteFromReferenceList(nil, nil)
	require.NoError(err, "server returns no error -> no error")
	require.Equal(int64(0), count)

	os.refList = []string{"user1"}
	count, err = client.DeleteFromReferenceList(nil, nil)
	require.NoError(err, "server returns no error -> no error")
	require.Equal(int64(1), count)

	count, err = client.DeleteFromReferenceList(nil, []string{"user2", "user1"})
	require.NoError(err, "server returns no error -> no error")
	require.Equal(int64(0), count)

	os.err = rpctypes.ErrGRPCNilLabel
	count, err = client.DeleteFromReferenceList(nil, nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")
	require.Equal(int64(0), count)

	errFoo := errors.New("fooErr")
	os.err = errFoo
	count, err = client.DeleteFromReferenceList(nil, nil)
	require.Equal(errFoo, err, "any other error -> client error")
	require.Equal(int64(0), count)
}

func TestClientDeleteReferenceList(t *testing.T) {
	require := require.New(t)

	var os stubObjectService
	client := Client{objService: &os}

	err := client.DeleteReferenceList(nil)
	require.NoError(err, "server returns no error -> no error")

	os.err = rpctypes.ErrGRPCNilLabel
	err = client.DeleteReferenceList(nil)
	require.Equal(rpctypes.ErrNilLabel, err, "server returns GRPC error -> client error")

	errFoo := errors.New("fooErr")
	os.err = errFoo
	err = client.DeleteReferenceList(nil)
	require.Equal(errFoo, err, "any other error -> client error")
}

func TestClientClosePanic(t *testing.T) {
	client := new(Client)
	require.Panics(t, func() {
		client.Close()
	}, "nil pointer dereference, because connection property is not set")
}

type stubObjectService struct {
	key, data []byte
	refList   []string
	status    pb.ObjectStatus
	err       error
	streamErr error
}

// SetObject implements pb.ObjectService.SetObject
func (os *stubObjectService) SetObject(ctx context.Context, in *pb.SetObjectRequest, opts ...grpc.CallOption) (*pb.SetObjectResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &pb.SetObjectResponse{}, nil
}

// GetObject implements pb.ObjectService.GetObject
func (os *stubObjectService) GetObject(ctx context.Context, in *pb.GetObjectRequest, opts ...grpc.CallOption) (*pb.GetObjectResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &pb.GetObjectResponse{
		Data:          os.data,
		ReferenceList: os.refList,
	}, nil
}

// DeleteObject implements pb.ObjectService.DeleteObject
func (os *stubObjectService) DeleteObject(ctx context.Context, in *pb.DeleteObjectRequest, opts ...grpc.CallOption) (*pb.DeleteObjectResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &pb.DeleteObjectResponse{}, nil
}

// GetObjectStatus implements pb.ObjectService.GetObjectStatus
func (os *stubObjectService) GetObjectStatus(ctx context.Context, in *pb.GetObjectStatusRequest, opts ...grpc.CallOption) (*pb.GetObjectStatusResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &pb.GetObjectStatusResponse{
		Status: os.status,
	}, nil
}

// ListObjectKeys implements pb.ObjectService.ListObjectKeys
func (os *stubObjectService) ListObjectKeys(ctx context.Context, in *pb.ListObjectKeysRequest, opts ...grpc.CallOption) (pb.ObjectManager_ListObjectKeysClient, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &stubListObjectKeysClient{
		ClientStream: nil,
		eof:          os.key != nil,
		key:          os.key,
		err:          os.streamErr,
	}, nil
}

type stubListObjectKeysClient struct {
	grpc.ClientStream
	key []byte
	err error
	eof bool
}

// Recv implements pb.ObjectManager_ListObjectKeysClient.Recv
func (stream *stubListObjectKeysClient) Recv() (*pb.ListObjectKeysResponse, error) {
	if stream.err != nil {
		return nil, stream.err
	}
	if stream.key == nil {
		if stream.eof {
			return nil, io.EOF
		}
		return &pb.ListObjectKeysResponse{}, nil
	}
	resp := &pb.ListObjectKeysResponse{Key: stream.key}
	stream.key = nil
	return resp, nil
}

// SetReferenceList implements pb.ObjectService.SetReferenceList
func (os *stubObjectService) SetReferenceList(ctx context.Context, in *pb.SetReferenceListRequest, opts ...grpc.CallOption) (*pb.SetReferenceListResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &pb.SetReferenceListResponse{}, nil
}

// GetReferenceList implements pb.ObjectService.GetReferenceList
func (os *stubObjectService) GetReferenceList(ctx context.Context, in *pb.GetReferenceListRequest, opts ...grpc.CallOption) (*pb.GetReferenceListResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &pb.GetReferenceListResponse{
		ReferenceList: os.refList,
	}, nil
}

// GetReferenceCount implements pb.ObjectService.GetReferenceCount
func (os *stubObjectService) GetReferenceCount(ctx context.Context, in *pb.GetReferenceCountRequest, opts ...grpc.CallOption) (*pb.GetReferenceCountResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &pb.GetReferenceCountResponse{
		Count: int64(len(os.refList)),
	}, nil
}

// AppendToReferenceList implements pb.ObjectService.AppendToReferenceList
func (os *stubObjectService) AppendToReferenceList(ctx context.Context, in *pb.AppendToReferenceListRequest, opts ...grpc.CallOption) (*pb.AppendToReferenceListResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &pb.AppendToReferenceListResponse{}, nil
}

// DeleteFromReferenceList implements pb.ObjectService.DeleteFromReferenceList
func (os *stubObjectService) DeleteFromReferenceList(ctx context.Context, in *pb.DeleteFromReferenceListRequest, opts ...grpc.CallOption) (*pb.DeleteFromReferenceListResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	refList := server.ReferenceList(os.refList)
	refList.RemoveReferences(server.ReferenceList(in.ReferenceList))
	os.refList = []string(refList)
	return &pb.DeleteFromReferenceListResponse{
		Count: int64(len(os.refList)),
	}, nil
}

// DeleteReferenceList implements pb.ObjectService.DeleteReferenceList
func (os *stubObjectService) DeleteReferenceList(ctx context.Context, in *pb.DeleteReferenceListRequest, opts ...grpc.CallOption) (*pb.DeleteReferenceListResponse, error) {
	if os.err != nil {
		return nil, os.err
	}
	return &pb.DeleteReferenceListResponse{}, nil
}

type stubNamespaceService struct {
	label                          string
	readRPH, writeRPH, nrOfObjects int64
	err                            error
}

// GetNamespace implements pb.NamespaceService.GetNamespace
func (ns *stubNamespaceService) GetNamespace(ctx context.Context, in *pb.GetNamespaceRequest, opts ...grpc.CallOption) (*pb.GetNamespaceResponse, error) {
	if ns.err != nil {
		return nil, ns.err
	}
	return &pb.GetNamespaceResponse{
		Label:               ns.label,
		ReadRequestPerHour:  ns.readRPH,
		WriteRequestPerHour: ns.writeRPH,
		NrObjects:           ns.nrOfObjects,
	}, nil
}
