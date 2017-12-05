package stor

import (
	"math"

	pb "github.com/zero-os/0-stor/server/schema"
	"google.golang.org/grpc"
)

// Client defines client interface to talk with 0-stor server
type Client interface {
	// Namespace gets detail view about namespace
	NamespaceGet() (*pb.Namespace, error)

	// ObjectList lists keys of the object in the namespace
	ObjectList(page, perPage int) ([]string, error)

	// ObjectCreate creates an object
	ObjectCreate(id, data []byte, refList []string) error

	// ObjectGet retrieve object from the store
	ObjectGet(id []byte) (*pb.Object, error)

	// ObjectDelete delete object from the store
	ObjectDelete(id []byte) error

	// ObjectExist tests if an object with this id exists
	ObjectExist(id []byte) (bool, error)

	// ObjectCheck test if an object is still healty or corrupted
	ObjectsCheck(ids [][]byte) ([]*pb.CheckResponse, error)

	// ReferenceSet sets reference list
	ReferenceSet(id []byte, refList []string) error

	ReferenceRemove(id []byte, refList []string) error

	ReferenceAppend(id []byte, refList []string) error
}

// Config defines 0-stor client config
type Config struct {
	Shard string `yaml:"shard"` // 0-stor server address
}

// NewClient creates new 0-stor client
func NewClient(addr, namespace, token string) (Client, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
			grpc.MaxCallSendMsgSize(math.MaxInt32),
		))
	if err != nil {
		return nil, err
	}
	return newGrpcClient(conn, namespace, token), nil
}
