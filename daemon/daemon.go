package daemon

import (
	"net"

	"github.com/zero-os/0-stor/client"
	pb "github.com/zero-os/0-stor/daemon/pb"
	"google.golang.org/grpc"
)

// Daemon represents a client proxy
type Daemon struct {
	grpcServer *grpc.Server
	client     *client.Client
	listener   net.Listener
}

// New creates new client proxy
func New(pol client.Policy, maxMsgSize int) (*Daemon, error) {
	client, err := client.New(pol)
	if err != nil {
		return nil, err
	}

	maxMsgSize = maxMsgSize * 1024 * 1024 // Mib to bytes

	proxy := &Daemon{
		grpcServer: grpc.NewServer(
			grpc.MaxRecvMsgSize(maxMsgSize),
			grpc.MaxSendMsgSize(maxMsgSize),
		),
		client: client,
	}
	pb.RegisterObjectServiceServer(proxy.grpcServer, newObjectSrv(client))
	pb.RegisterNamespaceServiceServer(proxy.grpcServer, newNamespaceSrv(client))

	return proxy, nil
}

// Listen listens to specified address
func (d *Daemon) Listen(addr string) error {
	list, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return d.grpcServer.Serve(list)
}
