// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mazes.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	mazes.proto

It has these top-level messages:
	RegisterClientRequest
	RegisterClientReply
	SolveMazeRequest
	SolveMazeResponse
	Direction
	Maze
	ListMazeRequest
	ListMazeReply
	CreateMazeRequest
	CreateMazeReply
	MazeConfig
	ClientConfig
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

type RegisterClientRequest struct {
	MazeId       string        `protobuf:"bytes,1,opt,name=mazeId" json:"mazeId,omitempty"`
	ClientConfig *ClientConfig `protobuf:"bytes,2,opt,name=client_config,json=clientConfig" json:"client_config,omitempty"`
}

func (m *RegisterClientRequest) Reset()                    { *m = RegisterClientRequest{} }
func (m *RegisterClientRequest) String() string            { return proto1.CompactTextString(m) }
func (*RegisterClientRequest) ProtoMessage()               {}
func (*RegisterClientRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *RegisterClientRequest) GetMazeId() string {
	if m != nil {
		return m.MazeId
	}
	return ""
}

func (m *RegisterClientRequest) GetClientConfig() *ClientConfig {
	if m != nil {
		return m.ClientConfig
	}
	return nil
}

type RegisterClientReply struct {
	Success  bool   `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
	Message  string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
	ClientId string `protobuf:"bytes,3,opt,name=client_id,json=clientId" json:"client_id,omitempty"`
}

func (m *RegisterClientReply) Reset()                    { *m = RegisterClientReply{} }
func (m *RegisterClientReply) String() string            { return proto1.CompactTextString(m) }
func (*RegisterClientReply) ProtoMessage()               {}
func (*RegisterClientReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RegisterClientReply) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *RegisterClientReply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *RegisterClientReply) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

// SolveMazeRequest is a message sent from the client trying to solve a maze
type SolveMazeRequest struct {
	MazeId    string `protobuf:"bytes,1,opt,name=mazeId" json:"mazeId,omitempty"`
	ClientId  string `protobuf:"bytes,2,opt,name=client_id,json=clientId" json:"client_id,omitempty"`
	Direction string `protobuf:"bytes,3,opt,name=direction" json:"direction,omitempty"`
	// on first connect, this must be set true, the direction field is ignored, the client does not move
	Initial bool `protobuf:"varint,4,opt,name=initial" json:"initial,omitempty"`
	// move client back to previous location, direction is ignored
	MoveBack bool `protobuf:"varint,5,opt,name=move_back,json=moveBack" json:"move_back,omitempty"`
}

func (m *SolveMazeRequest) Reset()                    { *m = SolveMazeRequest{} }
func (m *SolveMazeRequest) String() string            { return proto1.CompactTextString(m) }
func (*SolveMazeRequest) ProtoMessage()               {}
func (*SolveMazeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

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
	MazeId              string        `protobuf:"bytes,1,opt,name=maze_id,json=mazeId" json:"maze_id,omitempty"`
	ClientId            string        `protobuf:"bytes,2,opt,name=client_id,json=clientId" json:"client_id,omitempty"`
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
func (*SolveMazeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

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
func (*Direction) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

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
func (*Maze) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

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
func (*ListMazeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type ListMazeReply struct {
	Mazes []*Maze `protobuf:"bytes,1,rep,name=mazes" json:"mazes,omitempty"`
}

func (m *ListMazeReply) Reset()                    { *m = ListMazeReply{} }
func (m *ListMazeReply) String() string            { return proto1.CompactTextString(m) }
func (*ListMazeReply) ProtoMessage()               {}
func (*ListMazeReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

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
func (*CreateMazeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

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
func (*CreateMazeReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

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
	BgColor              string          `protobuf:"bytes,16,opt,name=BgColor" json:"BgColor,omitempty"`
	BorderColor          string          `protobuf:"bytes,17,opt,name=BorderColor" json:"BorderColor,omitempty"`
	WallColor            string          `protobuf:"bytes,18,opt,name=WallColor" json:"WallColor,omitempty"`
	CurrentLocationColor string          `protobuf:"bytes,20,opt,name=CurrentLocationColor" json:"CurrentLocationColor,omitempty"`
	GenDrawDelay         string          `protobuf:"bytes,25,opt,name=GenDrawDelay" json:"GenDrawDelay,omitempty"`
	CreateAlgo           string          `protobuf:"bytes,26,opt,name=CreateAlgo" json:"CreateAlgo,omitempty"`
	BraidProbability     float64         `protobuf:"fixed64,27,opt,name=BraidProbability" json:"BraidProbability,omitempty"`
	Id                   string          `protobuf:"bytes,28,opt,name=Id" json:"Id,omitempty"`
}

func (m *MazeConfig) Reset()                    { *m = MazeConfig{} }
func (m *MazeConfig) String() string            { return proto1.CompactTextString(m) }
func (*MazeConfig) ProtoMessage()               {}
func (*MazeConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

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

func (m *MazeConfig) GetCurrentLocationColor() string {
	if m != nil {
		return m.CurrentLocationColor
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

func (m *MazeConfig) GetBraidProbability() float64 {
	if m != nil {
		return m.BraidProbability
	}
	return 0
}

func (m *MazeConfig) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

// ClientConfig has all the per-client config settings in it
type ClientConfig struct {
	SolveAlgo        string `protobuf:"bytes,1,opt,name=SolveAlgo" json:"SolveAlgo,omitempty"`
	AvatarImage      string `protobuf:"bytes,14,opt,name=AvatarImage" json:"AvatarImage,omitempty"`
	VisitedCellColor string `protobuf:"bytes,15,opt,name=VisitedCellColor" json:"VisitedCellColor,omitempty"`
	PathColor        string `protobuf:"bytes,19,opt,name=PathColor" json:"PathColor,omitempty"`
	FromCellColor    string `protobuf:"bytes,21,opt,name=FromCellColor" json:"FromCellColor,omitempty"`
	ToCellColor      string `protobuf:"bytes,22,opt,name=ToCellColor" json:"ToCellColor,omitempty"`
	FromCell         string `protobuf:"bytes,23,opt,name=FromCell" json:"FromCell,omitempty"`
	ToCell           string `protobuf:"bytes,24,opt,name=ToCell" json:"ToCell,omitempty"`
	ShowFromToColors bool   `protobuf:"varint,27,opt,name=ShowFromToColors" json:"ShowFromToColors,omitempty"`
}

func (m *ClientConfig) Reset()                    { *m = ClientConfig{} }
func (m *ClientConfig) String() string            { return proto1.CompactTextString(m) }
func (*ClientConfig) ProtoMessage()               {}
func (*ClientConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *ClientConfig) GetSolveAlgo() string {
	if m != nil {
		return m.SolveAlgo
	}
	return ""
}

func (m *ClientConfig) GetAvatarImage() string {
	if m != nil {
		return m.AvatarImage
	}
	return ""
}

func (m *ClientConfig) GetVisitedCellColor() string {
	if m != nil {
		return m.VisitedCellColor
	}
	return ""
}

func (m *ClientConfig) GetPathColor() string {
	if m != nil {
		return m.PathColor
	}
	return ""
}

func (m *ClientConfig) GetFromCellColor() string {
	if m != nil {
		return m.FromCellColor
	}
	return ""
}

func (m *ClientConfig) GetToCellColor() string {
	if m != nil {
		return m.ToCellColor
	}
	return ""
}

func (m *ClientConfig) GetFromCell() string {
	if m != nil {
		return m.FromCell
	}
	return ""
}

func (m *ClientConfig) GetToCell() string {
	if m != nil {
		return m.ToCell
	}
	return ""
}

func (m *ClientConfig) GetShowFromToColors() bool {
	if m != nil {
		return m.ShowFromToColors
	}
	return false
}

type Cell struct {
}

func (m *Cell) Reset()                    { *m = Cell{} }
func (m *Cell) String() string            { return proto1.CompactTextString(m) }
func (*Cell) ProtoMessage()               {}
func (*Cell) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

// MazeLocation is a location in the maze
type MazeLocation struct {
	X int64 `protobuf:"varint,1,opt,name=X" json:"X,omitempty"`
	Y int64 `protobuf:"varint,2,opt,name=Y" json:"Y,omitempty"`
	Z int64 `protobuf:"varint,3,opt,name=Z" json:"Z,omitempty"`
}

func (m *MazeLocation) Reset()                    { *m = MazeLocation{} }
func (m *MazeLocation) String() string            { return proto1.CompactTextString(m) }
func (*MazeLocation) ProtoMessage()               {}
func (*MazeLocation) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

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
	proto1.RegisterType((*RegisterClientRequest)(nil), "proto.RegisterClientRequest")
	proto1.RegisterType((*RegisterClientReply)(nil), "proto.RegisterClientReply")
	proto1.RegisterType((*SolveMazeRequest)(nil), "proto.SolveMazeRequest")
	proto1.RegisterType((*SolveMazeResponse)(nil), "proto.SolveMazeResponse")
	proto1.RegisterType((*Direction)(nil), "proto.Direction")
	proto1.RegisterType((*Maze)(nil), "proto.Maze")
	proto1.RegisterType((*ListMazeRequest)(nil), "proto.ListMazeRequest")
	proto1.RegisterType((*ListMazeReply)(nil), "proto.ListMazeReply")
	proto1.RegisterType((*CreateMazeRequest)(nil), "proto.CreateMazeRequest")
	proto1.RegisterType((*CreateMazeReply)(nil), "proto.CreateMazeReply")
	proto1.RegisterType((*MazeConfig)(nil), "proto.MazeConfig")
	proto1.RegisterType((*ClientConfig)(nil), "proto.ClientConfig")
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
	// Register new client with an existing maze
	RegisterClient(ctx context.Context, in *RegisterClientRequest, opts ...grpc.CallOption) (*RegisterClientReply, error)
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

func (c *mazerClient) RegisterClient(ctx context.Context, in *RegisterClientRequest, opts ...grpc.CallOption) (*RegisterClientReply, error) {
	out := new(RegisterClientReply)
	err := grpc.Invoke(ctx, "/proto.Mazer/RegisterClient", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Mazer service

type MazerServer interface {
	// Display maze
	CreateMaze(context.Context, *CreateMazeRequest) (*CreateMazeReply, error)
	// List available mazes
	ListMazes(context.Context, *ListMazeRequest) (*ListMazeReply, error)
	// Solve a maze, streaming, bidi
	SolveMaze(Mazer_SolveMazeServer) error
	// Register new client with an existing maze
	RegisterClient(context.Context, *RegisterClientRequest) (*RegisterClientReply, error)
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

func _Mazer_RegisterClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterClientRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MazerServer).RegisterClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Mazer/RegisterClient",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MazerServer).RegisterClient(ctx, req.(*RegisterClientRequest))
	}
	return interceptor(ctx, in, info, handler)
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
		{
			MethodName: "RegisterClient",
			Handler:    _Mazer_RegisterClient_Handler,
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
	// 1059 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x56, 0x5f, 0x4f, 0x1b, 0xc7,
	0x17, 0xcd, 0xda, 0xd8, 0xd8, 0xd7, 0x26, 0xc0, 0x40, 0xc8, 0xfc, 0x0c, 0xfa, 0x89, 0x6e, 0xfb,
	0x40, 0xab, 0x0a, 0x45, 0xe4, 0x25, 0x55, 0xab, 0xa8, 0x60, 0xda, 0xc8, 0x12, 0xa8, 0xd1, 0x12,
	0x85, 0x24, 0x2f, 0x68, 0xd8, 0x1d, 0xcc, 0xc8, 0xeb, 0x1d, 0x77, 0x66, 0x31, 0x22, 0x1f, 0xa5,
	0xed, 0x87, 0xe9, 0xf7, 0xea, 0x4b, 0x75, 0xef, 0xcc, 0xda, 0xeb, 0x3f, 0xb4, 0xea, 0x13, 0xbe,
	0xe7, 0x9c, 0x99, 0xb9, 0x33, 0xf7, 0xdc, 0xcb, 0x42, 0x6b, 0x28, 0x3e, 0x4b, 0x7b, 0x38, 0x32,
	0x3a, 0xd7, 0xac, 0x46, 0x7f, 0x42, 0x05, 0xcf, 0x22, 0xd9, 0x57, 0x36, 0x97, 0xa6, 0x9b, 0x2a,
	0x99, 0xe5, 0x91, 0xfc, 0xf5, 0x4e, 0xda, 0x9c, 0xed, 0x40, 0x1d, 0xe5, 0xbd, 0x84, 0x07, 0xfb,
	0xc1, 0x41, 0x33, 0xf2, 0x11, 0x7b, 0x05, 0x6b, 0x31, 0x09, 0xaf, 0x62, 0x9d, 0xdd, 0xa8, 0x3e,
	0xaf, 0xec, 0x07, 0x07, 0xad, 0xa3, 0x2d, 0xb7, 0xed, 0xa1, 0xdb, 0xa4, 0x4b, 0x54, 0xd4, 0x8e,
	0x4b, 0x51, 0x78, 0x03, 0x5b, 0xf3, 0x47, 0x8d, 0xd2, 0x07, 0xc6, 0x61, 0xd5, 0xde, 0xc5, 0xb1,
	0xb4, 0x96, 0x4e, 0x6a, 0x44, 0x45, 0x88, 0xcc, 0x50, 0x5a, 0x2b, 0xfa, 0x92, 0x0e, 0x69, 0x46,
	0x45, 0xc8, 0x76, 0xa1, 0xe9, 0x93, 0x50, 0x09, 0xaf, 0x12, 0xd7, 0x70, 0x40, 0x2f, 0x09, 0xff,
	0x08, 0x60, 0xe3, 0x42, 0xa7, 0x63, 0x79, 0x2e, 0x3e, 0xcb, 0x7f, 0xbb, 0xce, 0xcc, 0x4e, 0x95,
	0xd9, 0x9d, 0xd8, 0x1e, 0x34, 0x13, 0x65, 0x64, 0x9c, 0x2b, 0x9d, 0xf9, 0x63, 0xa6, 0x00, 0xa6,
	0xa7, 0x32, 0x95, 0x2b, 0x91, 0xf2, 0x15, 0x97, 0xb8, 0x0f, 0x71, 0xd3, 0xa1, 0x1e, 0xcb, 0xab,
	0x6b, 0x11, 0x0f, 0x78, 0x8d, 0xb8, 0x06, 0x02, 0x27, 0x22, 0x1e, 0x84, 0xbf, 0x57, 0x61, 0xb3,
	0x94, 0x9e, 0x1d, 0xe9, 0xcc, 0x4a, 0xf6, 0x1c, 0x56, 0x31, 0x23, 0xcc, 0xe2, 0x3f, 0x24, 0xd8,
	0x85, 0x6d, 0x31, 0x16, 0x2a, 0x15, 0xd7, 0xa9, 0xbc, 0x9a, 0x64, 0x66, 0x79, 0x75, 0xbf, 0x7a,
	0xd0, 0x3a, 0xda, 0xf0, 0x35, 0x39, 0x2d, 0x88, 0x68, 0x6b, 0xa2, 0x9e, 0x60, 0xf6, 0x1f, 0xee,
	0xb1, 0x0d, 0x35, 0x69, 0x8c, 0x36, 0xfe, 0x0e, 0x2e, 0x60, 0x5f, 0xc2, 0x1a, 0xfd, 0xb8, 0x2a,
	0x8a, 0x53, 0xa7, 0xac, 0xda, 0x04, 0x9e, 0xfb, 0x0a, 0xbd, 0x86, 0x8d, 0xf8, 0xce, 0x18, 0xcc,
	0x3b, 0xd5, 0xb1, 0xa0, 0x17, 0x5c, 0x9d, 0x71, 0x0a, 0x5e, 0xff, 0xcc, 0x53, 0xd1, 0xba, 0x17,
	0x17, 0x00, 0x7b, 0x01, 0xcd, 0x1b, 0xa3, 0x87, 0x57, 0xb1, 0x4c, 0x53, 0xde, 0x78, 0x7c, 0x61,
	0x03, 0x55, 0x5d, 0x99, 0xa6, 0xec, 0x5b, 0x58, 0xcd, 0xb5, 0xd3, 0x37, 0x1f, 0xd7, 0xd7, 0x73,
	0x4d, 0xea, 0x1d, 0xa8, 0x5b, 0x2c, 0x42, 0xc2, 0x81, 0xee, 0xe6, 0xa3, 0xf0, 0x3b, 0x68, 0x4e,
	0x9e, 0x86, 0x31, 0x58, 0xc9, 0xc4, 0x50, 0xfa, 0x8a, 0xd0, 0x6f, 0x7c, 0xad, 0xb1, 0xb2, 0x2a,
	0x97, 0xae, 0x1a, 0x8d, 0xa8, 0x08, 0xc3, 0x1f, 0x60, 0x05, 0x8f, 0x7a, 0xd4, 0x6a, 0x7b, 0x45,
	0x25, 0x7b, 0x89, 0xe5, 0x95, 0xfd, 0x2a, 0xba, 0x69, 0x02, 0x84, 0x9b, 0xb0, 0x7e, 0xa6, 0x6c,
	0x5e, 0xf2, 0x6c, 0x78, 0x04, 0x6b, 0x53, 0x08, 0x5b, 0xe5, 0x0b, 0xa8, 0x51, 0x0b, 0xf3, 0x80,
	0xea, 0xdb, 0x2a, 0x5d, 0x30, 0x72, 0x4c, 0xf8, 0x1a, 0x36, 0xbb, 0x46, 0x8a, 0x7c, 0xc6, 0xfc,
	0x5f, 0x43, 0xdd, 0x37, 0x6b, 0x40, 0x2f, 0xb3, 0x59, 0x5a, 0xe8, 0x5b, 0xd5, 0x0b, 0xc2, 0x9f,
	0x60, 0xbd, 0xbc, 0x1e, 0x4f, 0xdd, 0x81, 0xfa, 0xf9, 0xcc, 0x7d, 0x5c, 0xc4, 0x3a, 0xd0, 0xe8,
	0xfa, 0xf4, 0x0b, 0x63, 0x16, 0x71, 0xf8, 0x57, 0x0d, 0x60, 0xba, 0x3b, 0x3e, 0x64, 0xa4, 0xef,
	0x5d, 0x83, 0x57, 0x23, 0xfa, 0x8d, 0x0f, 0xd9, 0xd5, 0xe9, 0xdd, 0x30, 0xb3, 0xb4, 0xba, 0x1a,
	0x15, 0x21, 0x0b, 0xa1, 0x7d, 0x9c, 0xa6, 0xfa, 0xfe, 0x52, 0x8a, 0xb1, 0xca, 0xfa, 0xd4, 0x79,
	0x8d, 0x68, 0x06, 0x63, 0x87, 0xc0, 0xfc, 0xcf, 0xb7, 0x46, 0x5f, 0x8b, 0x6b, 0x95, 0xaa, 0xfc,
	0x81, 0xfc, 0x1b, 0x44, 0x4b, 0x18, 0x7c, 0x7c, 0xac, 0xfb, 0xa5, 0x4a, 0xf2, 0x5b, 0xb2, 0x73,
	0x35, 0x9a, 0x02, 0xc8, 0x5e, 0x8a, 0x82, 0xad, 0x3b, 0x76, 0x02, 0x14, 0xec, 0xc5, 0x48, 0xc4,
	0x92, 0x4c, 0xec, 0x59, 0x02, 0x90, 0x7d, 0x2b, 0xf2, 0x5b, 0xb7, 0xb6, 0xe1, 0xd8, 0x09, 0xc0,
	0xbe, 0x81, 0x8d, 0x73, 0x61, 0x06, 0xef, 0x9d, 0x47, 0xf0, 0x44, 0x4b, 0xf6, 0x6c, 0x44, 0x0b,
	0x38, 0xde, 0xe9, 0xe2, 0x56, 0xdf, 0x9f, 0x2a, 0x9b, 0x8b, 0x2c, 0x96, 0xef, 0x45, 0x7a, 0x27,
	0xad, 0xf7, 0xe7, 0x12, 0x66, 0x5e, 0xdf, 0xd5, 0xa9, 0x36, 0x96, 0xb7, 0x16, 0xf5, 0x8e, 0x61,
	0x5f, 0xc1, 0xda, 0xc5, 0x40, 0x8d, 0xde, 0x18, 0x95, 0x74, 0x6f, 0x65, 0x3c, 0xe0, 0x6d, 0x92,
	0xce, 0x82, 0xec, 0x25, 0xc0, 0x2f, 0x66, 0x74, 0x2b, 0xb2, 0x73, 0x61, 0x07, 0x7c, 0x8d, 0x9c,
	0xb6, 0xb4, 0x95, 0x4a, 0x32, 0x2c, 0xe6, 0x49, 0x9f, 0x8e, 0xe1, 0x1b, 0x6e, 0x54, 0xfb, 0x90,
	0xed, 0x43, 0xeb, 0x44, 0x9b, 0x44, 0x1a, 0xc7, 0x6e, 0x12, 0x5b, 0x86, 0x8a, 0xe7, 0x75, 0x3c,
	0x73, 0x53, 0x76, 0x02, 0xb0, 0x23, 0xd8, 0xee, 0xce, 0xce, 0x06, 0x27, 0xdc, 0x26, 0xe1, 0x52,
	0x0e, 0x0d, 0xf4, 0x46, 0x66, 0xa7, 0x46, 0xdc, 0x9f, 0xca, 0x54, 0x3c, 0xf0, 0xff, 0xb9, 0x01,
	0x55, 0xc6, 0xd8, 0xff, 0x01, 0x9c, 0xd1, 0x8f, 0xd3, 0xbe, 0xe6, 0x1d, 0x52, 0x94, 0x10, 0x2c,
	0xdc, 0x89, 0x11, 0x2a, 0x29, 0xdb, 0x6b, 0x97, 0xec, 0xb5, 0x80, 0xb3, 0xa7, 0x50, 0xe9, 0x25,
	0x7c, 0x8f, 0xf6, 0xa8, 0xf4, 0x92, 0xf0, 0xcf, 0x0a, 0xb4, 0xcb, 0xff, 0x08, 0xf1, 0x8a, 0x34,
	0xf2, 0xe9, 0x2c, 0xd7, 0x45, 0x53, 0x00, 0x9f, 0xe8, 0x78, 0x2c, 0x72, 0x61, 0x7a, 0x43, 0x1c,
	0xa7, 0x4f, 0xdd, 0x13, 0x95, 0x20, 0x4c, 0xa6, 0xe4, 0x14, 0xf7, 0x00, 0xeb, 0x24, 0x5b, 0xc0,
	0x0b, 0x3f, 0x3a, 0xd1, 0x96, 0x3b, 0x6b, 0x02, 0xa0, 0x07, 0x7e, 0xf6, 0x13, 0xd3, 0x29, 0x9e,
	0x91, 0x62, 0x16, 0xc4, 0x8c, 0xde, 0xe9, 0xa9, 0x66, 0xc7, 0x65, 0x54, 0x82, 0xb0, 0xf9, 0x8b,
	0x25, 0xfc, 0xb9, 0x6b, 0xfe, 0x22, 0xc6, 0x81, 0xe1, 0xa4, 0x9c, 0xbb, 0x81, 0xe1, 0x22, 0xbc,
	0x05, 0xba, 0x12, 0x75, 0xef, 0xb4, 0x77, 0xeb, 0xae, 0xeb, 0x85, 0x79, 0x3c, 0xac, 0xc3, 0x0a,
	0xae, 0x09, 0x5f, 0x41, 0xbb, 0x6c, 0x3a, 0xd6, 0x86, 0xe0, 0x83, 0x1f, 0x23, 0xc1, 0x07, 0x8c,
	0x3e, 0xfa, 0xe9, 0x11, 0x7c, 0xc4, 0xe8, 0x13, 0x0d, 0x8b, 0x6a, 0x14, 0x7c, 0x3a, 0xfa, 0xad,
	0x02, 0x35, 0x5c, 0x6a, 0xd8, 0x8f, 0x45, 0xa9, 0x69, 0x3c, 0xf3, 0xe2, 0x4b, 0x65, 0x7e, 0x4c,
	0x76, 0x76, 0x96, 0x30, 0xa3, 0xf4, 0x21, 0x7c, 0xc2, 0xbe, 0x87, 0x66, 0x31, 0x89, 0x2d, 0x2b,
	0x64, 0x73, 0xe3, 0xba, 0xb3, 0xbd, 0x80, 0xbb, 0xc5, 0xa7, 0xbe, 0xf8, 0x74, 0xfa, 0x73, 0x2f,
	0x9a, 0xff, 0x40, 0xe9, 0xf0, 0x45, 0xc2, 0x7d, 0x1a, 0x84, 0x4f, 0x0e, 0x82, 0x17, 0x01, 0x3b,
	0x83, 0xa7, 0xb3, 0x5f, 0x4f, 0x6c, 0xcf, 0xaf, 0x58, 0xfa, 0xfd, 0xd6, 0xe9, 0x3c, 0xc2, 0x52,
	0x4e, 0xd7, 0x75, 0x22, 0x5f, 0xfe, 0x1d, 0x00, 0x00, 0xff, 0xff, 0xbb, 0x56, 0x7c, 0x5c, 0x13,
	0x0a, 0x00, 0x00,
}
