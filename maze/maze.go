package maze

import (
	"fmt"
	"log"
	"math/rand"
	"mazes/utils"
	"time"

	"mazes/colors"

	"math"

	"github.com/sasha-s/go-deadlock"
	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano()) // need to initialize the seed
}

// Location is x,y coordinate of a cell
type Location struct {
	X, Y int
}

// Grid defines the maze grid
type Maze struct {
	config           *Config
	rows             int
	columns          int
	cells            [][]*Cell
	mazeCells        map[*Cell]bool // cells that are in the maze, not orphaned (for cachine)
	orphanCells      map[*Cell]bool // cells that are orphaned (for caching)
	cellWidth        int            // cell width
	wallWidth        int
	pathWidth        int
	bgColor          colors.Color
	borderColor      colors.Color
	wallColor        colors.Color
	pathColor        colors.Color
	createTime       time.Duration // how long it took to apply the algorithm to create the grid
	fromCell, toCell *Cell         // save these for proper coloring

	SolvePath  *Path // the final solve path of the solver
	TravelPath *Path // the travel path of the solver, update in real time

	genCurrentLocation *Cell // the current location of generator

	deadlock.RWMutex
}

// NewGrid returns a new grid.
func NewMaze(c *Config) (*Maze, error) {
	if err := c.CheckConfig(); err != nil {
		return nil, err
	}
	m := &Maze{
		rows:        c.Rows,
		columns:     c.Columns,
		cells:       [][]*Cell{},
		cellWidth:   c.CellWidth,
		wallWidth:   c.WallWidth,
		pathWidth:   c.PathWidth,
		bgColor:     c.BgColor,
		borderColor: c.BorderColor,
		wallColor:   c.WallColor,
		pathColor:   c.PathColor,
		config:      c,

		SolvePath:  NewPath(),
		TravelPath: NewPath(),

		mazeCells:   make(map[*Cell]bool),
		orphanCells: make(map[*Cell]bool),
	}

	m.prepareGrid()
	m.configureCells()

	return m, nil
}

// prepareGrid initializes the grid with cells
func (m *Maze) prepareGrid() {
	m.cells = make([][]*Cell, m.columns)

	for x := 0; x < m.columns; x++ {
		m.cells[x] = make([]*Cell, m.rows)

		for y := 0; y < m.rows; y++ {
			m.cells[x][y] = NewCell(x, y, m.config)
		}
	}
}

// configureCells configures cells with their neighbors
func (m *Maze) configureCells() {
	for x := 0; x < m.columns; x++ {
		for y := 0; y < m.rows; y++ {
			cell, err := m.Cell(x, y)
			if err != nil {
				log.Fatalf("failed to initialize grid: %v", err)
			}
			// error is ignored, we just set nil if there is no neighbor
			cell.North, _ = m.Cell(x, y-1)
			cell.South, _ = m.Cell(x, y+1)
			cell.West, _ = m.Cell(x-1, y)
			cell.East, _ = m.Cell(x+1, y)
		}
	}

	for _, o := range m.config.OrphanMask {
		cell, err := m.Cell(o.X, o.Y)
		if err != nil {
			Fail(err)
		}
		cell.Orphan()
	}

}

// SetCurrentLocation sets the current cell location of the generator algorithm
func (m *Maze) SetGenCurrentLocation(cell *Cell) {
	m.Lock()
	defer m.Unlock()
	m.genCurrentLocation = cell
}

// GenCurrentLocation returns the current cell location of the generator algorithm
func (m *Maze) GenCurrentLocation() *Cell {
	m.RLock()
	defer m.RUnlock()
	return m.genCurrentLocation
}

func (m *Maze) SetCreateTime(t time.Duration) {
	m.Lock()
	defer m.Unlock()
	m.createTime = t
}

func (m *Maze) CreateTime() time.Duration {
	m.RLock()
	defer m.RUnlock()
	return m.createTime
}

// Dimensions returns the dimensions of the grid.
func (m *Maze) Dimensions() (int, int) {
	m.RLock()
	defer m.RUnlock()
	return m.columns, m.rows
}

