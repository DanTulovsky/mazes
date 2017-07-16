package dp_value_iteration

import (
	"log"
	"mazes/maze"
	"mazes/solvealgos"
	"time"

	"fmt"
	"mazes/genalgos/fromfile"
	"mazes/ml/dp"
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
	algo := &fromfile.Fromfile{}
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	df := 0.999
	theta := 0.00001
	policy, vf, err := dp.ValueIteration(m, clientID, df, theta, dp.DefaultActions)
	if err != nil {
		return fmt.Errorf("error calculating optimal policy: %v", err)
	}
	log.Printf("value function:\n%v", vf.Reshape(int(m.Config().Rows), int(m.Config().Columns)))
	log.Printf("optimal policy:\n%v", policy)
	solved := false
	// steps := 0

	for !solved {
		solved = true
		// animation delay
		time.Sleep(delay)

	}

	log.Printf("maze solved!")
	a.ShowStats()

	return nil
}
