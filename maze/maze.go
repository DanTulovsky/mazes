package maze

import (
	"fmt"
	"log"
	"math/rand"
	"mazes/utils"
	"time"

	"mazes/colors"

	"math"

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
	mazeCells        []*Cell // cells that are in the maze, not orphaned (for cachine)
	orphanCells      []*Cell // cells that are orphaned (for caching)
	cellWidth        int     // cell width
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
}

// NewGrid returns a new grid.
func NewGrid(c *Config) (*Maze, error) {
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

func (m *Maze) SetGenCurrentLocation(cell *Cell) {
	m.genCurrentLocation = cell
}

func (g *Maze) SetCreateTime(t time.Duration) {
	g.createTime = t
}

func (g *Maze) CreateTime() time.Duration {
	return g.createTime
}

// Dimensions returns the dimensions of the grid.
func (g *Maze) Dimensions() (int, int) {

	return g.columns, g.rows
}

func (g *Maze) String() string {
	output := "  "
	for x := 0; x < g.columns; x++ {
		output = fmt.Sprintf("%v%4v", output, x)
	}

	output = fmt.Sprintf("\n%v\n   ┌", output)
	for x := 0; x < g.columns-1; x++ {
		output = fmt.Sprintf("%v───┬", output)
	}
	output = output + "───┐" + "\n"

	for y := 0; y < g.rows; y++ {
		top := fmt.Sprintf("%-3v│", y)
		bottom := "   ├"

		for x := 0; x < g.columns; x++ {
			cell, err := g.Cell(x, y)
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
			if x == g.columns-1 {
				corner = "┤" // right wall
			}
			if x == g.columns-1 && y == g.rows-1 {
				corner = "┘"
			}
			if x == 0 && y == g.rows-1 {
				bottom = "   └"
			}
			if x < g.columns-1 && y == g.rows-1 {
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
func (g *Maze) DrawMaze(r *sdl.Renderer) *sdl.Renderer {
	// utils.TimeTrack(time.Now(), "DrawMaze")

	// If saved, draw distance colors
	if g.fromCell != nil {
		g.SetDistanceColors(g.fromCell)
	}
	if g.fromCell != nil && g.toCell != nil {
		g.SetFromToColors(g.fromCell, g.toCell)
	}

	// Each cell draws its background, half the wall as well as anything inside it
	for x := 0; x < g.columns; x++ {
		for y := 0; y < g.rows; y++ {
			cell, err := g.Cell(x, y)
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
	g.DrawBorder(r)

	// Draw location of the generator algorithm
	g.DrawGenCurrentLocation(r)

	// Draw the path
	g.DrawPath(r, g.TravelPath, g.config.MarkVisitedCells)

	// Draw the location of the solver algorithm
	g.DrawSolveCurrentLocation(r)

	return r
}

// DrawBorder renders the maze border in memory, display by calling Present
func (g *Maze) DrawBorder(r *sdl.Renderer) *sdl.Renderer {
	colors.SetDrawColor(g.borderColor, r)

	var bg sdl.Rect
	var rects []sdl.Rect
	winWidth := int32(g.columns*g.cellWidth + g.wallWidth*2)
	winHeight := int32(g.rows*g.cellWidth + g.wallWidth*2)
	wallWidth := int32(g.wallWidth)

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

func (g *Maze) DrawSolveCurrentLocation(r *sdl.Renderer) {

	if g.TravelPath == nil {
		return
	}

	segment := g.TravelPath.LastSegment()
	if segment != nil {
		cell := segment.Cell()
		if cell != nil {
			cell.DrawCurrentLocation(r)
		}
	}
}

func (g *Maze) DrawGenCurrentLocation(r *sdl.Renderer) *sdl.Renderer {
	if g.genCurrentLocation != nil {
		for _, cell := range g.Cells() {
			// reset all colors to default
			cell.bgColor = colors.GetColor("white")
		}

		g.genCurrentLocation.bgColor = colors.GetColor("yellow")
	}
	return r
}

// DrawPath renders the gui maze path in memory, display by calling Present
// This is drawing g.TravelPath if path == nil
func (g *Maze) DrawPath(r *sdl.Renderer, path *Path, markVisited bool) *sdl.Renderer {
	if path == nil {
		path = g.TravelPath
	}

	alreadyDone := make(map[*Cell]bool)

	var isSolution bool
	var isLast bool

	for x, segment := range path.segments {
		if _, ok := alreadyDone[segment.Cell()]; ok {
			continue
		}

		// cache state of this cell
		alreadyDone[segment.Cell()] = true

		if x == len(path.segments)-1 {
			isLast = true // last segment is drawn slightly different
		}

		if SegmentInPath(segment, g.SolvePath) {
			isSolution = true
		} else {
			isSolution = false
		}

		segment.DrawPath(r, g, isLast, isSolution) // solution is colored by a different color

		if markVisited {
			segment.Cell().DrawVisited(r)
		}

	}

	return r
}

// DrawVisited renders the gui maze visited dots in memory, display by calling Present
// This function draws the entire path at once
func (g *Maze) DrawVisited(r *sdl.Renderer) *sdl.Renderer {
	for x := 0; x < g.columns; x++ {
		for y := 0; y < g.rows; y++ {
			cell, err := g.Cell(x, y)
			if err != nil {
				Fail(fmt.Errorf("Error drawing cell (%v, %v): %v", x, y, err))
			}
			cell.DrawVisited(r)
		}
	}

	return r
}

// Cell returns the cell at r,c
func (g *Maze) Cell(x, y int) (*Cell, error) {
	if x < 0 || x >= g.columns || y < 0 || y >= g.rows {
		return nil, fmt.Errorf("(%v, %v) is outside the grid", x, y)
	}
	return g.cells[x][y], nil
}

// RandomCell returns a random cell out of all non-orphaned cells
func (g *Maze) RandomCell() *Cell {
	cells := g.Cells()
	return cells[utils.Random(0, len(cells))]
}

// RandomCellFromList returns a random cell from the provided list of cells
func (g *Maze) RandomCellFromList(cells []*Cell) *Cell {
	return cells[utils.Random(0, len(cells))]
}

// Size returns the number of cells in the grid
func (g *Maze) Size() int {
	return g.columns * g.rows
}

// Rows returns a list of rows (essentially the grid) - excluding the orphaned cells
func (g *Maze) Rows() [][]*Cell {
	rows := [][]*Cell{}

	for y := g.rows - 1; y >= 0; y-- {
		cells := []*Cell{}
		for x := g.columns - 1; x >= 0; x-- {
			cell, _ := g.Cell(x, y)
			if !cell.IsOrphan() {
				cells = append(cells, cell)
			}
		}
		rows = append(rows, cells)
	}
	return rows
}

// Cells returns a list of un-orphanded cells in the grid
func (g *Maze) Cells() []*Cell {
	if g.mazeCells != nil {
		return g.mazeCells
	}
	cells := []*Cell{}
	for y := g.rows - 1; y >= 0; y-- {
		for x := g.columns - 1; x >= 0; x-- {
			cell := g.cells[x][y]
			if !cell.IsOrphan() {
				cells = append(cells, cell)
			}
		}
	}

	// cache
	g.mazeCells = cells
	return cells
}

// OrphanCells returns a list of orphan cells in the grid
func (g *Maze) OrphanCells() []*Cell {
	if g.orphanCells != nil {
		return g.orphanCells
	}

	cells := []*Cell{}
	for y := 0; y < g.rows; y++ {
		for x := 0; x < g.columns; x++ {
			cell := g.cells[x][y]
			if cell.IsOrphan() {
				cells = append(cells, cell)
			}
		}
	}

	g.orphanCells = cells
	return cells
}

// UnvisitedCells returns a list of unvisited cells in the grid
func (g *Maze) UnvisitedCells() []*Cell {
	cells := []*Cell{}

	for _, cell := range g.Cells() {
		if !cell.Visited() {
			cells = append(cells, cell)
		}
	}

	return cells
}

// ConnectCells connects the list of cells in order by passageways
func (g *Maze) ConnectCells(cells []*Cell) {

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
func (g *Maze) LongestPath() (dist int, fromCell, toCell *Cell, path *Path) {
	// pick random starting point
	fromCell = g.RandomCell()

	// find furthest point
	furthest, _ := fromCell.FurthestCell()

	// now find the furthest point from that
	toCell, _ = furthest.FurthestCell()

	// now get the path
	dist, path = g.ShortestPath(furthest, toCell)

	return dist, furthest, toCell, path
}

func (p *Path) ReverseCells() {
	for i, j := 0, len(p.segments)-1; i < j; i, j = i+1, j-1 {
		p.segments[i], p.segments[j] = p.segments[j], p.segments[i]
	}
}

// SetFromToColors sets the opacity of the from and to cells to be highly visible
func (g *Maze) SetFromToColors(fromCell, toCell *Cell) {
	// Set path start and end colors
	fromCell.bgColor = colors.SetOpacity(fromCell.bgColor, 0)
	toCell.bgColor = colors.SetOpacity(toCell.bgColor, 255)

	// save these for color refresh.
	g.fromCell = fromCell
	g.toCell = toCell
}

// SetPathFromTo draws the given path from fromCell to toCell
func (g *Maze) SetPathFromTo(fromCell, toCell *Cell, path *Path) {
	// g.SetFromToColors(fromCell, toCell)

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
func (g *Maze) ShortestPath(fromCell, toCell *Cell) (int, *Path) {
	if path := fromCell.PathTo(toCell); path != nil {
		return path.Length(), path
	}

	var path = NewPath()

	// Get all distances from this cell
	d := fromCell.Distances()
	toCellDist, _ := d.Get(toCell)

	current := toCell

	for current != d.root {
		smallest := math.MaxInt16
		var next *Cell
		for _, link := range current.Links() {
			dist, _ := d.Get(link)
			if dist < smallest {
				smallest = dist
				next = link
			}
		}
		segment := NewSegment(next, "north") // arbitrary facing
		path.AddSegement(segment)
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
func (g *Maze) SetDistanceColors(c *Cell) {
	// figure out the distances if needed
	c.Distances()

	_, longestPath := c.FurthestCell()

	// use alpha blending, works for any color
	for _, cell := range g.Cells() {
		d, err := c.distances.Get(cell)
		if err != nil {
			log.Printf("failed to get distance from %v to %v", c, cell)
			return
		}
		// decrease the last parameter to make the longest cells brighter. max = 255 (good = 228)
		adjustedColor := utils.AffineTransform(float32(d), 0, float32(longestPath), 0, 228)
		cell.bgColor = colors.OpacityAdjust(g.bgColor, adjustedColor)

	}

	g.fromCell = c
}

// DeadEnds returns a list of cells that are deadends (only linked to one neighbor
func (g *Maze) DeadEnds() []*Cell {
	var deadends []*Cell

	for _, cell := range g.Cells() {
		if len(cell.Links()) == 1 {
			deadends = append(deadends, cell)
		}
	}

	return deadends
}

// ResetVisited sets all cells to be unvisited
func (g *Maze) ResetVisited() {
	for _, c := range g.Cells() {
		c.SetUnVisited()
	}

}

// GetFacingDirection returns the direction walker was facing when moving fromCell -> toCell
// north, south, east, west
func (g *Maze) GetFacingDirection(fromCell, toCell *Cell) string {
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