func (m *Maze) String() string {
	m.RLock()
	defer m.RUnlock()

	output := "  "
	for x := 0; x < m.columns; x++ {
		output = fmt.Sprintf("%v%4v", output, x)
	}

	output = fmt.Sprintf("\n%v\n   ┌", output)
	for x := 0; x < m.columns-1; x++ {
		output = fmt.Sprintf("%v───┬", output)
	}
	output = output + "───┐" + "\n"

	for y := 0; y < m.rows; y++ {
		top := fmt.Sprintf("%-3v│", y)
		bottom := "   ├"

		for x := 0; x < m.columns; x++ {
			cell, err := m.Cell(x, y)
			if err != nil {
				continue
			}
			body := "   "
			east_boundary := " "
			if !cell.Linked(cell.East) {
				east_boundary = "│"
			}
			top = fmt.Sprintf("%v%v%v", top, body, east_boundary)

			south_boundary := "   "
			if !cell.Linked(cell.South) {
				south_boundary = "───"
			}
			corner := "┼"
			if x == m.columns-1 {
				corner = "┤" // right wall
			}
			if x == m.columns-1 && y == m.rows-1 {
				corner = "┘"
			}
			if x == 0 && y == m.rows-1 {
				bottom = "   └"
			}
			if x < m.columns-1 && y == m.rows-1 {
				corner = "┴"
			}
			bottom = fmt.Sprintf("%v%v%v", bottom, south_boundary, corner)
		}
		output = fmt.Sprintf("%v%v\n", output, top)
		output = fmt.Sprintf("%v%v\n", output, bottom)
	}

	return output
}

// Draw renders the gui maze in memory, display by calling Present
func (m *Maze) DrawMaze(r *sdl.Renderer) *sdl.Renderer {
	// utils.TimeTrack(time.Now(), "DrawMaze")

	// If saved, draw distance colors
	if m.fromCell != nil {
		m.SetDistanceColors(m.fromCell)
	}
	if m.fromCell != nil && m.toCell != nil {
		m.SetFromToColors(m.fromCell, m.toCell)
	}

	// Each cell draws its background, half the wall as well as anything inside it
	for x := 0; x < m.columns; x++ {
		for y := 0; y < m.rows; y++ {
			cell, err := m.Cell(x, y)
			if err != nil {
				Fail(fmt.Errorf("Error drawing cell (%v, %v): %v", x, y, err))
			}

			if cell.IsOrphan() {
				// these are cells not connected to the maze
				continue
			}

			if cell.config.DarkMode && !cell.Visited() {
				// in dark mode don't draw unvisited cells
				continue
			}
			cell.Draw(r)
		}
	}

	// Draw outside border
	m.drawBorder(r)

	// Draw location of the generator algorithm
	m.drawGenCurrentLocation(r)

	// Draw the path and location of solver
	m.drawPath(r, m.TravelPath, m.config.MarkVisitedCells)

	return r
}

// DrawBorder renders the maze border in memory, display by calling Present
func (m *Maze) drawBorder(r *sdl.Renderer) *sdl.Renderer {
	colors.SetDrawColor(m.borderColor, r)

	var bg sdl.Rect
	var rects []sdl.Rect
	winWidth := int32(m.columns*m.cellWidth + m.wallWidth*2)
	winHeight := int32(m.rows*m.cellWidth + m.wallWidth*2)
	wallWidth := int32(m.wallWidth)

	// top
	bg = sdl.Rect{0, 0, winWidth, wallWidth}
	rects = append(rects, bg)

	// left
	bg = sdl.Rect{0, 0, wallWidth, winHeight}
	rects = append(rects, bg)

	// bottom
	bg = sdl.Rect{0, winHeight - wallWidth, winWidth, wallWidth}
	rects = append(rects, bg)

	// right
	bg = sdl.Rect{winWidth - wallWidth, 0, wallWidth, winHeight}
	rects = append(rects, bg)

	if err := r.FillRects(rects); err != nil {
		Fail(fmt.Errorf("error drawing border: %v", err))
	}
	return r
}

func (m *Maze) drawGenCurrentLocation(r *sdl.Renderer) *sdl.Renderer {

	current_location := m.GenCurrentLocation()

	if current_location != nil {
		for cell := range m.Cells() {
			if cell != nil {
				// reset all colors to default
				cell.SetBGColor(colors.GetColor("white"))
			}
		}

		current_location.SetBGColor(colors.GetColor("yellow"))
	}
	return r
}

// DrawPath renders the gui maze path in memory, display by calling Present
// This is drawing g.TravelPath if path == nil
func (m *Maze) drawPath(r *sdl.Renderer, path *Path, markVisited bool) *sdl.Renderer {
	// utils.TimeTrack(time.Now(), "drawPath")
	if path == nil {
		path = m.TravelPath
	}

	alreadyDone := make(map[*PathSegment]bool)

	var isSolution bool
	var isLast bool
	pathLength := len(path.segments)
	solvepathCells := m.SolvePath.ListCells()

	for x, segment := range path.segments {
		cell := segment.Cell()

		if x == pathLength-1 {
			isLast = true // last segment is drawn slightly different
		}

		if isLast {
			cell.DrawCurrentLocation(r)
		}

		if _, ok := alreadyDone[segment]; ok {
			continue
		}

		// cache state of this cell
		alreadyDone[segment] = true

		if SegmentInPath(segment, m.SolvePath) {
			isSolution = true
		} else {
			isSolution = false
		}

		segment.DrawPath(r, m, solvepathCells, isLast, isSolution) // solution is colored by a different color

		if markVisited {
			cell.DrawVisited(r)
		}

	}

	return r
}

