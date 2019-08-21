package manual

import (
	"log"
	"time"

	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/solvealgos"

	pb "github.com/DanTulovsky/mazes/proto"

	"github.com/nsf/termbox-go"
)

type Manual struct {
	solvealgos.Common
}

func getNextCell(key termbox.Key) string {
	dirMap := map[termbox.Key]string{
		termbox.KeyArrowUp:    "north",
		termbox.KeyArrowDown:  "south",
		termbox.KeyArrowLeft:  "west",
		termbox.KeyArrowRight: "east",
	}

	return dirMap[key]
}

func (a *Manual) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
	directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	log.Print("Solver is human...")

	currentCell := fromCell
	solved := false
	steps := 0

	for !solved {
		// get nextCell from user input based on key press

		ev := termbox.PollEvent()
		if ev.Key == termbox.KeyCtrlC {
			return nil
		}
		direction := getNextCell(ev.Key)

		reply, err := a.Move(mazeID, clientID, direction)
		if err != nil {
			continue
		}

		directions = reply.GetAvailableDirections()
		previousCell := currentCell
		currentCell = reply.GetCurrentLocation()

		// set current location in local maze
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
