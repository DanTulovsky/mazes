// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mazes.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	mazes.proto

It has these top-level messages:
	SolveMazeRequest
	SolveMazeResponse
	Direction
	Maze
	ListMazeRequest
	ListMazeReply
	CreateMazeRequest
	CreateMazeReply
	MazeConfig
	Cell
	MazeLocation
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

// SolveMazeRequest is a message sent from the client trying to solve a maze
type SolveMazeRequest struct {
	MazeId    string `protobuf:"bytes,1,opt,name=mazeId" json:"mazeId,omitempty"`
	ClientId  string `protobuf:"bytes,2,opt,name=clientId" json:"clientId,omitempty"`
	Direction string `protobuf:"bytes,3,opt,name=direction" json:"direction,omitempty"`
	// on first connect, this must be set true, the direction field is ignored, the client does not move
	Initial  bool `protobuf:"varint,4,opt,name=initial" json:"initial,omitempty"`
	MoveBack bool `protobuf:"varint,5,opt,name=move_back,json=moveBack" json:"move_back,omitempty"`
}

func (m *SolveMazeRequest) Reset()                    { *m = SolveMazeRequest{} }
func (m *SolveMazeRequest) String() string            { return proto1.CompactTextString(m) }
func (*SolveMazeRequest) ProtoMessage()               {}
func (*SolveMazeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *SolveMazeRequest) GetMazeId() string {
	if m != nil {
		return m.MazeId
	}
	return ""
}

func (m *SolveMazeRequest) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

func (m *SolveMazeRequest) GetDirection() string {
	if m != nil {
		return m.Direction
	}
	return ""
}

func (m *SolveMazeRequest) GetInitial() bool {
	if m != nil {
		return m.Initial
	}
	return false
}

func (m *SolveMazeRequest) GetMoveBack() bool {
	if m != nil {
		return m.MoveBack
	}
	return false
}

