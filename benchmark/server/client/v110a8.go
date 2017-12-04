package client

import (
	"context"
	"math"

	pb "github.com/zero-os/0-stor/benchmark/server/schemav110a8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// V110a8 represents a client to a V1.1.0 alpha 8 0-stor server
type V110a8 struct {
	conn             *grpc.ClientConn
	objService       pb.ObjectManagerClient
	namespaceService pb.NamespaceManagerClient

	jwtToken  string
	namespace string
}

// NewV110a8Client return a new  V1.1.0 alpha 8 0-stor benchmarking client
func NewV110a8Client(addr, namespace, jwtToken string) (*V110a8, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
			grpc.MaxCallSendMsgSize(math.MaxInt32),
		))
	if err != nil {
		return nil, err
	}

	return &V110a8{
		conn:             conn,
		objService:       pb.NewObjectManagerClient(conn),
		namespaceService: pb.NewNamespaceManagerClient(conn),
		jwtToken:         jwtToken,
		namespace:        namespace,
	}, nil
}

// ObjectCreate implements client.ObjectCreate
func (c *V110a8) ObjectCreate(id, data []byte, refList []string) error {
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
func (c *V110a8) ObjectGet(id []byte) (*pb.Object, error) {
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

func contextWithMetadata(jwt, label string) context.Context {
	md := metadata.Pairs("authorization", jwt, "label", label)
	return metadata.NewOutgoingContext(context.Background(), md)
}
