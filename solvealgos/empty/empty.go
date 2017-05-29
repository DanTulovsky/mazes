// empty is a place holder algorithm that does nothing
package empty

import (
	"time"

	pb "mazes/proto"
	"mazes/solvealgos"
)

type Empty struct {
	solvealgos.Common
}

func (a *Empty) Solve(stream pb.Mazer_SolveMazeClient, fromCell, toCell string, delay time.Duration) error {

	return nil
}
