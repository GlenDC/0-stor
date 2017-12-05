package proxy

import (
	"net"

	"github.com/zero-os/0-stor/client"
	pb "github.com/zero-os/0-stor/proxy/pb"
	"google.golang.org/grpc"
)

// Proxy represents a client proxy
type Proxy struct {
	grpcServer *grpc.Server
	client     *client.Client
	listener   net.Listener
}

// New creates new client proxy
func New(pol client.Policy) (*Proxy, error) {
	client, err := client.New(pol)
	if err != nil {
		return nil, err
	}

	var opts []grpc.ServerOption
	proxy := &Proxy{
		grpcServer: grpc.NewServer(opts...),
		client:     client,
	}
	pb.RegisterObjectServiceServer(proxy.grpcServer, newObjectSrv(client))
	pb.RegisterNamespaceServiceServer(proxy.grpcServer, newNamespaceSrv(client))

	return proxy, nil
}

// Listen listens to specified address
func (p *Proxy) Listen(addr string) (err error) {
	p.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	return p.grpcServer.Serve(p.listener)
}
