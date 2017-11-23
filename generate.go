//go:generate protoc -I specs/protobuf  specs/protobuf/store.proto --go_out=plugins=grpc:grpc_store
//go:generate protoc -I specs/protobuf  specs/protobuf/proxy.proto --go_out=plugins=grpc:proxy/pb

package main
