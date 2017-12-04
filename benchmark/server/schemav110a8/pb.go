// Code generated by protoc-gen-go. DO NOT EDIT.
// source: store.proto

/*
Package store is a generated protocol buffer package.

It is generated from these files:
	store.proto

It has these top-level messages:
	Emtpy
	Namespace
	Object
	GetNamespaceRequest
	GetNamespaceReply
	ListObjectsRequest
	CreateObjectRequest
	CreateObjectReply
	ExistsObjectRequest
	ExistsObjectReply
	GetObjectRequest
	GetObjectReply
	DeleteObjectRequest
	DeleteObjectReply
	UpdateReferenceListRequest
	UpdateReferenceListReply
	CheckRequest
	CheckResponse
*/
package schema

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type CheckResponse_Status int32

const (
	CheckResponse_OK        CheckResponse_Status = 0
	CheckResponse_CORRUPTED CheckResponse_Status = 1
	CheckResponse_MISSING   CheckResponse_Status = 2
)

var CheckResponse_Status_name = map[int32]string{
	0: "OK",
	1: "CORRUPTED",
	2: "MISSING",
}
var CheckResponse_Status_value = map[string]int32{
	"OK":        0,
	"CORRUPTED": 1,
	"MISSING":   2,
}

func (x CheckResponse_Status) String() string {
	return proto.EnumName(CheckResponse_Status_name, int32(x))
}
func (CheckResponse_Status) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{17, 0} }

// Types
type Emtpy struct {
}

func (m *Emtpy) Reset()                    { *m = Emtpy{} }
func (m *Emtpy) String() string            { return proto.CompactTextString(m) }
func (*Emtpy) ProtoMessage()               {}
func (*Emtpy) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// message ACL {
//     bool read = 1;
//     bool write = 2;
//     bool delete = 3;
//     bool admin = 4;
// }
type Namespace struct {
	Label               string `protobuf:"bytes,1,opt,name=label" json:"label,omitempty"`
	SpaceAvailale       int64  `protobuf:"varint,2,opt,name=spaceAvailale" json:"spaceAvailale,omitempty"`
	SpaceUsed           int64  `protobuf:"varint,3,opt,name=spaceUsed" json:"spaceUsed,omitempty"`
	ReadRequestPerHour  int64  `protobuf:"varint,4,opt,name=readRequestPerHour" json:"readRequestPerHour,omitempty"`
	WriteRequestPerHour int64  `protobuf:"varint,5,opt,name=writeRequestPerHour" json:"writeRequestPerHour,omitempty"`
	NrObjects           int64  `protobuf:"varint,6,opt,name=nrObjects" json:"nrObjects,omitempty"`
}

func (m *Namespace) Reset()                    { *m = Namespace{} }
func (m *Namespace) String() string            { return proto.CompactTextString(m) }
func (*Namespace) ProtoMessage()               {}
func (*Namespace) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Namespace) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *Namespace) GetSpaceAvailale() int64 {
	if m != nil {
		return m.SpaceAvailale
	}
	return 0
}

func (m *Namespace) GetSpaceUsed() int64 {
	if m != nil {
		return m.SpaceUsed
	}
	return 0
}

func (m *Namespace) GetReadRequestPerHour() int64 {
	if m != nil {
		return m.ReadRequestPerHour
	}
	return 0
}

func (m *Namespace) GetWriteRequestPerHour() int64 {
	if m != nil {
		return m.WriteRequestPerHour
	}
	return 0
}

func (m *Namespace) GetNrObjects() int64 {
	if m != nil {
		return m.NrObjects
	}
	return 0
}

type Object struct {
	Key           []byte   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value         []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	ReferenceList []string `protobuf:"bytes,3,rep,name=referenceList" json:"referenceList,omitempty"`
}

func (m *Object) Reset()                    { *m = Object{} }
func (m *Object) String() string            { return proto.CompactTextString(m) }
func (*Object) ProtoMessage()               {}
func (*Object) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Object) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *Object) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *Object) GetReferenceList() []string {
	if m != nil {
		return m.ReferenceList
	}
	return nil
}

// Namespace management
type GetNamespaceRequest struct {
	Label string `protobuf:"bytes,1,opt,name=label" json:"label,omitempty"`
}

