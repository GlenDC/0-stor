package stor

import (
	"context"
	"fmt"
	"io"

	"github.com/zero-os/0-stor/server/api"
	pb "github.com/zero-os/0-stor/server/schema"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ErrNotImplemented is the error return when a method is not implemented on the client
var ErrNotImplemented = fmt.Errorf("not implemented")

// client implement the stor.Client interface using grpc
type client struct {
	conn             *grpc.ClientConn
	objService       pb.ObjectManagerClient
	namespaceService pb.NamespaceManagerClient

	jwtToken  string
	namespace string
}

// New create a grpc client for the 0-stor
func newGrpcClient(conn *grpc.ClientConn, namespace, jwtToken string) *client {
	return &client{
		conn:             conn,
		objService:       pb.NewObjectManagerClient(conn),
		namespaceService: pb.NewNamespaceManagerClient(conn),
		jwtToken:         jwtToken,
		namespace:        namespace,
	}
}

func (c *client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *client) NamespaceGet() (*pb.Namespace, error) {
	resp, err := c.namespaceService.Get(contextWithMetadata(c.jwtToken, c.namespace), &pb.GetNamespaceRequest{Label: c.namespace})
	if err != nil {
		return nil, err
	}

	namespace := resp.GetNamespace()
	return namespace, nil
}

func (c *client) ObjectList(page, perPage int) ([]string, error) {
	stream, err := c.objService.List(contextWithMetadata(c.jwtToken, c.namespace), &pb.ListObjectsRequest{Label: c.namespace})
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, 100)

	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		keys = append(keys, string(obj.GetKey()))
	}

	return keys, nil
}

func (c *client) ObjectCreate(id, data []byte, refList []string) error {
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

func (c *client) ObjectGet(id []byte) (*pb.Object, error) {
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

func (c *client) ObjectDelete(id []byte) error {
	_, err := c.objService.Delete(contextWithMetadata(c.jwtToken, c.namespace), &pb.DeleteObjectRequest{
		Label: c.namespace,
		Key:   id,
	})

	return err
}

func (c *client) ObjectExist(id []byte) (bool, error) {
	resp, err := c.objService.Exists(contextWithMetadata(c.jwtToken, c.namespace), &pb.ExistsObjectRequest{
		Label: c.namespace,
		Key:   id,
	})

	return resp.GetExists(), err
}

func (c *client) ReferenceSet(id []byte, refList []string) error {
	_, err := c.objService.SetReferenceList(contextWithMetadata(c.jwtToken, c.namespace), &pb.UpdateReferenceListRequest{
		Label:         c.namespace,
		Key:           id,
		ReferenceList: refList,
	})

	return err
}

func (c *client) ReferenceAppend(id []byte, refList []string) error {
	_, err := c.objService.AppendReferenceList(contextWithMetadata(c.jwtToken, c.namespace), &pb.UpdateReferenceListRequest{
		Label:         c.namespace,
		Key:           id,
		ReferenceList: refList,
	})

	return err
}

func (c *client) ObjectsCheck(ids [][]byte) ([]*pb.CheckResponse, error) {
	sids := make([]string, len(ids))
	for i, id := range ids {
		sids[i] = string(id)
	}

	stream, err := c.objService.Check(contextWithMetadata(c.jwtToken, c.namespace), &pb.CheckRequest{Ids: sids, Label: c.namespace})
	if err != nil {
		return nil, err
	}

	checkResps := make([]*pb.CheckResponse, 0, len(sids))
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		checkResps = append(checkResps, &pb.CheckResponse{
			Id:     resp.Id,
			Status: resp.GetStatus(),
		})
	}

	return checkResps, nil
}

func (c *client) ReferenceRemove(id []byte, refList []string) error {
	_, err := c.objService.RemoveReferenceList(contextWithMetadata(c.jwtToken, c.namespace), &pb.UpdateReferenceListRequest{
		Label:         c.namespace,
		Key:           id,
		ReferenceList: refList,
	})

	return err
}
func contextWithMetadata(jwt, label string) context.Context {
	md := metadata.Pairs(api.GRPCMetaAuthKey, jwt, api.GRPCMetaLabelKey, label)
	return metadata.NewOutgoingContext(context.Background(), md)
}
