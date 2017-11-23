package proxy

import (
	"net"

	"github.com/zero-os/0-stor/client"
	pb "github.com/zero-os/0-stor/proxy/pb"
	"google.golang.org/grpc"
)

type Proxy struct {
	grpcServer *grpc.Server
	client     *client.Client
	listener   net.Listener
}

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

func (p *Proxy) Listen(addr string) (listenAddr string, err error) {
	p.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	go p.grpcServer.Serve(p.listener)

	listenAddr = p.listener.Addr().String()

	return
}