func (m *GetNamespaceRequest) Reset()                    { *m = GetNamespaceRequest{} }
func (m *GetNamespaceRequest) String() string            { return proto.CompactTextString(m) }
func (*GetNamespaceRequest) ProtoMessage()               {}
func (*GetNamespaceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *GetNamespaceRequest) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

type GetNamespaceReply struct {
	Namespace *Namespace `protobuf:"bytes,1,opt,name=namespace" json:"namespace,omitempty"`
}

func (m *GetNamespaceReply) Reset()                    { *m = GetNamespaceReply{} }
func (m *GetNamespaceReply) String() string            { return proto.CompactTextString(m) }
func (*GetNamespaceReply) ProtoMessage()               {}
func (*GetNamespaceReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *GetNamespaceReply) GetNamespace() *Namespace {
	if m != nil {
		return m.Namespace
	}
	return nil
}

// Object management
type ListObjectsRequest struct {
	Label string `protobuf:"bytes,1,opt,name=label" json:"label,omitempty"`
}

func (m *ListObjectsRequest) Reset()                    { *m = ListObjectsRequest{} }
func (m *ListObjectsRequest) String() string            { return proto.CompactTextString(m) }
func (*ListObjectsRequest) ProtoMessage()               {}
func (*ListObjectsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ListObjectsRequest) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

type CreateObjectRequest struct {
	Label  string  `protobuf:"bytes,1,opt,name=Label" json:"Label,omitempty"`
	Object *Object `protobuf:"bytes,2,opt,name=object" json:"object,omitempty"`
}

func (m *CreateObjectRequest) Reset()                    { *m = CreateObjectRequest{} }
func (m *CreateObjectRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateObjectRequest) ProtoMessage()               {}
func (*CreateObjectRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *CreateObjectRequest) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *CreateObjectRequest) GetObject() *Object {
	if m != nil {
		return m.Object
	}
	return nil
}

type CreateObjectReply struct {
}

