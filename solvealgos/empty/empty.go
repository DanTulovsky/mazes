// empty is a place holder algorithm that does nothing
package empty

import (
	"fmt"
	"log"
	"time"

	"mazes/maze"
	pb "mazes/proto"
	"mazes/solvealgos"
)

type Empty struct {
	solvealgos.Common
}

// directions is the initial available directions to travel
func (a *Empty) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
	directions []*pb.Direction, m *maze.Maze) error {

	log.Printf("fromCell: %v; toCell: %v; directions: %v", fromCell, toCell, directions)
	if len(directions) < 1 {
		return fmt.Errorf("no available directions to move: %v", directions)
	}

	solved := false

	for !solved {
		reply, err := a.Move(mazeID, clientID, directions[0].GetName())
		if err != nil {
			return err
		}
		solved = reply.Solved

		directions = reply.GetAvailableDirections()
		if len(directions) < 1 {
			return fmt.Errorf("no available directions to move now: %v", directions)
		}
	}
	return nil
}
