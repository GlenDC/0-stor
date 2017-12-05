//go:generate protoc -I schema schema/ztor.proto --go_out=plugins=grpc,import_path=schema:schema
//go:generate protoc -I schema schema/proxy.proto --go_out=plugins=grpc:../proxy/pb

package server
