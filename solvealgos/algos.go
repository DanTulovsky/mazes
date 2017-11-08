// Package solvealgos defines the interface for solver algorithms
package solvealgos

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	"mazes/maze"
	"mazes/ml"
	pb "mazes/proto"

	"context"

	"github.com/rcrowley/go-metrics"
)

var (
	// stats
	showStats = flag.Bool("stats", false, "show maze stats")
)

const (
	ADDRESS = "localhost:50051"
)

type Algorithmer interface {
	Move(d string, mazeID string, clientID string) (*pb.SolveMazeResponse, error) // move a direction
	MoveBack(mazeID string, clientID string) (*pb.SolveMazeResponse, error)       // move back
	SolvePath() *maze.Path                                                        // final path
	SolveSteps() int
	SolveTime() time.Duration
	SetPolicy(p *ml.Policy)
	SetSolvePath(p *maze.Path)
	SetSolveSteps(s int)
	SetSolveTime(t time.Duration)
	// m is the *local* maze for display only
	// p is the ML policy to follow, only used when from_policy algo is used
	Solve(mazeID, client string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
		directions []*pb.Direction, m *maze.Maze) error
	Stream() pb.Mazer_SolveMazeClient
	SetStream(pb.Mazer_SolveMazeClient)
	ShowStats()
	TravelPath() *maze.Path // all the cells traveled
	CellForLocation(m *maze.Maze, l *pb.MazeLocation) (*maze.Cell, error)
}

type Common struct {
	solvePath  *maze.Path    // path of the final solution
	solveSteps int           // how many cell visits it took (including duplicates)
	solveTime  time.Duration // how long the last solve time took
	stream     pb.Mazer_SolveMazeClient
	travelPath *maze.Path // all the cells visited in order
	policy     *ml.Policy
}

func (a *Common) CellForLocation(m *maze.Maze, l *pb.MazeLocation) (*maze.Cell, error) {
	cell, err := m.Cell(l.GetX(), l.GetY(), l.GetZ())
	if err != nil {
		return nil, err
	}
	return cell, nil
}

// Solve should write the path of the solution to the grid
func (a *Common) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
	directions []*pb.Direction, m *maze.Maze) error {
	return errors.New("Solve() not implemented")
}

// SetPolicy sets the solvePath
func (a *Common) SetPolicy(p *ml.Policy) {
	a.policy = p
}

// Policy sets the solvePath
func (a *Common) Policy() *ml.Policy {
	return a.policy
}

// SetSolvePath sets the solvePath
func (a *Common) SetSolvePath(p *maze.Path) {
	a.solvePath = p
}

// TimeTrack tracks sets the time it took for the algorithm to run
func TimeTrack(a Algorithmer, start time.Time) {
	a.SetSolveTime(time.Since(start))
}

// SolveTime returns the time it took to solve the maze
func (a *Common) SolveTime() time.Duration {
	return a.solveTime
}

// SetSolveTime sets solveTime
func (a *Common) SetSolveTime(t time.Duration) {
	a.solveTime = t
}

// SolvePath returns the path for the solution
func (a *Common) SolvePath() *maze.Path {
	return a.solvePath
}

// SolveSteps returns the number of steps (visits to cells) it took to solve the maze
func (a *Common) SolveSteps() int {
	return a.travelPath.Length()
}

// SetSolveSteps sets the solveSteps
func (a *Common) SetSolveSteps(s int) {
	a.solveSteps = s
}

// TravelPath returns the entire path traveled (often the same as the solution path)
func (a *Common) TravelPath() *maze.Path {
	return a.travelPath
}

// SetTravelPath sets the solvePath
func (a *Common) SetTravelPath(p *maze.Path) {
	a.travelPath = p
}

func (a *Common) Stream() pb.Mazer_SolveMazeClient {
	return a.stream
}

func (a *Common) SetStream(s pb.Mazer_SolveMazeClient) {
	a.stream = s
}

// NewClient creates a server connection and returns a new SoleMazeClient
func NewClient() (*grpc.ClientConn, pb.MazerClient) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(ADDRESS, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// defer conn.Close()
	return conn, pb.NewMazerClient(conn)
}

func (a *Common) ResetClient(mazeID, clientID string) (*pb.ResetClientReply, error) {

	conn, c := NewClient()
	defer conn.Close()

	ctx := context.Background()
	r, err := c.ResetClient(ctx,
		&pb.ResetClientRequest{
			MazeId:   mazeID,
			ClientId: clientID,
		})

	return r, err
}

// Move sends a move request to the server and returns the reply
func (a *Common) Move(mazeID, clientID, d string) (*pb.SolveMazeResponse, error) {
	t := metrics.GetOrRegisterTimer("solver.step.latency", nil)
	defer t.UpdateSince(time.Now())

	stream := a.Stream()

	r := &pb.SolveMazeRequest{
		MazeId:    mazeID,
		ClientId:  clientID,
		Direction: d,
	}
	if err := stream.Send(r); err != nil {
		log.Printf(">> %v", err)
		return nil, err
	}

	reply, err := stream.Recv()
	if err != nil {
		// log.Printf(">>> %v", err)
		return reply, err
	}

	if reply.GetError() {
		// log.Printf(">>>> %v", reply.GetErrorMessage())
		return reply, fmt.Errorf("%v", reply.GetErrorMessage())
	}

	return reply, nil

}

// MoveBack moves the client back to the previous location (where they just came from)
func (a *Common) MoveBack(mazeID, clientID string) (*pb.SolveMazeResponse, error) {
	t := metrics.GetOrRegisterTimer("solver.step.latency", nil)
	defer t.UpdateSince(time.Now())

	stream := a.Stream()

	r := &pb.SolveMazeRequest{
		MazeId:   mazeID,
		ClientId: clientID,
		MoveBack: true,
	}
	if err := stream.Send(r); err != nil {
		return nil, err
	}

	reply, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	if reply.GetError() {
		return nil, fmt.Errorf("%v", reply.GetErrorMessage())
	}

	return reply, nil
}

// UpdateClientViewAndLocation sets the current location of the client in the local maze
// steps is the number of steps it took to get to this cell, overwritten by latest visit
func (a *Common) UpdateClientViewAndLocation(clientID string, m *maze.Maze, currentCell, previousCell *pb.MazeLocation, steps int) error {
	if m == nil {
		return nil // no client maze requested
	}

	client, err := m.Client(clientID)
	if err != nil {
		return err
	}

	var cell, pcell *maze.Cell

	if cell, err = a.CellForLocation(m, currentCell); err != nil {
		return err
	}

	client.SetCurrentLocation(cell)
	cell.SetVisited(clientID)
	cell.SetDistance(steps)

	if previousCell != nil {
		if pcell, err = a.CellForLocation(m, previousCell); err != nil {
			return err
		}
		m.Link(pcell, cell)
	}

	return nil
}

func (a *Common) ShowStats() {
	if *showStats {
		log.Printf("TODO: show stats")
	}
}
