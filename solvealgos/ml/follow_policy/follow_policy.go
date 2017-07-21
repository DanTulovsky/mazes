package ml_follow_policy

import (
	"fmt"
	"log"
	"mazes/maze"
	"mazes/ml"
	"mazes/solvealgos"
	"mazes/utils"
	"time"

	pb "mazes/proto"
)

type MLFollowPolicy struct {
	solvealgos.Common
}

func (a *MLFollowPolicy) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
	directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	if a.Policy() == nil {
		return fmt.Errorf("no policy available for solution")
	}

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
		action := a.Policy().BestRandomActionsForState(state)
		// log.Printf("At: %v (state=%v); moving to: %v", loc, state, dp.ActionToText[action])

		reply, err := a.Move(mazeID, clientID, ml.ActionToText[action])
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
