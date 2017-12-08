package grpc

import (
	"golang.org/x/net/context"

	serverAPI "github.com/zero-os/0-stor/server/api"
	pb "github.com/zero-os/0-stor/server/api/grpc/schema"
	"github.com/zero-os/0-stor/server/db"
	"github.com/zero-os/0-stor/server/stats"
)

var _ (pb.NamespaceManagerServer) = (*NamespaceAPI)(nil)

// NamespaceAPI implements pb.NamespaceManagerServer
type NamespaceAPI struct {
	db db.DB
}

// NewNamespaceAPI returns a NamespaceAPI
func NewNamespaceAPI(db db.DB) *NamespaceAPI {
	if db == nil {
		panic("no database given to NamespaceAPI")
	}

	return &NamespaceAPI{
		db: db,
	}
}

// GetNamespace implements NamespaceManagerServer.GetNamespace
func (api *NamespaceAPI) GetNamespace(ctx context.Context, req *pb.GetNamespaceRequest) (*pb.GetNamespaceResponse, error) {
	label, err := extractStringFromContext(ctx, serverAPI.GRPCMetaLabelKey)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	count, err := db.CountKeys(api.db, []byte(label))
	if err != nil {
		return nil, err
	}
	read, write := stats.Rate(label)

	resp := &pb.GetNamespaceResponse{
		Label:               label,
		ReadRequestPerHour:  read,
		WriteRequestPerHour: write,
		NrObjects:           int64(count),
	}

	return resp, nil
}
