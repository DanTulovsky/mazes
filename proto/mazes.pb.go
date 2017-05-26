// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mazes.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	mazes.proto

It has these top-level messages:
	ShowMazeRequest
	ShowMazeReply
	MazeConfig
	Cell
	MazeLocation
	HelloRequest
	HelloReply
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type ShowMazeRequest struct {
	Config *MazeConfig `protobuf:"bytes,1,opt,name=config" json:"config,omitempty"`
}

func (m *ShowMazeRequest) Reset()                    { *m = ShowMazeRequest{} }
func (m *ShowMazeRequest) String() string            { return proto1.CompactTextString(m) }
func (*ShowMazeRequest) ProtoMessage()               {}
func (*ShowMazeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ShowMazeRequest) GetConfig() *MazeConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

type ShowMazeReply struct {
}

func (m *ShowMazeReply) Reset()                    { *m = ShowMazeReply{} }
func (m *ShowMazeReply) String() string            { return proto1.CompactTextString(m) }
func (*ShowMazeReply) ProtoMessage()               {}
func (*ShowMazeReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// MazeConfig is the full config for a maze
type MazeConfig struct {
	Rows                 int64           `protobuf:"varint,1,opt,name=Rows" json:"Rows,omitempty"`
	Columns              int64           `protobuf:"varint,2,opt,name=Columns" json:"Columns,omitempty"`
	AllowWeaving         bool            `protobuf:"varint,3,opt,name=AllowWeaving" json:"AllowWeaving,omitempty"`
	WeavingProbability   float64         `protobuf:"fixed64,4,opt,name=WeavingProbability" json:"WeavingProbability,omitempty"`
	CellWidth            int64           `protobuf:"varint,5,opt,name=CellWidth" json:"CellWidth,omitempty"`
	WallWidth            int64           `protobuf:"varint,6,opt,name=WallWidth" json:"WallWidth,omitempty"`
	WallSpace            int64           `protobuf:"varint,7,opt,name=WallSpace" json:"WallSpace,omitempty"`
	PathWidth            int64           `protobuf:"varint,8,opt,name=PathWidth" json:"PathWidth,omitempty"`
	MarkVisitedCells     bool            `protobuf:"varint,9,opt,name=MarkVisitedCells" json:"MarkVisitedCells,omitempty"`
	ShowDistanceValues   bool            `protobuf:"varint,10,opt,name=ShowDistanceValues" json:"ShowDistanceValues,omitempty"`
	ShowDistanceColors   bool            `protobuf:"varint,11,opt,name=ShowDistanceColors" json:"ShowDistanceColors,omitempty"`
	SkipGridCheck        bool            `protobuf:"varint,12,opt,name=SkipGridCheck" json:"SkipGridCheck,omitempty"`
	OrphanMask           []*MazeLocation `protobuf:"bytes,13,rep,name=OrphanMask" json:"OrphanMask,omitempty"`
	AvatarImage          string          `protobuf:"bytes,14,opt,name=AvatarImage" json:"AvatarImage,omitempty"`
	VisitedCellColor     string          `protobuf:"bytes,15,opt,name=VisitedCellColor" json:"VisitedCellColor,omitempty"`
	BgColor              string          `protobuf:"bytes,16,opt,name=BgColor" json:"BgColor,omitempty"`
	BorderColor          string          `protobuf:"bytes,17,opt,name=BorderColor" json:"BorderColor,omitempty"`
	WallColor            string          `protobuf:"bytes,18,opt,name=WallColor" json:"WallColor,omitempty"`
	PathColor            string          `protobuf:"bytes,19,opt,name=PathColor" json:"PathColor,omitempty"`
	CurrentLocationColor string          `protobuf:"bytes,20,opt,name=CurrentLocationColor" json:"CurrentLocationColor,omitempty"`
	FromCellColor        string          `protobuf:"bytes,21,opt,name=FromCellColor" json:"FromCellColor,omitempty"`
	ToCellColor          string          `protobuf:"bytes,22,opt,name=ToCellColor" json:"ToCellColor,omitempty"`
	FromCell             string          `protobuf:"bytes,23,opt,name=FromCell" json:"FromCell,omitempty"`
	ToCell               string          `protobuf:"bytes,24,opt,name=ToCell" json:"ToCell,omitempty"`
}

func (m *MazeConfig) Reset()                    { *m = MazeConfig{} }
func (m *MazeConfig) String() string            { return proto1.CompactTextString(m) }
func (*MazeConfig) ProtoMessage()               {}
func (*MazeConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *MazeConfig) GetRows() int64 {
	if m != nil {
		return m.Rows
	}
	return 0
}

func (m *MazeConfig) GetColumns() int64 {
	if m != nil {
		return m.Columns
	}
	return 0
}

func (m *MazeConfig) GetAllowWeaving() bool {
	if m != nil {
		return m.AllowWeaving
	}
	return false
}

func (m *MazeConfig) GetWeavingProbability() float64 {
	if m != nil {
		return m.WeavingProbability
	}
	return 0
}

func (m *MazeConfig) GetCellWidth() int64 {
	if m != nil {
		return m.CellWidth
	}
	return 0
}

func (m *MazeConfig) GetWallWidth() int64 {
	if m != nil {
		return m.WallWidth
	}
	return 0
}

func (m *MazeConfig) GetWallSpace() int64 {
	if m != nil {
		return m.WallSpace
	}
	return 0
}

func (m *MazeConfig) GetPathWidth() int64 {
	if m != nil {
		return m.PathWidth
	}
	return 0
}

func (m *MazeConfig) GetMarkVisitedCells() bool {
	if m != nil {
		return m.MarkVisitedCells
	}
	return false
}

func (m *MazeConfig) GetShowDistanceValues() bool {
	if m != nil {
		return m.ShowDistanceValues
	}
	return false
}

func (m *MazeConfig) GetShowDistanceColors() bool {
	if m != nil {
		return m.ShowDistanceColors
	}
	return false
}

func (m *MazeConfig) GetSkipGridCheck() bool {
	if m != nil {
		return m.SkipGridCheck
	}
	return false
}

func (m *MazeConfig) GetOrphanMask() []*MazeLocation {
	if m != nil {
		return m.OrphanMask
	}
	return nil
}

func (m *MazeConfig) GetAvatarImage() string {
	if m != nil {
		return m.AvatarImage
	}
	return ""
}

func (m *MazeConfig) GetVisitedCellColor() string {
	if m != nil {
		return m.VisitedCellColor
	}
	return ""
}

func (m *MazeConfig) GetBgColor() string {
	if m != nil {
		return m.BgColor
	}
	return ""
}

func (m *MazeConfig) GetBorderColor() string {
	if m != nil {
		return m.BorderColor
	}
	return ""
}

func (m *MazeConfig) GetWallColor() string {
	if m != nil {
		return m.WallColor
	}
	return ""
}

func (m *MazeConfig) GetPathColor() string {
	if m != nil {
		return m.PathColor
	}
	return ""
}

func (m *MazeConfig) GetCurrentLocationColor() string {
	if m != nil {
		return m.CurrentLocationColor
	}
	return ""
}

func (m *MazeConfig) GetFromCellColor() string {
	if m != nil {
		return m.FromCellColor
	}
	return ""
}

func (m *MazeConfig) GetToCellColor() string {
	if m != nil {
		return m.ToCellColor
	}
	return ""
}

func (m *MazeConfig) GetFromCell() string {
	if m != nil {
		return m.FromCell
	}
	return ""
}

func (m *MazeConfig) GetToCell() string {
	if m != nil {
		return m.ToCell
	}
	return ""
}

type Cell struct {
}

func (m *Cell) Reset()                    { *m = Cell{} }
func (m *Cell) String() string            { return proto1.CompactTextString(m) }
func (*Cell) ProtoMessage()               {}
func (*Cell) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// MazeLocation is a location in the maze
type MazeLocation struct {
	X int64 `protobuf:"varint,1,opt,name=X" json:"X,omitempty"`
	Y int64 `protobuf:"varint,2,opt,name=Y" json:"Y,omitempty"`
	Z int64 `protobuf:"varint,3,opt,name=Z" json:"Z,omitempty"`
}

func (m *MazeLocation) Reset()                    { *m = MazeLocation{} }
func (m *MazeLocation) String() string            { return proto1.CompactTextString(m) }
func (*MazeLocation) ProtoMessage()               {}
func (*MazeLocation) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *MazeLocation) GetX() int64 {
	if m != nil {
		return m.X
	}
	return 0
}