// SolveMazeResponse is a response sent from the server as the client tries to solve a maze
type SolveMazeResponse struct {
	MazeId              string        `protobuf:"bytes,1,opt,name=mazeId" json:"mazeId,omitempty"`
	ClientId            string        `protobuf:"bytes,2,opt,name=clientId" json:"clientId,omitempty"`
	AvailableDirections []*Direction  `protobuf:"bytes,3,rep,name=available_directions,json=availableDirections" json:"available_directions,omitempty"`
	Initial             bool          `protobuf:"varint,4,opt,name=initial" json:"initial,omitempty"`
	Error               bool          `protobuf:"varint,5,opt,name=error" json:"error,omitempty"`
	ErrorMessage        string        `protobuf:"bytes,6,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
	CurrentLocation     *MazeLocation `protobuf:"bytes,7,opt,name=current_location,json=currentLocation" json:"current_location,omitempty"`
	FromCell            *MazeLocation `protobuf:"bytes,8,opt,name=from_cell,json=fromCell" json:"from_cell,omitempty"`
	ToCell              *MazeLocation `protobuf:"bytes,9,opt,name=to_cell,json=toCell" json:"to_cell,omitempty"`
	Solved              bool          `protobuf:"varint,10,opt,name=solved" json:"solved,omitempty"`
}

func (m *SolveMazeResponse) Reset()                    { *m = SolveMazeResponse{} }
func (m *SolveMazeResponse) String() string            { return proto1.CompactTextString(m) }
func (*SolveMazeResponse) ProtoMessage()               {}
func (*SolveMazeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *SolveMazeResponse) GetMazeId() string {
	if m != nil {
		return m.MazeId
	}
	return ""
}

func (m *SolveMazeResponse) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

func (m *SolveMazeResponse) GetAvailableDirections() []*Direction {
	if m != nil {
		return m.AvailableDirections
	}
	return nil
}

func (m *SolveMazeResponse) GetInitial() bool {
	if m != nil {
		return m.Initial
	}
	return false
}

func (m *SolveMazeResponse) GetError() bool {
	if m != nil {
		return m.Error
	}
	return false
}

func (m *SolveMazeResponse) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	return ""
}

func (m *SolveMazeResponse) GetCurrentLocation() *MazeLocation {
	if m != nil {
		return m.CurrentLocation
	}
	return nil
}

func (m *SolveMazeResponse) GetFromCell() *MazeLocation {
	if m != nil {
		return m.FromCell
	}
	return nil
}

func (m *SolveMazeResponse) GetToCell() *MazeLocation {
	if m != nil {
		return m.ToCell
	}
	return nil
}

func (m *SolveMazeResponse) GetSolved() bool {
	if m != nil {
		return m.Solved
	}
	return false
}

type Direction struct {
	Name    string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Visited bool   `protobuf:"varint,2,opt,name=visited" json:"visited,omitempty"`
}

func (m *Direction) Reset()                    { *m = Direction{} }
func (m *Direction) String() string            { return proto1.CompactTextString(m) }
func (*Direction) ProtoMessage()               {}
func (*Direction) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Direction) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Direction) GetVisited() bool {
	if m != nil {
		return m.Visited
	}
	return false
}

// Maze defines a maze and its clients
type Maze struct {
	MazeId    string   `protobuf:"bytes,1,opt,name=mazeId" json:"mazeId,omitempty"`
	ClientIds []string `protobuf:"bytes,2,rep,name=clientIds" json:"clientIds,omitempty"`
}

func (m *Maze) Reset()                    { *m = Maze{} }
func (m *Maze) String() string            { return proto1.CompactTextString(m) }
func (*Maze) ProtoMessage()               {}
func (*Maze) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Maze) GetMazeId() string {
	if m != nil {
		return m.MazeId
	}
	return ""
}

func (m *Maze) GetClientIds() []string {
	if m != nil {
		return m.ClientIds
	}
	return nil
}

type ListMazeRequest struct {
}

func (m *ListMazeRequest) Reset()                    { *m = ListMazeRequest{} }
func (m *ListMazeRequest) String() string            { return proto1.CompactTextString(m) }
func (*ListMazeRequest) ProtoMessage()               {}
func (*ListMazeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type ListMazeReply struct {
	Mazes []*Maze `protobuf:"bytes,1,rep,name=mazes" json:"mazes,omitempty"`
}

func (m *ListMazeReply) Reset()                    { *m = ListMazeReply{} }
func (m *ListMazeReply) String() string            { return proto1.CompactTextString(m) }
func (*ListMazeReply) ProtoMessage()               {}
func (*ListMazeReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ListMazeReply) GetMazes() []*Maze {
	if m != nil {
		return m.Mazes
	}
	return nil
}

type CreateMazeRequest struct {
	Config *MazeConfig `protobuf:"bytes,1,opt,name=config" json:"config,omitempty"`
}

func (m *CreateMazeRequest) Reset()                    { *m = CreateMazeRequest{} }
func (m *CreateMazeRequest) String() string            { return proto1.CompactTextString(m) }
func (*CreateMazeRequest) ProtoMessage()               {}
func (*CreateMazeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *CreateMazeRequest) GetConfig() *MazeConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

type CreateMazeReply struct {
	MazeId   string `protobuf:"bytes,1,opt,name=MazeId" json:"MazeId,omitempty"`
	ClientId string `protobuf:"bytes,2,opt,name=ClientId" json:"ClientId,omitempty"`
}

func (m *CreateMazeReply) Reset()                    { *m = CreateMazeReply{} }
func (m *CreateMazeReply) String() string            { return proto1.CompactTextString(m) }
func (*CreateMazeReply) ProtoMessage()               {}
func (*CreateMazeReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *CreateMazeReply) GetMazeId() string {
	if m != nil {
		return m.MazeId
	}
	return ""
}

func (m *CreateMazeReply) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

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
	GenDrawDelay         string          `protobuf:"bytes,25,opt,name=GenDrawDelay" json:"GenDrawDelay,omitempty"`
	CreateAlgo           string          `protobuf:"bytes,26,opt,name=CreateAlgo" json:"CreateAlgo,omitempty"`
	ShowFromToColors     bool            `protobuf:"varint,27,opt,name=ShowFromToColors" json:"ShowFromToColors,omitempty"`
	Id                   string          `protobuf:"bytes,28,opt,name=Id" json:"Id,omitempty"`
}

func (m *MazeConfig) Reset()                    { *m = MazeConfig{} }
func (m *MazeConfig) String() string            { return proto1.CompactTextString(m) }
func (*MazeConfig) ProtoMessage()               {}
func (*MazeConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

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

func (m *MazeConfig) GetGenDrawDelay() string {
	if m != nil {
		return m.GenDrawDelay
	}
	return ""
}

func (m *MazeConfig) GetCreateAlgo() string {
	if m != nil {
		return m.CreateAlgo
	}
	return ""
}

func (m *MazeConfig) GetShowFromToColors() bool {
	if m != nil {
		return m.ShowFromToColors
	}
	return false
}

func (m *MazeConfig) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type Cell struct {
}

func (m *Cell) Reset()                    { *m = Cell{} }
func (m *Cell) String() string            { return proto1.CompactTextString(m) }
func (*Cell) ProtoMessage()               {}
func (*Cell) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

// MazeLocation is a location in the maze
type MazeLocation struct {
	X int64 `protobuf:"varint,1,opt,name=X" json:"X,omitempty"`
	Y int64 `protobuf:"varint,2,opt,name=Y" json:"Y,omitempty"`
	Z int64 `protobuf:"varint,3,opt,name=Z" json:"Z,omitempty"`
}

func (m *MazeLocation) Reset()                    { *m = MazeLocation{} }
func (m *MazeLocation) String() string            { return proto1.CompactTextString(m) }
func (*MazeLocation) ProtoMessage()               {}
func (*MazeLocation) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

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

func init() {
	proto1.RegisterType((*SolveMazeRequest)(nil), "proto.SolveMazeRequest")
	proto1.RegisterType((*SolveMazeResponse)(nil), "proto.SolveMazeResponse")
	proto1.RegisterType((*Direction)(nil), "proto.Direction")
	proto1.RegisterType((*Maze)(nil), "proto.Maze")
	proto1.RegisterType((*ListMazeRequest)(nil), "proto.ListMazeRequest")
	proto1.RegisterType((*ListMazeReply)(nil), "proto.ListMazeReply")
	proto1.RegisterType((*CreateMazeRequest)(nil), "proto.CreateMazeRequest")
	proto1.RegisterType((*CreateMazeReply)(nil), "proto.CreateMazeReply")
	proto1.RegisterType((*MazeConfig)(nil), "proto.MazeConfig")
	proto1.RegisterType((*Cell)(nil), "proto.Cell")
	proto1.RegisterType((*MazeLocation)(nil), "proto.MazeLocation")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Mazer service

type MazerClient interface {
	// Display maze
	CreateMaze(ctx context.Context, in *CreateMazeRequest, opts ...grpc.CallOption) (*CreateMazeReply, error)
	// List available mazes
	ListMazes(ctx context.Context, in *ListMazeRequest, opts ...grpc.CallOption) (*ListMazeReply, error)
	// Solve a maze, streaming, bidi
	SolveMaze(ctx context.Context, opts ...grpc.CallOption) (Mazer_SolveMazeClient, error)
}

type mazerClient struct {
	cc *grpc.ClientConn
}

func NewMazerClient(cc *grpc.ClientConn) MazerClient {
	return &mazerClient{cc}
}

func (c *mazerClient) CreateMaze(ctx context.Context, in *CreateMazeRequest, opts ...grpc.CallOption) (*CreateMazeReply, error) {
	out := new(CreateMazeReply)
	err := grpc.Invoke(ctx, "/proto.Mazer/CreateMaze", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mazerClient) ListMazes(ctx context.Context, in *ListMazeRequest, opts ...grpc.CallOption) (*ListMazeReply, error) {
	out := new(ListMazeReply)
	err := grpc.Invoke(ctx, "/proto.Mazer/ListMazes", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mazerClient) SolveMaze(ctx context.Context, opts ...grpc.CallOption) (Mazer_SolveMazeClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Mazer_serviceDesc.Streams[0], c.cc, "/proto.Mazer/SolveMaze", opts...)
	if err != nil {
		return nil, err
	}
	x := &mazerSolveMazeClient{stream}
	return x, nil
}

type Mazer_SolveMazeClient interface {
	Send(*SolveMazeRequest) error
	Recv() (*SolveMazeResponse, error)
	grpc.ClientStream
}

type mazerSolveMazeClient struct {
	grpc.ClientStream
}

func (x *mazerSolveMazeClient) Send(m *SolveMazeRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *mazerSolveMazeClient) Recv() (*SolveMazeResponse, error) {
	m := new(SolveMazeResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Mazer service

type MazerServer interface {
	// Display maze
	CreateMaze(context.Context, *CreateMazeRequest) (*CreateMazeReply, error)
	// List available mazes
	ListMazes(context.Context, *ListMazeRequest) (*ListMazeReply, error)
	// Solve a maze, streaming, bidi
	SolveMaze(Mazer_SolveMazeServer) error
}

func RegisterMazerServer(s *grpc.Server, srv MazerServer) {
	s.RegisterService(&_Mazer_serviceDesc, srv)
}

func _Mazer_CreateMaze_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateMazeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MazerServer).CreateMaze(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Mazer/CreateMaze",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MazerServer).CreateMaze(ctx, req.(*CreateMazeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Mazer_ListMazes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMazeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MazerServer).ListMazes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Mazer/ListMazes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MazerServer).ListMazes(ctx, req.(*ListMazeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Mazer_SolveMaze_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MazerServer).SolveMaze(&mazerSolveMazeServer{stream})
}

type Mazer_SolveMazeServer interface {
	Send(*SolveMazeResponse) error
	Recv() (*SolveMazeRequest, error)
	grpc.ServerStream
}

type mazerSolveMazeServer struct {
	grpc.ServerStream
}

func (x *mazerSolveMazeServer) Send(m *SolveMazeResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *mazerSolveMazeServer) Recv() (*SolveMazeRequest, error) {
	m := new(SolveMazeRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Mazer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Mazer",
	HandlerType: (*MazerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateMaze",
			Handler:    _Mazer_CreateMaze_Handler,
		},
		{
			MethodName: "ListMazes",
			Handler:    _Mazer_ListMazes_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SolveMaze",
			Handler:       _Mazer_SolveMaze_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "mazes.proto",
}

func init() { proto1.RegisterFile("mazes.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 938 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x56, 0xdd, 0x6e, 0x1a, 0x47,
	0x14, 0xce, 0x1a, 0x1b, 0xc3, 0x01, 0xc7, 0x30, 0xa6, 0x64, 0x4a, 0xac, 0x8a, 0x6e, 0x7b, 0x41,
	0xab, 0xca, 0x8a, 0xc8, 0x4d, 0xab, 0x56, 0x51, 0x6d, 0x68, 0x23, 0xa4, 0xa0, 0x46, 0xeb, 0x28,
	0x4e, 0x72, 0x83, 0x86, 0x65, 0x02, 0x23, 0x86, 0x1d, 0x3a, 0xb3, 0xc6, 0x72, 0xde, 0xa4, 0xea,
	0x2b, 0xf5, 0x2d, 0xfa, 0x22, 0xd5, 0x9c, 0xd9, 0x5d, 0x96, 0x1f, 0xf7, 0x22, 0x57, 0x9e, 0xf3,
	0x7d, 0xdf, 0xec, 0x9c, 0x99, 0xf3, 0x9d, 0x83, 0xa1, 0xb2, 0x60, 0x9f, 0xb8, 0xb9, 0x58, 0x6a,
	0x15, 0x2b, 0x72, 0x84, 0x7f, 0xfc, 0xbf, 0x3d, 0xa8, 0x5d, 0x2b, 0xb9, 0xe2, 0x43, 0xf6, 0x89,
	0x07, 0xfc, 0xcf, 0x5b, 0x6e, 0x62, 0xd2, 0x84, 0xa2, 0x95, 0x0e, 0x26, 0xd4, 0x6b, 0x7b, 0x9d,
	0x72, 0x90, 0x44, 0xa4, 0x05, 0xa5, 0x50, 0x0a, 0x1e, 0xc5, 0x83, 0x09, 0x3d, 0x40, 0x26, 0x8b,
	0xc9, 0x39, 0x94, 0x27, 0x42, 0xf3, 0x30, 0x16, 0x2a, 0xa2, 0x05, 0x24, 0xd7, 0x00, 0xa1, 0x70,
	0x2c, 0x22, 0x11, 0x0b, 0x26, 0xe9, 0x61, 0xdb, 0xeb, 0x94, 0x82, 0x34, 0x24, 0x4f, 0xa1, 0xbc,
	0x50, 0x2b, 0x3e, 0x1a, 0xb3, 0x70, 0x4e, 0x8f, 0x90, 0x2b, 0x59, 0xe0, 0x8a, 0x85, 0x73, 0xff,
	0xaf, 0x02, 0xd4, 0x73, 0xd9, 0x99, 0xa5, 0x8a, 0x0c, 0xff, 0xac, 0xf4, 0x7a, 0xd0, 0x60, 0x2b,
	0x26, 0x24, 0x1b, 0x4b, 0x3e, 0xca, 0xf2, 0x32, 0xb4, 0xd0, 0x2e, 0x74, 0x2a, 0xdd, 0x9a, 0x7b,
	0x94, 0x8b, 0x7e, 0x4a, 0x04, 0x67, 0x99, 0x3a, 0xc3, 0xcc, 0xff, 0xdc, 0xa2, 0x01, 0x47, 0x5c,
	0x6b, 0xa5, 0x93, 0x1b, 0xb8, 0x80, 0x7c, 0x03, 0x27, 0xb8, 0x18, 0x2d, 0xb8, 0x31, 0x6c, 0xca,
	0x69, 0x11, 0xb3, 0xaa, 0x22, 0x38, 0x74, 0x18, 0x79, 0x01, 0xb5, 0xf0, 0x56, 0x6b, 0x1e, 0xc5,
	0x23, 0xa9, 0x42, 0x86, 0xef, 0x77, 0xdc, 0xf6, 0x3a, 0x95, 0xee, 0x59, 0x92, 0x95, 0xbd, 0xfc,
	0xab, 0x84, 0x0a, 0x4e, 0x13, 0x71, 0x0a, 0x90, 0x67, 0x50, 0xfe, 0xa8, 0xd5, 0x62, 0x14, 0x72,
	0x29, 0x69, 0xe9, 0xe1, 0x8d, 0x25, 0xab, 0xea, 0x71, 0x29, 0xc9, 0x0f, 0x70, 0x1c, 0x2b, 0xa7,
	0x2f, 0x3f, 0xac, 0x2f, 0xc6, 0x0a, 0xd5, 0x4d, 0x28, 0x1a, 0x5b, 0x82, 0x09, 0x05, 0xbc, 0x5b,
	0x12, 0xf9, 0x3f, 0x41, 0x39, 0x7b, 0x1a, 0x42, 0xe0, 0x30, 0x62, 0x0b, 0x9e, 0x14, 0x04, 0xd7,
	0xf6, 0xb5, 0x56, 0xc2, 0x88, 0x98, 0xbb, 0x6a, 0x94, 0x82, 0x34, 0xf4, 0x7f, 0x81, 0x43, 0x7b,
	0xd4, 0x83, 0x85, 0x3c, 0x87, 0x72, 0x5a, 0x38, 0x43, 0x0f, 0xda, 0x05, 0xeb, 0xa5, 0x0c, 0xf0,
	0xeb, 0x70, 0xfa, 0x4a, 0x98, 0x38, 0x67, 0x58, 0xbf, 0x0b, 0x27, 0x6b, 0x68, 0x29, 0xef, 0xc9,
	0xd7, 0x70, 0x84, 0x66, 0xa7, 0x1e, 0xd6, 0xb7, 0x92, 0xbb, 0x60, 0xe0, 0x18, 0xff, 0x05, 0xd4,
	0x7b, 0x9a, 0xb3, 0x78, 0xc3, 0xf9, 0xdf, 0x41, 0x31, 0x54, 0xd1, 0x47, 0x31, 0xc5, 0x8c, 0x2a,
	0xdd, 0x7a, 0x6e, 0x63, 0x0f, 0x89, 0x20, 0x11, 0xf8, 0xbf, 0xc1, 0x69, 0x7e, 0xbf, 0x3d, 0xb5,
	0x09, 0xc5, 0xe1, 0xc6, 0x7d, 0x86, 0x99, 0x31, 0x7b, 0x5b, 0xc6, 0x4c, 0x63, 0xff, 0xdf, 0x63,
	0x80, 0xf5, 0xd7, 0xed, 0x43, 0x06, 0xea, 0xce, 0xe0, 0x07, 0x0a, 0x01, 0xae, 0xed, 0x43, 0xf6,
	0x94, 0xbc, 0x5d, 0x44, 0x06, 0x77, 0x17, 0x82, 0x34, 0x24, 0x3e, 0x54, 0x2f, 0xa5, 0x54, 0x77,
	0x37, 0x9c, 0xad, 0x44, 0x34, 0xc5, 0xbe, 0x2b, 0x05, 0x1b, 0x18, 0xb9, 0x00, 0x92, 0x2c, 0x5f,
	0x6b, 0x35, 0x66, 0x63, 0x21, 0x45, 0x7c, 0x8f, 0xfe, 0xf5, 0x82, 0x3d, 0x8c, 0x7d, 0x7c, 0x5b,
	0xf7, 0x1b, 0x31, 0x89, 0x67, 0x68, 0xe7, 0x42, 0xb0, 0x06, 0x2c, 0x7b, 0xc3, 0x52, 0xb6, 0xe8,
	0xd8, 0x0c, 0x48, 0xd9, 0xeb, 0x25, 0x0b, 0x39, 0x9a, 0x38, 0x61, 0x11, 0xb0, 0xec, 0x6b, 0x16,
	0xcf, 0xdc, 0xde, 0x92, 0x63, 0x33, 0x80, 0x7c, 0x0f, 0xb5, 0x21, 0xd3, 0xf3, 0xb7, 0xce, 0x23,
	0xf6, 0x44, 0x83, 0xf6, 0x2c, 0x05, 0x3b, 0xb8, 0xbd, 0xd3, 0xf5, 0x4c, 0xdd, 0xf5, 0x85, 0x89,
	0x59, 0x14, 0xf2, 0xb7, 0x4c, 0xde, 0x72, 0x93, 0xf8, 0x73, 0x0f, 0xb3, 0xad, 0xef, 0x29, 0xa9,
	0xb4, 0xa1, 0x95, 0x5d, 0xbd, 0x63, 0xc8, 0xb7, 0x70, 0x72, 0x3d, 0x17, 0xcb, 0x97, 0x5a, 0x4c,
	0x7a, 0x33, 0x1e, 0xce, 0x69, 0x15, 0xa5, 0x9b, 0x20, 0x79, 0x0e, 0xf0, 0x87, 0x5e, 0xce, 0x58,
	0x34, 0x64, 0x66, 0x4e, 0x4f, 0xd0, 0x69, 0x7b, 0x5b, 0x29, 0x27, 0x23, 0x6d, 0xa8, 0x5c, 0xae,
	0x58, 0xcc, 0xf4, 0x60, 0x61, 0x27, 0xc2, 0x63, 0xb4, 0x43, 0x1e, 0xb2, 0x0f, 0x91, 0xbb, 0x2c,
	0x66, 0x44, 0x4f, 0x51, 0xb6, 0x83, 0x5b, 0x6b, 0x5c, 0x4d, 0x9d, 0xa4, 0x86, 0x92, 0x34, 0xb4,
	0xe7, 0x5c, 0x29, 0x3d, 0xe1, 0xda, 0xb1, 0x75, 0x77, 0x4e, 0x0e, 0x4a, 0x8b, 0xe5, 0x78, 0xe2,
	0x26, 0x76, 0x06, 0xa4, 0xc5, 0x72, 0xec, 0x99, 0x63, 0x33, 0x80, 0x74, 0xa1, 0xd1, 0xdb, 0x9c,
	0x43, 0x4e, 0xd8, 0x40, 0xe1, 0x5e, 0xce, 0x3e, 0xea, 0xef, 0xc9, 0x08, 0x72, 0xe2, 0x2f, 0x50,
	0xbc, 0x09, 0xda, 0xbc, 0xdf, 0xa8, 0xb5, 0xa6, 0xe9, 0xf2, 0xce, 0x41, 0xb6, 0x9b, 0xd2, 0x2d,
	0xf4, 0x89, 0xeb, 0xa6, 0x34, 0xb6, 0x1d, 0xe8, 0xa4, 0x94, 0xba, 0x0e, 0x74, 0x91, 0x6d, 0x94,
	0x97, 0x3c, 0xea, 0x6b, 0x76, 0xd7, 0xe7, 0x92, 0xdd, 0xd3, 0x2f, 0xdd, 0x20, 0xce, 0x63, 0xe4,
	0x2b, 0x00, 0xd7, 0xd0, 0x97, 0x72, 0xaa, 0x68, 0x0b, 0x15, 0x39, 0xc4, 0xd6, 0xc5, 0x5a, 0xc5,
	0x9e, 0xf5, 0x46, 0x25, 0x16, 0x7a, 0xea, 0x0c, 0xba, 0x8d, 0x93, 0xc7, 0x70, 0x30, 0x98, 0xd0,
	0x73, 0xfc, 0xc6, 0xc1, 0x60, 0xe2, 0x17, 0xe1, 0xd0, 0xe6, 0xe1, 0xff, 0x08, 0xd5, 0xbc, 0x33,
	0x48, 0x15, 0xbc, 0x77, 0x49, 0xaf, 0x7b, 0xef, 0x6c, 0xf4, 0x3e, 0x69, 0x71, 0xef, 0xbd, 0x8d,
	0x3e, 0x60, 0x47, 0x17, 0x02, 0xef, 0x43, 0xf7, 0x1f, 0x0f, 0x8e, 0xec, 0x56, 0x4d, 0x7e, 0x4d,
	0xf3, 0xc4, 0x19, 0x4a, 0x13, 0xc3, 0xed, 0xcc, 0xb2, 0x56, 0x73, 0x0f, 0xb3, 0x94, 0xf7, 0xfe,
	0x23, 0xf2, 0x33, 0x94, 0xd3, 0x71, 0x69, 0x48, 0x2a, 0xdb, 0x9a, 0xa9, 0xad, 0xc6, 0x0e, 0xee,
	0x36, 0xf7, 0xa1, 0x9c, 0xfd, 0x24, 0x93, 0x27, 0x89, 0x68, 0xfb, 0x5f, 0x88, 0x16, 0xdd, 0x25,
	0xdc, 0xaf, 0xb7, 0xff, 0xa8, 0xe3, 0x3d, 0xf3, 0xc6, 0x45, 0xa4, 0x9f, 0xff, 0x17, 0x00, 0x00,
	0xff, 0xff, 0xd1, 0x6b, 0x99, 0xbb, 0x94, 0x08, 0x00, 0x00,
}
