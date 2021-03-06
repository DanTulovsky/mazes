// ml_follow_policy implements solving a maze by following a deterministic policy
package ml_follow_policy

import (
	"fmt"
	"log"
	"time"

	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/ml"
	"github.com/DanTulovsky/mazes/solvealgos"
	"github.com/DanTulovsky/mazes/utils"

	pb "github.com/DanTulovsky/mazes/proto"
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
	var action int

	availableDirections := directions

	for !solved {
		// animation delay
		time.Sleep(delay)

		loc := &pb.MazeLocation{X: currentCell.GetX(), Y: currentCell.GetY(), Z: currentCell.GetZ()}
		state, err := utils.StateFromLocation(m.Config().Rows, m.Config().Columns, loc)
		if err != nil {
			return fmt.Errorf("error converting [%v] to location: %v", state, err)
		}

		// Pick out of the valid directions only.  This resolves the problem of an unvisited state
		// that still has the same value of all actions
		action = a.Policy().BestValidDeterministicActionForState(state, availableDirections)

		// log.Printf("At: %v (state=%v); moving to: %v", loc, state, dp.ActionToText[action])

		reply, err := a.Move(mazeID, clientID, ml.ActionToText[action])
		if err != nil {
			return err
		}

		previousCell := currentCell
		currentCell = reply.GetCurrentLocation()
		availableDirections = reply.GetAvailableDirections()
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
