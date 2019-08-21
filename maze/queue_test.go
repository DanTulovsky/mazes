package maze

import (
	"container/heap"
	"testing"

	pb "github.com/DanTulovsky/mazes/proto"
)

type testCell struct {
	cell   *Cell
	weight int
}

func cellAt(c *Cell, l Location) bool {
	if c.x != int64(l.X) || c.y != int64(l.Y) || c.z != int64(l.Z) {
		return false
	}
	return true
}

func TestQueue(t *testing.T) {

	var config = &pb.MazeConfig{
		Rows:    10,
		Columns: 15,
	}

	cells := []testCell{
		{NewCell(0, 0, 0, config), 25},
		{NewCell(0, 1, 0, config), 5},
		{NewCell(0, 2, 0, config), 15},
		{NewCell(0, 3, 0, config), 10},
		{NewCell(0, 4, 0, config), 20},
		{NewCell(0, 5, 0, config), 0},
	}

	queue := make(CellPriorityQueue, 0)
	heap.Init(&queue)

	for _, c := range cells {
		heap.Push(&queue, c.cell)
	}

	if queue.Len() != len(cells) {
		t.Fatalf("expected length: %v; received: %v", len(cells), queue.Len())
	}

	h := heap.Pop(&queue).(*Cell)
	if !cellAt(h, Location{0, 0, 0}) {
		t.Fatalf("expected cell [0,0,0], received: %v", h)
	}

}
