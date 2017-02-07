// Package wilsons implements wilson's algorithm for maze generation

// The algorithm starts by choosing a point on the grid—any point—and marking it visited. Then it
// chooses any unvisited cell in the grid and does one of these loop-erased random walks until it
// encounters a visited cell. At that point it adds the path it followed to the maze, marking as visited
// each of the cells along that path, and then it goes again. The process repeats until all the cells in
// the grid have been visited.
package wilsons

import (
	"mazes/genalgos"
	"mazes/maze"
	"mazes/utils"
	"time"
)

type Wilsons struct {
	genalgos.Common
}

// Apply applies wilson's algorithm to generate the maze.
func (a *Wilsons) Apply(g *maze.Maze, delay time.Duration) (*maze.Maze, error) {

	defer genalgos.TimeTrack(g, time.Now())

	var currentCell *maze.Cell
	var randomCell *maze.Cell
	var walkPath []*maze.Cell
	var visitedCells []*maze.Cell

	start := g.RandomCell()
	start.SetVisited()
	visitedCells = append(visitedCells, start)

	for len(g.UnvisitedCells()) > 0 {
		time.Sleep(delay) // animation delay

		// pick random, unvisited cell
		randomCell = g.RandomCellFromList(g.UnvisitedCells())
		currentCell = randomCell
		g.SetGenCurrentLocation(currentCell)

		// walk until you hit a visited cell
		for !maze.CellInCellList(currentCell, visitedCells) {
			time.Sleep(delay) // animation delay

			// handle loop
			if maze.CellInCellList(currentCell, walkPath) {
				i := utils.SliceIndex(len(walkPath), func(i int) bool { return walkPath[i] == currentCell })
				walkPath = walkPath[0:i]
			}

			// add it to walkPath
			walkPath = append(walkPath, currentCell)

			// visit random neighbors until you come to a visited cell
			currentCell = currentCell.RandomNeighbor()

			g.SetGenCurrentLocation(currentCell)

		}

		// add path to the maze
		walkPath = append(walkPath, currentCell)
		g.ConnectCells(walkPath)
		for _, c := range walkPath {
			visitedCells = append(visitedCells, c)
			c.SetVisited()
		}

		// clear path
		walkPath = []*maze.Cell{}
	}

	a.Cleanup(g)
	return g, nil
}
