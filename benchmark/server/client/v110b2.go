package client

import (
	"context"
	"math"

	pb "github.com/zero-os/0-stor/benchmark/server/schemav110b2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// V110b2 represents a client to a V1.1.0 beta 2 0-stor server
type V110b2 struct {
	conn             *grpc.ClientConn
	objService       pb.ObjectManagerClient
	namespaceService pb.NamespaceManagerClient

	jwtToken  string
	namespace string
}

// NewV110b2Client return a new  V1.1.0 beta 2 0-stor benchmarking client
func NewV110b2Client(addr, namespace, jwtToken string) (*V110b2, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
			grpc.MaxCallSendMsgSize(math.MaxInt32),
		))
	if err != nil {
		return nil, err
	}

	return &V110b2{
		conn:             conn,
		objService:       pb.NewObjectManagerClient(conn),
		namespaceService: pb.NewNamespaceManagerClient(conn),
		jwtToken:         jwtToken,
		namespace:        namespace,
	}, nil
}

// ObjectCreate implements client.ObjectCreate
func (c *V110b2) ObjectCreate(id, data []byte, refList []string) error {
	_, err := c.objService.Create(contextWithMetadata(c.jwtToken, c.namespace), &pb.CreateObjectRequest{
		Label: c.namespace,
		Object: &pb.Object{
			Key:           id,
			Value:         data,
			ReferenceList: refList,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// ObjectGet implements client.ObjectGet
func (c *V110b2) ObjectGet(id []byte) (*pb.Object, error) {
	resp, err := c.objService.Get(contextWithMetadata(c.jwtToken, c.namespace), &pb.GetObjectRequest{
		Label: c.namespace,
		Key:   id,
	})
	if err != nil {
		return nil, err
	}

	obj := resp.GetObject()
	return obj, nil
}
