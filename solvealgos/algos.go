// Package solvealgos defines the interface for solver algorithms
package solvealgos

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/rcrowley/go-metrics"
	"mazes/maze"
	pb "mazes/proto"
)

var (
	// stats
	showStats = flag.Bool("stats", false, "show maze stats")
)

type Algorithmer interface {
	Move(d string, mazeID string, clientID string) (*pb.SolveMazeResponse, error) // move a direction
	MoveBack(mazeID string, clientID string) (*pb.SolveMazeResponse, error)       // move back
	SolvePath() *maze.Path                                                        // final path
	SolveSteps() int
	SolveTime() time.Duration
	SetSolvePath(p *maze.Path)
	SetSolveSteps(s int)
	SetSolveTime(t time.Duration)
	Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration, directions []*pb.Direction) error
	Stream() pb.Mazer_SolveMazeClient
	SetStream(pb.Mazer_SolveMazeClient)
	ShowStats()
	TravelPath() *maze.Path // all the cells traveled
}

type Common struct {
	solvePath  *maze.Path    // path of the final solution
	solveSteps int           // how many cell visits it tooks (including duplicates)
	solveTime  time.Duration // how long the last solve time took
	stream     pb.Mazer_SolveMazeClient
	travelPath *maze.Path // all the cells visited in order
}

// Solve should write the path of the solution to the grid
func (a *Common) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration, directions []*pb.Direction) error {
	return errors.New("Solve() not implemented")
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
		return nil, err
	}

	reply, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	if reply.Error {
		return nil, fmt.Errorf("%v", reply.ErrorMessage)
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

	if reply.Error {
		return nil, fmt.Errorf("%v", reply.ErrorMessage)
	}

	return reply, nil
}

func (a *Common) ShowStats() {
	if *showStats {
		log.Printf("TODO: show stats")
	}
}
