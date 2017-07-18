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
	"mazes/utils"
)

type DPValueIteration struct {
	solvealgos.Common
}

func (a *DPValueIteration) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
	directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	log.Printf("Reading maze from: %v", m.Config().GetFromFile())
	log.Printf("in solver, fromCell: %v; toCell: %v", fromCell, toCell)

	// create local copy of maze running on the server, based on maze_id passed on command line
	// in order to use dp_value_iteration, we need to know the exact model of the environment
	algo := &fromfile.Fromfile{}
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	df := 1.0
	theta := 0.000000001
	log.Printf("Determining optimal policy...")
	policy, vf, err := dp.ValueIteration(m, clientID, df, theta, dp.DefaultActions)
	//policy, vf, err := dp.PolicyImprovement(m, clientID, df, theta, dp.DefaultActions)
	if err != nil {
		return fmt.Errorf("error calculating optimal policy: %v", err)
	}
	log.Printf("value function:\n%v", vf.Reshape(int(m.Config().Rows), int(m.Config().Columns)))
	// log.Printf("optimal policy:\n%v", policy)
	solved := false
	steps := 0

	currentCell := fromCell

	for !solved {
		// animation delay
		time.Sleep(delay)

		loc := &pb.MazeLocation{X: currentCell.GetX(), Y: currentCell.GetY(), Z: currentCell.GetZ()}
		state, err := utils.StateFromLocation(m.Config().Rows, m.Config().Columns, loc)
		if err != nil {
			return fmt.Errorf("error converting [%v] to location: %v", state, err)
		}
		// not random because only one action is 1, the rest 0
		action := policy.BestRandomActionsForState(state)
		// log.Printf("At: %v (state=%v); moving to: %v", loc, state, dp.ActionToText[action])

		reply, err := a.Move(mazeID, clientID, dp.ActionToText[action])
		if err != nil {
			return err
		}

		previousCell := currentCell
		currentCell = reply.GetCurrentLocation()
		// availableDirections := reply.GetAvailableDirections()
		steps++

		if err := a.UpdateClientViewAndLocation(clientID, m, currentCell, previousCell, steps); err != nil {
			return err
		}

		solved = reply.Solved
	}

	log.Printf("maze solved in %v steps!", steps)
	a.ShowStats()

	return nil
}
