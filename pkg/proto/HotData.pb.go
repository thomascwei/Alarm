// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: proto/HotData.proto

// A package is a unique name, so that differing protocol buffers don’t set on each other.
// The names aren’t tied to Go packages, but Go uses them.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// The messages passed by the RPC Hello service. They each have one string property
type HotDataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ObjectID string `protobuf:"bytes,1,opt,name=ObjectID,proto3" json:"ObjectID,omitempty"`
	Value    string `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
}

func (x *HotDataRequest) Reset() {
	*x = HotDataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_HotData_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HotDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HotDataRequest) ProtoMessage() {}

func (x *HotDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_HotData_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HotDataRequest.ProtoReflect.Descriptor instead.
func (*HotDataRequest) Descriptor() ([]byte, []int) {
	return file_proto_HotData_proto_rawDescGZIP(), []int{0}
}

func (x *HotDataRequest) GetObjectID() string {
	if x != nil {
		return x.ObjectID
	}
	return ""
}

func (x *HotDataRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type HotDataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ObjectID string `protobuf:"bytes,1,opt,name=ObjectID,proto3" json:"ObjectID,omitempty"`
	Value    string `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
	Message  string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *HotDataResponse) Reset() {
	*x = HotDataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_HotData_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HotDataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HotDataResponse) ProtoMessage() {}

func (x *HotDataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_HotData_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HotDataResponse.ProtoReflect.Descriptor instead.
func (*HotDataResponse) Descriptor() ([]byte, []int) {
	return file_proto_HotData_proto_rawDescGZIP(), []int{1}
}

func (x *HotDataResponse) GetObjectID() string {
	if x != nil {
		return x.ObjectID
	}
	return ""
}

func (x *HotDataResponse) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *HotDataResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_proto_HotData_proto protoreflect.FileDescriptor

var file_proto_HotData_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x48, 0x6f, 0x74, 0x44, 0x61, 0x74, 0x61, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x6d, 0x61, 0x69, 0x6e, 0x22, 0x42, 0x0a, 0x0e, 0x48,
	0x6f, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22,
	0x5d, 0x0a, 0x0f, 0x48, 0x6f, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x44, 0x12, 0x14,
	0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0x4a,
	0x0a, 0x0f, 0x48, 0x6f, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65,
	0x72, 0x12, 0x37, 0x0a, 0x06, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x12, 0x14, 0x2e, 0x6d, 0x61,
	0x69, 0x6e, 0x2e, 0x48, 0x6f, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x15, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x48, 0x6f, 0x74, 0x44, 0x61, 0x74, 0x61,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x08, 0x5a, 0x06, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_HotData_proto_rawDescOnce sync.Once
	file_proto_HotData_proto_rawDescData = file_proto_HotData_proto_rawDesc
)

func file_proto_HotData_proto_rawDescGZIP() []byte {
	file_proto_HotData_proto_rawDescOnce.Do(func() {
		file_proto_HotData_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_HotData_proto_rawDescData)
	})
	return file_proto_HotData_proto_rawDescData
}

var file_proto_HotData_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_HotData_proto_goTypes = []interface{}{
	(*HotDataRequest)(nil),  // 0: main.HotDataRequest
	(*HotDataResponse)(nil), // 1: main.HotDataResponse
}
var file_proto_HotData_proto_depIdxs = []int32{
	0, // 0: main.HotDataReceiver.Insert:input_type -> main.HotDataRequest
	1, // 1: main.HotDataReceiver.Insert:output_type -> main.HotDataResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_HotData_proto_init() }
func file_proto_HotData_proto_init() {
	if File_proto_HotData_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_HotData_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HotDataRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_HotData_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HotDataResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_HotData_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_HotData_proto_goTypes,
		DependencyIndexes: file_proto_HotData_proto_depIdxs,
		MessageInfos:      file_proto_HotData_proto_msgTypes,
	}.Build()
	File_proto_HotData_proto = out.File
	file_proto_HotData_proto_rawDesc = nil
	file_proto_HotData_proto_goTypes = nil
	file_proto_HotData_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// HotDataReceiverClient is the client API for HotDataReceiver service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type HotDataReceiverClient interface {
	// An RPC service call named Say that receives HelloRequest and returns HelloResponse
	Insert(ctx context.Context, in *HotDataRequest, opts ...grpc.CallOption) (*HotDataResponse, error)
}

type hotDataReceiverClient struct {
	cc grpc.ClientConnInterface
}

func NewHotDataReceiverClient(cc grpc.ClientConnInterface) HotDataReceiverClient {
	return &hotDataReceiverClient{cc}
}

func (c *hotDataReceiverClient) Insert(ctx context.Context, in *HotDataRequest, opts ...grpc.CallOption) (*HotDataResponse, error) {
	out := new(HotDataResponse)
	err := c.cc.Invoke(ctx, "/main.HotDataReceiver/Insert", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HotDataReceiverServer is the server API for HotDataReceiver service.
type HotDataReceiverServer interface {
	// An RPC service call named Say that receives HelloRequest and returns HelloResponse
	Insert(context.Context, *HotDataRequest) (*HotDataResponse, error)
}

// UnimplementedHotDataReceiverServer can be embedded to have forward compatible implementations.
type UnimplementedHotDataReceiverServer struct {
}

func (*UnimplementedHotDataReceiverServer) Insert(context.Context, *HotDataRequest) (*HotDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Insert not implemented")
}

func RegisterHotDataReceiverServer(s *grpc.Server, srv HotDataReceiverServer) {
	s.RegisterService(&_HotDataReceiver_serviceDesc, srv)
}

func _HotDataReceiver_Insert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HotDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HotDataReceiverServer).Insert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.HotDataReceiver/Insert",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HotDataReceiverServer).Insert(ctx, req.(*HotDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _HotDataReceiver_serviceDesc = grpc.ServiceDesc{
	ServiceName: "main.HotDataReceiver",
	HandlerType: (*HotDataReceiverServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Insert",
			Handler:    _HotDataReceiver_Insert_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/HotData.proto",
}