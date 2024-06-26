// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: monaco.proto

package pb

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

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_monaco_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_monaco_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_monaco_proto_rawDescGZIP(), []int{0}
}

type ExecCodeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Language uint32 `protobuf:"varint,1,opt,name=language,proto3" json:"language,omitempty"`
	Code     string `protobuf:"bytes,2,opt,name=code,proto3" json:"code,omitempty"`
}

func (x *ExecCodeRequest) Reset() {
	*x = ExecCodeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_monaco_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecCodeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecCodeRequest) ProtoMessage() {}

func (x *ExecCodeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_monaco_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecCodeRequest.ProtoReflect.Descriptor instead.
func (*ExecCodeRequest) Descriptor() ([]byte, []int) {
	return file_monaco_proto_rawDescGZIP(), []int{1}
}

func (x *ExecCodeRequest) GetLanguage() uint32 {
	if x != nil {
		return x.Language
	}
	return 0
}

func (x *ExecCodeRequest) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

type ExecCodeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tip     string `protobuf:"bytes,1,opt,name=tip,proto3" json:"tip,omitempty"`
	Success bool   `protobuf:"varint,2,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *ExecCodeResponse) Reset() {
	*x = ExecCodeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_monaco_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecCodeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecCodeResponse) ProtoMessage() {}

func (x *ExecCodeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_monaco_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecCodeResponse.ProtoReflect.Descriptor instead.
func (*ExecCodeResponse) Descriptor() ([]byte, []int) {
	return file_monaco_proto_rawDescGZIP(), []int{2}
}

func (x *ExecCodeResponse) GetTip() string {
	if x != nil {
		return x.Tip
	}
	return ""
}

func (x *ExecCodeResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_monaco_proto protoreflect.FileDescriptor

var file_monaco_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x6d, 0x6f, 0x6e, 0x61, 0x63, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x6d, 0x6f, 0x6e, 0x61, 0x63, 0x6f, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x41, 0x0a, 0x0f, 0x45, 0x78, 0x65, 0x63, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x6f,
	0x64, 0x65, 0x22, 0x3e, 0x0a, 0x10, 0x45, 0x78, 0x65, 0x63, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x69, 0x70, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x69, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x32, 0x54, 0x0a, 0x13, 0x4d, 0x6f, 0x6e, 0x61, 0x63, 0x6f, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3d, 0x0a, 0x08, 0x45, 0x78, 0x65,
	0x63, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x17, 0x2e, 0x6d, 0x6f, 0x6e, 0x61, 0x63, 0x6f, 0x2e, 0x45,
	0x78, 0x65, 0x63, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18,
	0x2e, 0x6d, 0x6f, 0x6e, 0x61, 0x63, 0x6f, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x43, 0x6f, 0x64, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x05, 0x5a, 0x03, 0x2f, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_monaco_proto_rawDescOnce sync.Once
	file_monaco_proto_rawDescData = file_monaco_proto_rawDesc
)

func file_monaco_proto_rawDescGZIP() []byte {
	file_monaco_proto_rawDescOnce.Do(func() {
		file_monaco_proto_rawDescData = protoimpl.X.CompressGZIP(file_monaco_proto_rawDescData)
	})
	return file_monaco_proto_rawDescData
}

var file_monaco_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_monaco_proto_goTypes = []interface{}{
	(*Empty)(nil),            // 0: monaco.Empty
	(*ExecCodeRequest)(nil),  // 1: monaco.ExecCodeRequest
	(*ExecCodeResponse)(nil), // 2: monaco.ExecCodeResponse
}
var file_monaco_proto_depIdxs = []int32{
	1, // 0: monaco.MonacoServerService.ExecCode:input_type -> monaco.ExecCodeRequest
	2, // 1: monaco.MonacoServerService.ExecCode:output_type -> monaco.ExecCodeResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_monaco_proto_init() }
func file_monaco_proto_init() {
	if File_monaco_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_monaco_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
		file_monaco_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecCodeRequest); i {
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
		file_monaco_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecCodeResponse); i {
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
			RawDescriptor: file_monaco_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_monaco_proto_goTypes,
		DependencyIndexes: file_monaco_proto_depIdxs,
		MessageInfos:      file_monaco_proto_msgTypes,
	}.Build()
	File_monaco_proto = out.File
	file_monaco_proto_rawDesc = nil
	file_monaco_proto_goTypes = nil
	file_monaco_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// MonacoServerServiceClient is the client API for MonacoServerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MonacoServerServiceClient interface {
	ExecCode(ctx context.Context, in *ExecCodeRequest, opts ...grpc.CallOption) (*ExecCodeResponse, error)
}

type monacoServerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMonacoServerServiceClient(cc grpc.ClientConnInterface) MonacoServerServiceClient {
	return &monacoServerServiceClient{cc}
}

func (c *monacoServerServiceClient) ExecCode(ctx context.Context, in *ExecCodeRequest, opts ...grpc.CallOption) (*ExecCodeResponse, error) {
	out := new(ExecCodeResponse)
	err := c.cc.Invoke(ctx, "/monaco.MonacoServerService/ExecCode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MonacoServerServiceServer is the server API for MonacoServerService service.
type MonacoServerServiceServer interface {
	ExecCode(context.Context, *ExecCodeRequest) (*ExecCodeResponse, error)
}

// UnimplementedMonacoServerServiceServer can be embedded to have forward compatible implementations.
type UnimplementedMonacoServerServiceServer struct {
}

func (*UnimplementedMonacoServerServiceServer) ExecCode(context.Context, *ExecCodeRequest) (*ExecCodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecCode not implemented")
}

func RegisterMonacoServerServiceServer(s *grpc.Server, srv MonacoServerServiceServer) {
	s.RegisterService(&_MonacoServerService_serviceDesc, srv)
}

func _MonacoServerService_ExecCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecCodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonacoServerServiceServer).ExecCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monaco.MonacoServerService/ExecCode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonacoServerServiceServer).ExecCode(ctx, req.(*ExecCodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _MonacoServerService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "monaco.MonacoServerService",
	HandlerType: (*MonacoServerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ExecCode",
			Handler:    _MonacoServerService_ExecCode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "monaco.proto",
}
