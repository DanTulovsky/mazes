package prim

import "github.com/DanTulovsky/mazes/maze"

// CellPriorityQueue implements a priority queue for cells by cell weight, there is no update method
type CellPriorityQueue []*maze.Cell

func (pq CellPriorityQueue) Len() int { return len(pq) }

func (pq CellPriorityQueue) Less(i, j int) bool {
	return pq[i].Weight() < pq[j].Weight()
}

func (pq CellPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *CellPriorityQueue) Push(x interface{}) {
	item := x
	*pq = append(*pq, item.(*maze.Cell))
}

func (pq *CellPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
