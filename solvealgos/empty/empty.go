// empty is a place holder algorithm that does nothing
package empty

import (
	"fmt"
	"log"
	"time"

	pb "mazes/proto"
	"mazes/solvealgos"
)

type Empty struct {
	solvealgos.Common
}

// directions is the initial available directions to travel
func (a *Empty) Solve(stream pb.Mazer_SolveMazeClient, mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration, directions []string) error {

	log.Printf("fromCell: %v; toCell: %v; directions: %v", fromCell, toCell, directions)
	if len(directions) < 1 {
		return fmt.Errorf("no available directions to move: %v", directions)
	}

	currentCell := fromCell

	for currentCell != toCell {

		r := &pb.SolveMazeRequest{
			MazeId:    mazeID,
			ClientId:  clientID,
			Direction: directions[0],
		}

		// send move request to server
		log.Printf("sending move request to server: %v", r)
		if err := stream.Send(r); err != nil {
			return err
		}
		log.Printf("sent")

		// get response
		log.Printf("waiting for move reply from server")
		reply, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Printf("received: %v", r)

		directions = reply.GetAvailableDirections()
		log.Printf("i am at: %v and can go: %v", reply.GetCurrentLocation(), reply.GetAvailableDirections())
		if len(directions) < 1 {
			return fmt.Errorf("no available directions to move now: %v", directions)
		}
	}
	return nil
}