// DrawVisited renders the gui maze visited dots in memory, display by calling Present
// This function draws the entire path at once
//func (m *Maze) DrawVisited(r *sdl.Renderer) *sdl.Renderer {
//	m.RLock()
//	defer m.RUnlock()
//
//	for x := 0; x < m.columns; x++ {
//		for y := 0; y < m.rows; y++ {
//			cell, err := m.Cell(x, y)
//			if err != nil {
//				Fail(fmt.Errorf("Error drawing cell (%v, %v): %v", x, y, err))
//			}
//			cell.DrawVisited(r)
//		}
//	}
//
//	return r
//}

// Cell returns the cell at r,c
func (m *Maze) Cell(x, y int) (*Cell, error) {
	if x < 0 || x >= m.columns || y < 0 || y >= m.rows {
		return nil, fmt.Errorf("(%v, %v) is outside the grid", x, y)
	}
	return m.cells[x][y], nil
}

func cellMapKeys(m map[*Cell]bool) []*Cell {
	var keys []*Cell
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// RandomCell returns a random cell out of all non-orphaned cells
func (m *Maze) RandomCell() *Cell {
	cells := cellMapKeys(m.Cells())

	return cells[utils.Random(0, len(cells))]
}

// RandomCellFromList returns a random cell from the provided list of cells
func (g *Maze) RandomCellFromList(cells []*Cell) *Cell {
	return cells[utils.Random(0, len(cells))]
}

// Size returns the number of cells in the grid
func (m *Maze) Size() int {
	m.RLock()
	defer m.RUnlock()

	return m.columns * m.rows
}

// Rows returns a list of rows (essentially the grid) - excluding the orphaned cells
func (m *Maze) Rows() [][]*Cell {
	m.RLock()
	defer m.RUnlock()

	rows := [][]*Cell{}

	for y := m.rows - 1; y >= 0; y-- {
		cells := []*Cell{}
		for x := m.columns - 1; x >= 0; x-- {
			cell, _ := m.Cell(x, y)
			if !cell.IsOrphan() {
				cells = append(cells, cell)
			}
		}
		rows = append(rows, cells)
	}
	return rows
}

// Cells returns a list of un-orphanded cells in the grid
func (m *Maze) Cells() map[*Cell]bool {

	mazeCells := m.MazeCells()

	if len(mazeCells) != 0 {
		return mazeCells
	}

	cells := make(map[*Cell]bool)
	for y := m.rows - 1; y >= 0; y-- {
		for x := m.columns - 1; x >= 0; x-- {
			cell := m.cells[x][y]
			if !cell.IsOrphan() {
				cells[cell] = true
			}
		}
	}

	// cache
	m.setMazeCells(cells)
	return cells
}

func (m *Maze) MazeCells() map[*Cell]bool {
	m.RLock()
	defer m.RUnlock()
	return m.mazeCells
}

func (m *Maze) setMazeCells(cells map[*Cell]bool) {
	m.Lock()
	defer m.Unlock()
	m.mazeCells = cells
}

// OrphanCells returns a list of orphan cells in the grid
func (m *Maze) OrphanCells() map[*Cell]bool {
	m.RLock()
	defer m.RUnlock()

	if m.orphanCells != nil {
		return m.orphanCells
	}

	cells := make(map[*Cell]bool)
	for y := 0; y < m.rows; y++ {
		for x := 0; x < m.columns; x++ {
			cell := m.cells[x][y]
			if cell.IsOrphan() {
				cells[cell] = true
			}
		}
	}

	m.setOrphanMazeCells(cells)
	return cells
}

func (m *Maze) setOrphanMazeCells(cells map[*Cell]bool) {
	m.Lock()
	defer m.Unlock()
	m.orphanCells = cells
}

// UnvisitedCells returns a list of unvisited cells in the grid
func (m *Maze) UnvisitedCells() []*Cell {
	cells := []*Cell{}

	for cell := range m.Cells() {
		if !cell.Visited() {
			cells = append(cells, cell)
		}
	}

	return cells
}

// ConnectCells connects the list of cells in order by passageways
func (m *Maze) ConnectCells(cells []*Cell) {

	for x := 0; x < len(cells)-1; x++ {
		cell := cells[x]
		for _, n := range []*Cell{cell.North, cell.South, cell.East, cell.West} {
			if n == cells[x+1] {
				cell.Link(n)
				break
			}
		}
	}
}

// LongestPath returns the longest path through the maze
func (m *Maze) LongestPath() (dist int, fromCell, toCell *Cell, path *Path) {
	// pick random starting point
	fromCell = m.RandomCell()

	// find furthest point
	furthest, _ := fromCell.FurthestCell()

	// now find the furthest point from that
	toCell, _ = furthest.FurthestCell()

	// now get the path
	dist, path = m.ShortestPath(furthest, toCell)

	return dist, furthest, toCell, path
}

// SetFromToColors sets the opacity of the from and to cells to be highly visible
func (m *Maze) SetFromToColors(fromCell, toCell *Cell) {
	m.Lock()
	defer m.Unlock()

	// Set path start and end colors
	fromCell.SetBGColor(colors.SetOpacity(fromCell.bgColor, 0))
	toCell.SetBGColor(colors.SetOpacity(toCell.bgColor, 255))

	// save these for color refresh.
	m.fromCell = fromCell
	m.toCell = toCell
}

// SetPathFromTo draws the given path from fromCell to toCell
func (m *Maze) SetPathFromTo(fromCell, toCell *Cell, path *Path) {

	var prev, next *Cell
	for x := 0; x < path.Length(); x++ {
		if x > 0 {
			prev = path.segments[x-1].Cell()
		}

		if x < path.Length()-1 {
			next = path.segments[x+1].Cell()
		}

		path.segments[x].Cell().SetPaths(prev, next)
	}
}

// ShortestPath finds the shortest path from fromCell to toCell
func (m *Maze) ShortestPath(fromCell, toCell *Cell) (int, *Path) {
	if path := fromCell.PathTo(toCell); path != nil {
		return path.Length(), path
	}

	var path = NewPath()

	// Get all distances from this cell
	d := fromCell.Distances()
	toCellDist, _ := d.Get(toCell)

	current := toCell

	for current != d.root {
		smallest := math.MaxInt64
		var next *Cell
		for _, link := range current.Links() {
			dist, err := d.Get(link)
			if err != nil {
				continue
			}
			if dist < smallest {
				smallest = dist
				next = link
			}
		}
		segment := NewSegment(next, "north") // arbitrary facing
		path.AddSegement(segment)
		if next == nil {
			log.Fatalf("failed to find next cell from: %v", current)
		}
		current = next
	}

	// add toCell to path
	path.ReverseCells()
	segment := NewSegment(toCell, "north") // arbitrary facing
	path.AddSegement(segment)

	// record path for caching
	fromCell.SetPathTo(toCell, path)

	return toCellDist, path
}

// SetDistanceColors colors the graph based on distances from c
func (m *Maze) SetDistanceColors(c *Cell) {
	// figure out the distances if needed
	c.Distances()

	_, longestPath := c.FurthestCell()

	// use alpha blending, works for any color
	for cell := range m.Cells() {
		d, err := c.distances.Get(cell)
		if err != nil {
			log.Printf("failed to get distance from %v to %v", c, cell)
			return
		}
		// decrease the last parameter to make the longest cells brighter. max = 255 (good = 228)
		adjustedColor := utils.AffineTransform(float32(d), 0, float32(longestPath), 0, 228)
		cell.SetBGColor(colors.OpacityAdjust(m.bgColor, adjustedColor))
	}

	m.Lock()
	defer m.Unlock()

	m.fromCell = c
}

// DeadEnds returns a list of cells that are deadends (only linked to one neighbor
func (m *Maze) DeadEnds() []*Cell {
	var deadends []*Cell

	for cell := range m.Cells() {
		if len(cell.Links()) == 1 {
			deadends = append(deadends, cell)
		}
	}

	return deadends
}

// Reset resets vital maze stats for a new solver run
func (m *Maze) Reset() {
	m.TravelPath = NewPath()
	m.SolvePath = NewPath()
	m.ResetVisited()

}

// ResetVisited sets all cells to be unvisited
func (m *Maze) ResetVisited() {
	for c := range m.Cells() {
		c.SetUnVisited()
	}

}

// GetFacingDirection returns the direction walker was facing when moving fromCell -> toCell
// north, south, east, west
func (m *Maze) GetFacingDirection(fromCell, toCell *Cell) string {
	facing := ""

	if fromCell.North == toCell {
		facing = "north"
	}
	if fromCell.East == toCell {
		facing = "east"
	}
	if fromCell.West == toCell {
		facing = "west"
	}
	if fromCell.South == toCell {
		facing = "south"
	}
	return facing
}
