package prim

import (
	"container/heap"
	"log"
	"mazes/genalgos"
	"mazes/maze"
	"mazes/utils"
	"time"
)

type Prim struct {
	genalgos.Common
}

func (a *Prim) Apply(m *maze.Maze, delay time.Duration) error {
	defer genalgos.TimeTrack(m, time.Now())

	// Setup costs for all cells
	for c := range m.Cells() {
		w := utils.Random(0, m.Size()*100)
		c.SetWeight(w)
	}

	active := make(CellPriorityQueue, 0)
	heap.Init(&active)

	// random start
	startCell := m.RandomCell()
	heap.Push(&active, startCell)

	// while we have unprocessed cells
	for active.Len() > 0 {
		time.Sleep(delay) // animation delay

		// grab a cell with the lowest weight
		cell := active[0]

		neighborsQueue := make(CellPriorityQueue, 0)
		heap.Init(&neighborsQueue)

		neighbors := cell.UnLinked()
		if len(neighbors) == 0 {
			// no more neighbors, remove
			popped := heap.Pop(&active)
			if cell != popped {
				log.Fatalf("popped (%v) and top (%v) not the same", popped, cell)
			}
		} else {
			for _, n := range neighbors {
				heap.Push(&neighborsQueue, n)
			}

			n := heap.Pop(&neighborsQueue).(*maze.Cell)
			m.Link(cell, n)
			heap.Push(&active, n)
		}

	}

	a.Cleanup(m)
	return nil
}
