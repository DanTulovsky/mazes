package kruskal

// CellPriorityQueue implements a reverse priority queue for cells by cell weight, there is no update method
type NeighborPairPriorityQueue []*neighborPair

func (pq NeighborPairPriorityQueue) Len() int { return len(pq) }

func (pq NeighborPairPriorityQueue) Less(i, j int) bool {
	return pq[i].Weight() < pq[j].Weight()
}

func (pq NeighborPairPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *NeighborPairPriorityQueue) Push(x interface{}) {
	item := x
	*pq = append(*pq, item.(*neighborPair))
}

func (pq *NeighborPairPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
