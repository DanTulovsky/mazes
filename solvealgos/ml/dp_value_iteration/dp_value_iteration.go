package dp_value_iteration

import (
	"log"
	"mazes/maze"
	"mazes/solvealgos"
	"time"

	"fmt"
	"mazes/algos"
	pb "mazes/proto"
)

type DPValueIteration struct {
	solvealgos.Common
}

// nextDirection gives the next direction from this cell using the dp policy
func nextDirection() string {
	return ""
}

func (a *DPValueIteration) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
	directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	log.Printf("Reading maze from: %v", m.Config().GetFromFile())

	// create local copy of maze running on the server, based on maze_id passed on command line
	// in order to use dp_value_iteration, we need to know the exact model of the environment
	algo := algos.Algorithms["fromfile"]
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	solved := false
	// steps := 0

	for !solved {
		// animation delay
		time.Sleep(delay)

	}

	log.Printf("maze solved!")
	a.ShowStats()

	return nil
}