func (m *MazeLocation) GetY() int64 {
	if m != nil {
		return m.Y
	}
	return 0
}

func (m *MazeLocation) GetZ() int64 {
	if m != nil {
		return m.Z
	}
	return 0
}

// The request message containing the user's name.
type HelloRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *HelloRequest) Reset()                    { *m = HelloRequest{} }
func (m *HelloRequest) String() string            { return proto1.CompactTextString(m) }
func (*HelloRequest) ProtoMessage()               {}
func (*HelloRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *HelloRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

// The response message containing the greetings
type HelloReply struct {
	Message string `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
}

func (m *HelloReply) Reset()                    { *m = HelloReply{} }
func (m *HelloReply) String() string            { return proto1.CompactTextString(m) }
func (*HelloReply) ProtoMessage()               {}
func (*HelloReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *HelloReply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto1.RegisterType((*ShowMazeRequest)(nil), "proto.ShowMazeRequest")
	proto1.RegisterType((*ShowMazeReply)(nil), "proto.ShowMazeReply")
	proto1.RegisterType((*MazeConfig)(nil), "proto.MazeConfig")
	proto1.RegisterType((*Cell)(nil), "proto.Cell")
	proto1.RegisterType((*MazeLocation)(nil), "proto.MazeLocation")
	proto1.RegisterType((*HelloRequest)(nil), "proto.HelloRequest")
	proto1.RegisterType((*HelloReply)(nil), "proto.HelloReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Mazer service

type MazerClient interface {
	// Sends a greeting
	SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error)
	// Display maze
	ShowMaze(ctx context.Context, in *ShowMazeRequest, opts ...grpc.CallOption) (*ShowMazeReply, error)
}

type mazerClient struct {
	cc *grpc.ClientConn
}

func NewMazerClient(cc *grpc.ClientConn) MazerClient {
	return &mazerClient{cc}
}

func (c *mazerClient) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error) {
	out := new(HelloReply)
	err := grpc.Invoke(ctx, "/proto.Mazer/SayHello", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mazerClient) ShowMaze(ctx context.Context, in *ShowMazeRequest, opts ...grpc.CallOption) (*ShowMazeReply, error) {
	out := new(ShowMazeReply)
	err := grpc.Invoke(ctx, "/proto.Mazer/ShowMaze", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Mazer service

type MazerServer interface {
	// Sends a greeting
	SayHello(context.Context, *HelloRequest) (*HelloReply, error)
	// Display maze
	ShowMaze(context.Context, *ShowMazeRequest) (*ShowMazeReply, error)
}

func RegisterMazerServer(s *grpc.Server, srv MazerServer) {
	s.RegisterService(&_Mazer_serviceDesc, srv)
}

func _Mazer_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MazerServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Mazer/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MazerServer).SayHello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Mazer_ShowMaze_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShowMazeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MazerServer).ShowMaze(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Mazer/ShowMaze",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MazerServer).ShowMaze(ctx, req.(*ShowMazeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Mazer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Mazer",
	HandlerType: (*MazerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _Mazer_SayHello_Handler,
		},
		{
			MethodName: "ShowMaze",
			Handler:    _Mazer_ShowMaze_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mazes.proto",
}

func init() { proto1.RegisterFile("mazes.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 584 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x94, 0xcf, 0x6f, 0xd3, 0x30,
	0x14, 0xc7, 0x09, 0xed, 0xb2, 0xf6, 0xb5, 0x63, 0x9b, 0x37, 0x86, 0x35, 0x71, 0x88, 0x22, 0x84,
	0x0a, 0x87, 0x1d, 0x3a, 0x0e, 0x08, 0x71, 0xd9, 0x82, 0xf8, 0x21, 0x31, 0x31, 0xa5, 0x68, 0xbf,
	0x6e, 0x5e, 0x6b, 0xda, 0xa8, 0x4e, 0x1c, 0xec, 0x74, 0x55, 0xf7, 0xef, 0xf2, 0x8f, 0x20, 0x3f,
	0xc7, 0x4d, 0xba, 0xf6, 0x54, 0xbf, 0xef, 0xf7, 0xf3, 0xfa, 0xfc, 0xfc, 0x62, 0x43, 0x27, 0x65,
	0x8f, 0x5c, 0x9f, 0xe4, 0x4a, 0x16, 0x92, 0x6c, 0xe1, 0x4f, 0xf8, 0x19, 0x76, 0x07, 0x13, 0x39,
	0xbf, 0x60, 0x8f, 0x3c, 0xe6, 0x7f, 0x67, 0x5c, 0x17, 0xe4, 0x1d, 0xf8, 0x43, 0x99, 0xfd, 0x49,
	0xc6, 0xd4, 0x0b, 0xbc, 0x5e, 0xa7, 0xbf, 0x6f, 0x33, 0x4e, 0x0c, 0x13, 0xa1, 0x11, 0x97, 0x40,
	0xb8, 0x0b, 0x3b, 0x55, 0x76, 0x2e, 0x16, 0xe1, 0x3f, 0x1f, 0xa0, 0xe2, 0x08, 0x81, 0x66, 0x2c,
	0xe7, 0x1a, 0xff, 0xa8, 0x11, 0xe3, 0x9a, 0x50, 0xd8, 0x8e, 0xa4, 0x98, 0xa5, 0x99, 0xa6, 0xcf,
	0x51, 0x76, 0x21, 0x09, 0xa1, 0x7b, 0x26, 0x84, 0x9c, 0x5f, 0x73, 0xf6, 0x90, 0x64, 0x63, 0xda,
	0x08, 0xbc, 0x5e, 0x2b, 0x5e, 0xd1, 0xc8, 0x09, 0x90, 0x72, 0x79, 0xa9, 0xe4, 0x3d, 0xbb, 0x4f,
	0x44, 0x52, 0x2c, 0x68, 0x33, 0xf0, 0x7a, 0x5e, 0xbc, 0xc1, 0x21, 0xaf, 0xa1, 0x1d, 0x71, 0x21,
	0xae, 0x93, 0x51, 0x31, 0xa1, 0x5b, 0x58, 0xaf, 0x12, 0x8c, 0x7b, 0xcd, 0x9c, 0xeb, 0x5b, 0x77,
	0x29, 0x38, 0x77, 0x90, 0xb3, 0x21, 0xa7, 0xdb, 0x95, 0x8b, 0x82, 0x71, 0x2f, 0x59, 0x31, 0xb1,
	0xb9, 0x2d, 0xeb, 0x2e, 0x05, 0xf2, 0x1e, 0xf6, 0x2e, 0x98, 0x9a, 0x5e, 0x25, 0x3a, 0x29, 0xf8,
	0xc8, 0x54, 0xd4, 0xb4, 0x8d, 0xfd, 0xac, 0xe9, 0xa6, 0x27, 0x73, 0x8a, 0x5f, 0x12, 0x5d, 0xb0,
	0x6c, 0xc8, 0xaf, 0x98, 0x98, 0x71, 0x4d, 0x01, 0xe9, 0x0d, 0xce, 0x53, 0x3e, 0x92, 0x42, 0x2a,
	0x4d, 0x3b, 0xeb, 0xbc, 0x75, 0xc8, 0x1b, 0xd8, 0x19, 0x4c, 0x93, 0xfc, 0x9b, 0x4a, 0x46, 0xd1,
	0x84, 0x0f, 0xa7, 0xb4, 0x8b, 0xe8, 0xaa, 0x48, 0x4e, 0x01, 0x7e, 0xa9, 0x7c, 0xc2, 0xb2, 0x0b,
	0xa6, 0xa7, 0x74, 0x27, 0x68, 0xf4, 0x3a, 0xfd, 0x83, 0xda, 0xe8, 0x7f, 0xca, 0x21, 0x2b, 0x12,
	0x99, 0xc5, 0x35, 0x8c, 0x04, 0xd0, 0x39, 0x7b, 0x60, 0x05, 0x53, 0x3f, 0x52, 0x36, 0xe6, 0xf4,
	0x45, 0xe0, 0xf5, 0xda, 0x71, 0x5d, 0x32, 0x07, 0x51, 0x6b, 0x16, 0x77, 0x44, 0x77, 0x11, 0x5b,
	0xd3, 0xcd, 0xa7, 0x71, 0x3e, 0xb6, 0xc8, 0x1e, 0x22, 0x2e, 0x34, 0x75, 0xce, 0xa5, 0x1a, 0x71,
	0x65, 0xdd, 0x7d, 0x5b, 0xa7, 0x26, 0xb9, 0x61, 0x59, 0x9f, 0xa0, 0x5f, 0x09, 0x6e, 0x58, 0xd6,
	0x3d, 0xb0, 0xee, 0x52, 0x20, 0x7d, 0x38, 0x8c, 0x66, 0x4a, 0xf1, 0xac, 0x70, 0x4d, 0x5a, 0xf0,
	0x10, 0xc1, 0x8d, 0x9e, 0x39, 0xd4, 0xaf, 0x4a, 0xa6, 0x55, 0x53, 0x2f, 0x11, 0x5e, 0x15, 0xcd,
	0xbe, 0x7f, 0xcb, 0x8a, 0x39, 0xb2, 0xfb, 0xae, 0x49, 0xe4, 0x18, 0x5a, 0x2e, 0x85, 0xbe, 0x42,
	0x7b, 0x19, 0x93, 0x23, 0xf0, 0x2d, 0x4a, 0x29, 0x3a, 0x65, 0x14, 0xfa, 0xd0, 0xc4, 0xdf, 0x8f,
	0xd0, 0xad, 0x4f, 0x86, 0x74, 0xc1, 0xbb, 0x29, 0xef, 0x9a, 0x77, 0x63, 0xa2, 0xdb, 0xf2, 0x8a,
	0x79, 0xb7, 0x26, 0xba, 0xc3, 0x1b, 0xd5, 0x88, 0xbd, 0xbb, 0x30, 0x84, 0xee, 0x77, 0x2e, 0x84,
	0x74, 0x77, 0x9e, 0x40, 0x33, 0x63, 0x29, 0xc7, 0xe4, 0x76, 0x8c, 0xeb, 0xf0, 0x2d, 0x40, 0xc9,
	0xe4, 0x62, 0x61, 0x66, 0x93, 0x72, 0xad, 0xcd, 0x94, 0x2d, 0xe4, 0xc2, 0xfe, 0x02, 0xb6, 0xcc,
	0x2e, 0x14, 0xf9, 0x00, 0xad, 0x01, 0x5b, 0x60, 0x0e, 0x71, 0x5f, 0x4e, 0xbd, 0xca, 0xf1, 0xfe,
	0xaa, 0x68, 0x1e, 0x8c, 0x67, 0xe4, 0x13, 0xb4, 0xdc, 0x1b, 0x42, 0x8e, 0x4a, 0xe0, 0xc9, 0x93,
	0x74, 0x7c, 0xb8, 0xa6, 0x63, 0xee, 0xbd, 0x8f, 0xf2, 0xe9, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x77, 0x77, 0x26, 0x90, 0xda, 0x04, 0x00, 0x00,
}