func (m *CreateObjectReply) Reset()                    { *m = CreateObjectReply{} }
func (m *CreateObjectReply) String() string            { return proto.CompactTextString(m) }
func (*CreateObjectReply) ProtoMessage()               {}
func (*CreateObjectReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type ExistsObjectRequest struct {
	Label string `protobuf:"bytes,1,opt,name=Label" json:"Label,omitempty"`
	Key   []byte `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *ExistsObjectRequest) Reset()                    { *m = ExistsObjectRequest{} }
func (m *ExistsObjectRequest) String() string            { return proto.CompactTextString(m) }
func (*ExistsObjectRequest) ProtoMessage()               {}
func (*ExistsObjectRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *ExistsObjectRequest) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *ExistsObjectRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type ExistsObjectReply struct {
	Exists bool `protobuf:"varint,1,opt,name=exists" json:"exists,omitempty"`
}

func (m *ExistsObjectReply) Reset()                    { *m = ExistsObjectReply{} }
func (m *ExistsObjectReply) String() string            { return proto.CompactTextString(m) }
func (*ExistsObjectReply) ProtoMessage()               {}
func (*ExistsObjectReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *ExistsObjectReply) GetExists() bool {
	if m != nil {
		return m.Exists
	}
	return false
}

type GetObjectRequest struct {
	Label string `protobuf:"bytes,1,opt,name=Label" json:"Label,omitempty"`
	Key   []byte `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *GetObjectRequest) Reset()                    { *m = GetObjectRequest{} }
func (m *GetObjectRequest) String() string            { return proto.CompactTextString(m) }
func (*GetObjectRequest) ProtoMessage()               {}
func (*GetObjectRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *GetObjectRequest) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *GetObjectRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type GetObjectReply struct {
	Object *Object `protobuf:"bytes,2,opt,name=object" json:"object,omitempty"`
}

func (m *GetObjectReply) Reset()                    { *m = GetObjectReply{} }
func (m *GetObjectReply) String() string            { return proto.CompactTextString(m) }
func (*GetObjectReply) ProtoMessage()               {}
func (*GetObjectReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *GetObjectReply) GetObject() *Object {
	if m != nil {
		return m.Object
	}
	return nil
}

type DeleteObjectRequest struct {
	Label string `protobuf:"bytes,1,opt,name=Label" json:"Label,omitempty"`
	Key   []byte `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *DeleteObjectRequest) Reset()                    { *m = DeleteObjectRequest{} }
func (m *DeleteObjectRequest) String() string            { return proto.CompactTextString(m) }
func (*DeleteObjectRequest) ProtoMessage()               {}
func (*DeleteObjectRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *DeleteObjectRequest) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *DeleteObjectRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type DeleteObjectReply struct {
}

func (m *DeleteObjectReply) Reset()                    { *m = DeleteObjectReply{} }
func (m *DeleteObjectReply) String() string            { return proto.CompactTextString(m) }
func (*DeleteObjectReply) ProtoMessage()               {}
func (*DeleteObjectReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

type UpdateReferenceListRequest struct {
	Label         string   `protobuf:"bytes,1,opt,name=label" json:"label,omitempty"`
	Key           []byte   `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	ReferenceList []string `protobuf:"bytes,3,rep,name=referenceList" json:"referenceList,omitempty"`
}

func (m *UpdateReferenceListRequest) Reset()                    { *m = UpdateReferenceListRequest{} }
func (m *UpdateReferenceListRequest) String() string            { return proto.CompactTextString(m) }
func (*UpdateReferenceListRequest) ProtoMessage()               {}
func (*UpdateReferenceListRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

func (m *UpdateReferenceListRequest) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *UpdateReferenceListRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *UpdateReferenceListRequest) GetReferenceList() []string {
	if m != nil {
		return m.ReferenceList
	}
	return nil
}

type UpdateReferenceListReply struct {
}

func (m *UpdateReferenceListReply) Reset()                    { *m = UpdateReferenceListReply{} }
func (m *UpdateReferenceListReply) String() string            { return proto.CompactTextString(m) }
func (*UpdateReferenceListReply) ProtoMessage()               {}
func (*UpdateReferenceListReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

type CheckRequest struct {
	Label string   `protobuf:"bytes,1,opt,name=label" json:"label,omitempty"`
	Ids   []string `protobuf:"bytes,2,rep,name=ids" json:"ids,omitempty"`
}

func (m *CheckRequest) Reset()                    { *m = CheckRequest{} }
func (m *CheckRequest) String() string            { return proto.CompactTextString(m) }
func (*CheckRequest) ProtoMessage()               {}
func (*CheckRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

func (m *CheckRequest) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *CheckRequest) GetIds() []string {
	if m != nil {
		return m.Ids
	}
	return nil
}

type CheckResponse struct {
	Id     string               `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Status CheckResponse_Status `protobuf:"varint,2,opt,name=status,enum=CheckResponse_Status" json:"status,omitempty"`
}

func (m *CheckResponse) Reset()                    { *m = CheckResponse{} }
func (m *CheckResponse) String() string            { return proto.CompactTextString(m) }
func (*CheckResponse) ProtoMessage()               {}
func (*CheckResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

func (m *CheckResponse) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *CheckResponse) GetStatus() CheckResponse_Status {
	if m != nil {
		return m.Status
	}
	return CheckResponse_OK
}

func init() {
	proto.RegisterType((*Emtpy)(nil), "Emtpy")
	proto.RegisterType((*Namespace)(nil), "Namespace")
	proto.RegisterType((*Object)(nil), "Object")
	proto.RegisterType((*GetNamespaceRequest)(nil), "GetNamespaceRequest")
	proto.RegisterType((*GetNamespaceReply)(nil), "GetNamespaceReply")
	proto.RegisterType((*ListObjectsRequest)(nil), "ListObjectsRequest")
	proto.RegisterType((*CreateObjectRequest)(nil), "CreateObjectRequest")
	proto.RegisterType((*CreateObjectReply)(nil), "CreateObjectReply")
	proto.RegisterType((*ExistsObjectRequest)(nil), "ExistsObjectRequest")
	proto.RegisterType((*ExistsObjectReply)(nil), "ExistsObjectReply")
	proto.RegisterType((*GetObjectRequest)(nil), "GetObjectRequest")
	proto.RegisterType((*GetObjectReply)(nil), "GetObjectReply")
	proto.RegisterType((*DeleteObjectRequest)(nil), "DeleteObjectRequest")
	proto.RegisterType((*DeleteObjectReply)(nil), "DeleteObjectReply")
	proto.RegisterType((*UpdateReferenceListRequest)(nil), "UpdateReferenceListRequest")
	proto.RegisterType((*UpdateReferenceListReply)(nil), "UpdateReferenceListReply")
	proto.RegisterType((*CheckRequest)(nil), "CheckRequest")
	proto.RegisterType((*CheckResponse)(nil), "CheckResponse")
	proto.RegisterEnum("CheckResponse_Status", CheckResponse_Status_name, CheckResponse_Status_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for NamespaceManager service

type NamespaceManagerClient interface {
	Get(ctx context.Context, in *GetNamespaceRequest, opts ...grpc.CallOption) (*GetNamespaceReply, error)
}

type namespaceManagerClient struct {
	cc *grpc.ClientConn
}

func NewNamespaceManagerClient(cc *grpc.ClientConn) NamespaceManagerClient {
	return &namespaceManagerClient{cc}
}

func (c *namespaceManagerClient) Get(ctx context.Context, in *GetNamespaceRequest, opts ...grpc.CallOption) (*GetNamespaceReply, error) {
	out := new(GetNamespaceReply)
	err := grpc.Invoke(ctx, "/NamespaceManager/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for NamespaceManager service

type NamespaceManagerServer interface {
	Get(context.Context, *GetNamespaceRequest) (*GetNamespaceReply, error)
}

func RegisterNamespaceManagerServer(s *grpc.Server, srv NamespaceManagerServer) {
	s.RegisterService(&_NamespaceManager_serviceDesc, srv)
}

func _NamespaceManager_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNamespaceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NamespaceManagerServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NamespaceManager/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NamespaceManagerServer).Get(ctx, req.(*GetNamespaceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _NamespaceManager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "NamespaceManager",
	HandlerType: (*NamespaceManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _NamespaceManager_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "store.proto",
}

// Client API for ObjectManager service

type ObjectManagerClient interface {
	Create(ctx context.Context, in *CreateObjectRequest, opts ...grpc.CallOption) (*CreateObjectReply, error)
	List(ctx context.Context, in *ListObjectsRequest, opts ...grpc.CallOption) (ObjectManager_ListClient, error)
	Get(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectReply, error)
	Exists(ctx context.Context, in *ExistsObjectRequest, opts ...grpc.CallOption) (*ExistsObjectReply, error)
	Delete(ctx context.Context, in *DeleteObjectRequest, opts ...grpc.CallOption) (*DeleteObjectReply, error)
	UpdateReferenceList(ctx context.Context, in *UpdateReferenceListRequest, opts ...grpc.CallOption) (*UpdateReferenceListReply, error)
	Check(ctx context.Context, in *CheckRequest, opts ...grpc.CallOption) (ObjectManager_CheckClient, error)
}

type objectManagerClient struct {
	cc *grpc.ClientConn
}

func NewObjectManagerClient(cc *grpc.ClientConn) ObjectManagerClient {
	return &objectManagerClient{cc}
}

func (c *objectManagerClient) Create(ctx context.Context, in *CreateObjectRequest, opts ...grpc.CallOption) (*CreateObjectReply, error) {
	out := new(CreateObjectReply)
	err := grpc.Invoke(ctx, "/ObjectManager/Create", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectManagerClient) List(ctx context.Context, in *ListObjectsRequest, opts ...grpc.CallOption) (ObjectManager_ListClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ObjectManager_serviceDesc.Streams[0], c.cc, "/ObjectManager/List", opts...)
	if err != nil {
		return nil, err
	}
	x := &objectManagerListClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ObjectManager_ListClient interface {
	Recv() (*Object, error)
	grpc.ClientStream
}

type objectManagerListClient struct {
	grpc.ClientStream
}

func (x *objectManagerListClient) Recv() (*Object, error) {
	m := new(Object)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *objectManagerClient) Get(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectReply, error) {
	out := new(GetObjectReply)
	err := grpc.Invoke(ctx, "/ObjectManager/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectManagerClient) Exists(ctx context.Context, in *ExistsObjectRequest, opts ...grpc.CallOption) (*ExistsObjectReply, error) {
	out := new(ExistsObjectReply)
	err := grpc.Invoke(ctx, "/ObjectManager/Exists", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectManagerClient) Delete(ctx context.Context, in *DeleteObjectRequest, opts ...grpc.CallOption) (*DeleteObjectReply, error) {
	out := new(DeleteObjectReply)
	err := grpc.Invoke(ctx, "/ObjectManager/Delete", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectManagerClient) UpdateReferenceList(ctx context.Context, in *UpdateReferenceListRequest, opts ...grpc.CallOption) (*UpdateReferenceListReply, error) {
	out := new(UpdateReferenceListReply)
	err := grpc.Invoke(ctx, "/ObjectManager/UpdateReferenceList", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectManagerClient) Check(ctx context.Context, in *CheckRequest, opts ...grpc.CallOption) (ObjectManager_CheckClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ObjectManager_serviceDesc.Streams[1], c.cc, "/ObjectManager/Check", opts...)
	if err != nil {
		return nil, err
	}
	x := &objectManagerCheckClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ObjectManager_CheckClient interface {
	Recv() (*CheckResponse, error)
	grpc.ClientStream
}

type objectManagerCheckClient struct {
	grpc.ClientStream
}

func (x *objectManagerCheckClient) Recv() (*CheckResponse, error) {
	m := new(CheckResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for ObjectManager service

type ObjectManagerServer interface {
	Create(context.Context, *CreateObjectRequest) (*CreateObjectReply, error)
	List(*ListObjectsRequest, ObjectManager_ListServer) error
	Get(context.Context, *GetObjectRequest) (*GetObjectReply, error)
	Exists(context.Context, *ExistsObjectRequest) (*ExistsObjectReply, error)
	Delete(context.Context, *DeleteObjectRequest) (*DeleteObjectReply, error)
	UpdateReferenceList(context.Context, *UpdateReferenceListRequest) (*UpdateReferenceListReply, error)
	Check(*CheckRequest, ObjectManager_CheckServer) error
}

func RegisterObjectManagerServer(s *grpc.Server, srv ObjectManagerServer) {
	s.RegisterService(&_ObjectManager_serviceDesc, srv)
}

func _ObjectManager_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectManagerServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ObjectManager/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectManagerServer).Create(ctx, req.(*CreateObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectManager_List_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListObjectsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObjectManagerServer).List(m, &objectManagerListServer{stream})
}

type ObjectManager_ListServer interface {
	Send(*Object) error
	grpc.ServerStream
}

type objectManagerListServer struct {
	grpc.ServerStream
}

func (x *objectManagerListServer) Send(m *Object) error {
	return x.ServerStream.SendMsg(m)
}

func _ObjectManager_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectManagerServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ObjectManager/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectManagerServer).Get(ctx, req.(*GetObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectManager_Exists_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExistsObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectManagerServer).Exists(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ObjectManager/Exists",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectManagerServer).Exists(ctx, req.(*ExistsObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectManager_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectManagerServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ObjectManager/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectManagerServer).Delete(ctx, req.(*DeleteObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectManager_UpdateReferenceList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateReferenceListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectManagerServer).UpdateReferenceList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ObjectManager/UpdateReferenceList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectManagerServer).UpdateReferenceList(ctx, req.(*UpdateReferenceListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectManager_Check_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CheckRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObjectManagerServer).Check(m, &objectManagerCheckServer{stream})
}

type ObjectManager_CheckServer interface {
	Send(*CheckResponse) error
	grpc.ServerStream
}

type objectManagerCheckServer struct {
	grpc.ServerStream
}

func (x *objectManagerCheckServer) Send(m *CheckResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _ObjectManager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ObjectManager",
	HandlerType: (*ObjectManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _ObjectManager_Create_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _ObjectManager_Get_Handler,
		},
		{
			MethodName: "Exists",
			Handler:    _ObjectManager_Exists_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _ObjectManager_Delete_Handler,
		},
		{
			MethodName: "UpdateReferenceList",
			Handler:    _ObjectManager_UpdateReferenceList_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "List",
			Handler:       _ObjectManager_List_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Check",
			Handler:       _ObjectManager_Check_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "store.proto",
}

func init() { proto.RegisterFile("store.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 651 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x55, 0x4d, 0x4f, 0xdb, 0x4a,
	0x14, 0x75, 0x6c, 0x62, 0x5e, 0x6e, 0x48, 0x5e, 0x72, 0x9d, 0xf7, 0xe4, 0xe7, 0x57, 0xa9, 0x68,
	0xd4, 0x45, 0x04, 0xed, 0x08, 0xd2, 0xaa, 0x8b, 0x4a, 0x2c, 0x2a, 0x88, 0x52, 0x54, 0x20, 0xc8,
	0x94, 0xee, 0x4d, 0x72, 0xdb, 0x1a, 0x4c, 0xe2, 0xda, 0x13, 0xda, 0x48, 0xdd, 0xf6, 0xa7, 0xf6,
	0x7f, 0x54, 0x33, 0x63, 0x42, 0x8c, 0xcd, 0x87, 0xba, 0xf3, 0xdc, 0x39, 0x67, 0xce, 0xdc, 0x33,
	0xe7, 0xca, 0x50, 0x4f, 0xc5, 0x34, 0x21, 0x1e, 0x27, 0x53, 0x31, 0x65, 0xab, 0x50, 0xed, 0x5f,
	0x8a, 0x78, 0xce, 0x7e, 0x55, 0xa0, 0x76, 0x14, 0x5c, 0x52, 0x1a, 0x07, 0x23, 0xc2, 0x0e, 0x54,
	0xa3, 0xe0, 0x8c, 0x22, 0xb7, 0xb2, 0x5e, 0xe9, 0xd6, 0x7c, 0xbd, 0xc0, 0x67, 0xd0, 0x50, 0xdb,
	0x6f, 0xaf, 0x82, 0x30, 0x0a, 0x22, 0x72, 0xcd, 0xf5, 0x4a, 0xd7, 0xf2, 0xf3, 0x45, 0x7c, 0x02,
	0x35, 0x55, 0x38, 0x4d, 0x69, 0xec, 0x5a, 0x0a, 0x71, 0x53, 0x40, 0x0e, 0x98, 0x50, 0x30, 0xf6,
	0xe9, 0xeb, 0x8c, 0x52, 0x71, 0x4c, 0xc9, 0xbb, 0xe9, 0x2c, 0x71, 0x57, 0x14, 0xac, 0x64, 0x07,
	0xb7, 0xc0, 0xf9, 0x96, 0x84, 0x82, 0x6e, 0x11, 0xaa, 0x8a, 0x50, 0xb6, 0x25, 0xf5, 0x27, 0xc9,
	0xf0, 0xec, 0x9c, 0x46, 0x22, 0x75, 0x6d, 0xad, 0xbf, 0x28, 0xb0, 0x8f, 0x60, 0xeb, 0x4f, 0x6c,
	0x81, 0x75, 0x41, 0x73, 0xd5, 0xe1, 0x9a, 0x2f, 0x3f, 0x65, 0xd7, 0x57, 0x41, 0x34, 0xd3, 0x7d,
	0xad, 0xf9, 0x7a, 0x21, 0xbb, 0x4e, 0xe8, 0x13, 0x25, 0x34, 0x19, 0xd1, 0x41, 0x98, 0x0a, 0xd7,
	0x5a, 0xb7, 0xba, 0x35, 0x3f, 0x5f, 0x64, 0x9b, 0xe0, 0x0c, 0x48, 0x2c, 0x1c, 0xcc, 0xee, 0x54,
	0x6e, 0x24, 0xdb, 0x81, 0x76, 0x1e, 0x1c, 0x47, 0x73, 0xec, 0x42, 0x6d, 0x72, 0x5d, 0x51, 0xf0,
	0x7a, 0x0f, 0xf8, 0x0d, 0xe6, 0x66, 0x93, 0x6d, 0x00, 0x4a, 0xcd, 0xac, 0xa5, 0xfb, 0xa5, 0x0e,
	0xc0, 0xd9, 0x4d, 0x28, 0x10, 0xa4, 0xd1, 0x4b, 0xe0, 0x83, 0x65, 0xb0, 0x5a, 0xe0, 0x53, 0xb0,
	0xa7, 0x0a, 0xa6, 0x1c, 0xa8, 0xf7, 0x56, 0x79, 0xc6, 0xca, 0xca, 0xcc, 0x81, 0x76, 0xfe, 0xb4,
	0x38, 0x9a, 0xb3, 0x1d, 0x70, 0xfa, 0xdf, 0xc3, 0x54, 0xa4, 0x8f, 0x91, 0xc8, 0x5c, 0x37, 0x17,
	0xae, 0xb3, 0x4d, 0x68, 0xe7, 0xe9, 0xd2, 0x8c, 0x7f, 0xc1, 0x26, 0x55, 0x54, 0xec, 0xbf, 0xfc,
	0x6c, 0xc5, 0xde, 0x40, 0x6b, 0x40, 0xe2, 0xcf, 0x84, 0xb6, 0xa1, 0xb9, 0xc4, 0x95, 0x2a, 0x0f,
	0xf6, 0xbb, 0x03, 0xce, 0x1e, 0x45, 0xf4, 0x38, 0xf7, 0x8a, 0x8a, 0x0e, 0xb4, 0xf3, 0x74, 0x69,
	0xd7, 0x39, 0x78, 0xa7, 0xf1, 0x38, 0x90, 0xb9, 0x5d, 0x0a, 0xd0, 0xbd, 0xaf, 0x58, 0x3c, 0xfa,
	0x91, 0xa9, 0xf4, 0xc0, 0x2d, 0xd5, 0x92, 0xf7, 0x78, 0x0d, 0x6b, 0xbb, 0x5f, 0x68, 0x74, 0xf1,
	0xa0, 0x72, 0x38, 0x4e, 0x5d, 0x53, 0x9d, 0x2e, 0x3f, 0xd9, 0x0f, 0x68, 0x64, 0xbc, 0x34, 0x9e,
	0x4e, 0x52, 0xc2, 0x26, 0x98, 0xe1, 0x38, 0x63, 0x99, 0xe1, 0x18, 0x5f, 0x80, 0x9d, 0x8a, 0x40,
	0xcc, 0x52, 0x75, 0xdf, 0x66, 0xef, 0x1f, 0x9e, 0xc3, 0xf3, 0x13, 0xb5, 0xe9, 0x67, 0x20, 0xf6,
	0x1c, 0x6c, 0x5d, 0x41, 0x1b, 0xcc, 0xe1, 0xfb, 0x96, 0x81, 0x0d, 0xa8, 0xed, 0x0e, 0x7d, 0xff,
	0xf4, 0xf8, 0x43, 0x7f, 0xaf, 0x55, 0xc1, 0x3a, 0xac, 0x1e, 0xee, 0x9f, 0x9c, 0xec, 0x1f, 0x0d,
	0x5a, 0x66, 0xaf, 0x0f, 0xad, 0xc5, 0x4c, 0x1c, 0x06, 0x93, 0xe0, 0x33, 0x25, 0xb8, 0x0d, 0xd6,
	0x80, 0x04, 0x76, 0x78, 0xc9, 0x04, 0x7a, 0xc8, 0x0b, 0xa3, 0xc6, 0x8c, 0xde, 0x4f, 0x0b, 0x1a,
	0xfa, 0x51, 0xae, 0x0f, 0x79, 0x05, 0xb6, 0x8e, 0x36, 0x76, 0x78, 0xc9, 0xc4, 0x78, 0xc8, 0x8b,
	0xc9, 0x37, 0xb0, 0x0b, 0x2b, 0xd2, 0x51, 0x74, 0x78, 0x71, 0x22, 0xbd, 0xeb, 0x38, 0x31, 0x63,
	0xab, 0x82, 0x9b, 0xfa, 0x92, 0x6d, 0x7e, 0x3b, 0xbf, 0xde, 0xdf, 0x3c, 0x1f, 0x4b, 0x66, 0xc8,
	0xcb, 0xe8, 0x99, 0xc0, 0x0e, 0x2f, 0x99, 0x2d, 0x0f, 0x79, 0x61, 0x64, 0x34, 0x4b, 0xc7, 0x0d,
	0x3b, 0xbc, 0x24, 0xb6, 0x1e, 0xf2, 0x62, 0x1a, 0x0d, 0x1c, 0x82, 0x53, 0x92, 0x11, 0xfc, 0x9f,
	0xdf, 0x9d, 0x52, 0xef, 0x3f, 0x7e, 0x67, 0xac, 0x0c, 0xdc, 0x80, 0xaa, 0x7a, 0x70, 0x6c, 0xf0,
	0xe5, 0x80, 0x79, 0xcd, 0x7c, 0x0e, 0xa4, 0x2b, 0x67, 0xb6, 0xfa, 0x0d, 0xbd, 0xfc, 0x1d, 0x00,
	0x00, 0xff, 0xff, 0xa0, 0x75, 0xc9, 0x12, 0x95, 0x06, 0x00, 0x00,
}
