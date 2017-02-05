package grid

import (
	"fmt"
	"log"
	"math/rand"
	"mazes/utils"
	"time"

	"os"

	"mazes/colors"

	"math"

	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano()) // need to initialize the seed
}

// Fail fails the process due to an unrecoverable error
func Fail(err error) {
	fmt.Println(err)
	os.Exit(1)

}

// Config defines the configuration parameters passed to the Grid
type Config struct {
	Rows                 int
	Columns              int
	CellWidth            int // cell width
	WallWidth            int
	PathWidth            int
	MarkVisitedCells     bool
	VisitedCellColor     colors.Color
	BgColor              colors.Color
	BorderColor          colors.Color
	WallColor            colors.Color
	PathColor            colors.Color
	CurrentLocationColor colors.Color
}

// CheckConfig makes sure the config is valid
func (c *Config) CheckConfig() error {

	if c.Rows <= 0 || c.Columns <= 0 {
		return fmt.Errorf("rows and columns must be > 0: %#v", c)
	}
	return nil
}

// Location is x,y coordinate of a cell
type Location struct {
	X, Y int
}

// LocInLocList returns true if lo is in locList
func LocInLocList(l Location, locList []Location) bool {
	for _, loc := range locList {
		if l.X == loc.X && l.Y == loc.Y {
			return true
		}
	}
	return false
}

// Grid defines the maze grid
type Grid struct {
	config           *Config
	rows             int
	columns          int
	cells            [][]*Cell
	cellWidth        int // cell width
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
func NewGrid(c *Config) (*Grid, error) {
	if err := c.CheckConfig(); err != nil {
		return nil, err
	}
	g := &Grid{
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

	g.prepareGrid()
	g.configureCells()

	return g, nil
}

func (g *Grid) SetGenCurrentLocation(cell *Cell) {
	g.genCurrentLocation = cell
}

func (g *Grid) SetCreateTime(t time.Duration) {
	g.createTime = t
}

func (g *Grid) CreateTime() time.Duration {
	return g.createTime
}

// Dimensions returns the dimensions of the grid.
func (g *Grid) Dimensions() (int, int) {

	return g.columns, g.rows
}

// ClearDrawPresent clears the buffer, draws the maze in buffer, and displays on the screen
func (g *Grid) ClearDrawPresent(r *sdl.Renderer, w *sdl.Window) {
	if r == nil {
		log.Fatal("trying to render on an uninitialied render, did you pass --gui?")
	}
	r.Clear()     // clears buffer
	g.DrawMaze(r) // populate buffer
	// g.DrawPath(r)

	r.Present() // redraw screen
}

func (g *Grid) String() string {
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

// DrawBorder renders the maze border in memory, display by calling Present
func (g *Grid) DrawBorder(r *sdl.Renderer) *sdl.Renderer {
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

// Draw renders the gui maze in memory, display by calling Present
func (g *Grid) DrawMaze(r *sdl.Renderer) *sdl.Renderer {
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
			cell.Draw(r)
		}
	}

	// Draw outside border
	g.DrawBorder(r)

	// Draw location of the generator algorithm
	g.DrawGenCurrentLocation(r)

	// Draw the path
	g.DrawPath(r, g.TravelPath, g.config.MarkVisitedCells)

	// Draw the final path
	// g.DrawPath(r, g.SolvePath, false, true)

	// Draw the location of the solver algorithm
	g.DrawSolveCurrentLocation(r)
	return r
}

func (g *Grid) DrawSolveCurrentLocation(r *sdl.Renderer) {

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

func (g *Grid) DrawGenCurrentLocation(r *sdl.Renderer) *sdl.Renderer {

	if g.genCurrentLocation != nil {
		for _, cell := range g.Cells() {
			cell.bgColor = colors.GetColor("white")
		}

		g.genCurrentLocation.bgColor = colors.GetColor("yellow")
	}
	return r
}

// DrawPath renders the gui maze path in memory, display by calling Present
// This is drawing g.TravelPath if path == nil
func (g *Grid) DrawPath(r *sdl.Renderer, path *Path, markVisited bool) *sdl.Renderer {
	if path == nil {
		path = g.TravelPath
	}

	var isSolution bool
	var isLast bool

	for x, segment := range path.segments {
		if x == len(path.segments)-1 {
			isLast = true // last segment is drawn slightly different
		}

		if CellInCellList(segment.Cell(), g.SolvePath.ListCells()) {
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
func (g *Grid) DrawVisited(r *sdl.Renderer) *sdl.Renderer {
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

// prepareGrid initializes the grid with cells
func (g *Grid) prepareGrid() {
	g.cells = make([][]*Cell, g.columns)

	for x := 0; x < g.columns; x++ {
		g.cells[x] = make([]*Cell, g.rows)

		for y := 0; y < g.rows; y++ {
			g.cells[x][y] = NewCell(x, y, g.config)
		}
	}
}

// configureCells configures cells with their neighbors
func (g *Grid) configureCells() {
	for x := 0; x < g.columns; x++ {
		for y := 0; y < g.rows; y++ {
			cell, err := g.Cell(x, y)
			if err != nil {
				log.Fatalf("failed to initialize grid: %v", err)
			}
			// error is ignored, we just set nil if there is no neighbor
			cell.North, _ = g.Cell(x, y-1)
			cell.South, _ = g.Cell(x, y+1)
			cell.West, _ = g.Cell(x-1, y)
			cell.East, _ = g.Cell(x+1, y)
		}
	}
}

// Cell returns the cell at r,c
func (g *Grid) Cell(x, y int) (*Cell, error) {
	if x < 0 || x >= g.columns || y < 0 || y >= g.rows {
		return nil, fmt.Errorf("(%v, %v) is outside the grid", x, y)
	}
	return g.cells[x][y], nil
}

// RandomCell returns a random cell
func (g *Grid) RandomCell() *Cell {
	return g.cells[utils.Random(0, g.columns)][utils.Random(0, g.rows)]
}

// RandomCellFromList returns a random cell from the provided list of cells
func (g *Grid) RandomCellFromList(cells []*Cell) *Cell {
	return cells[utils.Random(0, len(cells))]
}

// Size returns the number of cells in the grid
func (g *Grid) Size() int {
	return g.columns * g.rows
}

// Rows returns a list of rows (essentially the grid
func (g *Grid) Rows() [][]*Cell {
	rows := [][]*Cell{}

	for y := 0; y < g.rows; y++ {
		cells := []*Cell{}
		for x := 0; x < g.columns; x++ {
			cell, _ := g.Cell(x, y)
			cells = append(cells, cell)
		}
		rows = append(rows, cells)
	}
	return rows
}

// Cells returns a list of cells in the grid
func (g *Grid) Cells() []*Cell {
	cells := []*Cell{}
	for y := 0; y < g.rows; y++ {
		for x := 0; x < g.columns; x++ {
			cells = append(cells, g.cells[x][y])
		}
	}
	return cells
}

// UnvisitedCells returns a list of unvisited cells in the grid
func (g *Grid) UnvisitedCells() []*Cell {
	cells := []*Cell{}
	for x := 0; x < g.columns; x++ {
		for y := 0; y < g.rows; y++ {
			if !g.cells[x][y].Visited() {
				cells = append(cells, g.cells[x][y])
			}
		}
	}
	return cells
}

// ConnectCells connects the list of cells in order by passageways
func (g *Grid) ConnectCells(cells []*Cell) {

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
func (g *Grid) LongestPath() (dist int, fromCell, toCell *Cell, path *Path) {

	utils.TimeTrack(time.Now(), "LongestPath")

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

func (p *Path) reverseCells() {
	for i, j := 0, len(p.segments)-1; i < j; i, j = i+1, j-1 {
		p.segments[i], p.segments[j] = p.segments[j], p.segments[i]
	}
}

// SetFromToColors sets the opacity of the from and to cells to be highly visible
func (g *Grid) SetFromToColors(fromCell, toCell *Cell) {
	// Set path start and end colors
	fromCell.bgColor = colors.SetOpacity(fromCell.bgColor, 0)
	toCell.bgColor = colors.SetOpacity(toCell.bgColor, 255)

	// save thse for coor refresh.
	g.fromCell = fromCell
	g.toCell = toCell
}

// SetPath draws the shortest path from fromCell to toCell
// TODO(dant): Move this into solver
func (g *Grid) SetPath(fromCell, toCell *Cell) {
	_, path := g.ShortestPath(fromCell, toCell)
	g.SetFromToColors(fromCell, toCell)

	var prev, next *Cell
	for x := 0; x < len(path.ListCells()); x++ {
		if x > 0 {
			prev = path.ListCells()[x-1]
		} else {
			prev = path.ListCells()[x]
		}

		if x < len(path.ListCells())-1 {
			next = path.ListCells()[x+1]
		}
		path.ListCells()[x].SetPaths(prev, next)
	}
}

// SetPathFromTo draws the given path from fromCell to toCell
func (g *Grid) SetPathFromTo(fromCell, toCell *Cell, path *Path) {
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
func (g *Grid) ShortestPath(fromCell, toCell *Cell) (int, *Path) {
	utils.TimeTrack(time.Now(), "ShortestPath")

	if path := fromCell.PathTo(toCell); path != nil {
		return len(path.ListCells()), path
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
	path.reverseCells()
	segment := NewSegment(toCell, "north") // arbitrary facing
	path.AddSegement(segment)

	// record path for caching
	fromCell.SetPathTo(toCell, path)

	return toCellDist, path
}

// SetDistanceColors colors the graph based on distances from c
func (g *Grid) SetDistanceColors(c *Cell) {
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
func (g *Grid) DeadEnds() []*Cell {
	var deadends []*Cell

	for _, cell := range g.Cells() {
		if len(cell.Links()) == 1 {
			deadends = append(deadends, cell)
		}
	}

	return deadends
}

// ResetVisited sets all cells to be unvisited
func (g *Grid) ResetVisited() {
	for _, c := range g.Cells() {
		c.SetUnVisited()
	}

}

// GetFacingDirection returns the direction walker was facing when moving fromCell -> toCell
// north, south, east, west
func (g *Grid) GetFacingDirection(fromCell, toCell *Cell) string {
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

type Distances struct {
	root  *Cell // the root cell
	cells SafeMap
}

func NewDistances(c *Cell) *Distances {
	sm := NewSafeMap()
	sm.Insert(c, 0)

	return &Distances{
		root:  c,
		cells: sm,
	}
}

func (d *Distances) String() string {
	// TODO(dan): Implenet
	return "TODO"
}

// Cells returns a list of cells that we have distance information for
func (d *Distances) Cells() []*Cell {
	var cells []*Cell
	for _, c := range d.cells.Keys() {
		cells = append(cells, c)
	}
	return cells
}

// Set sets the distance to the provided cell
func (d *Distances) Set(c *Cell, dist int) {
	d.cells.Update(c, func(d interface{}, exists bool) interface{} {
		return dist
	})
}

// Get returns the distance to c
func (d *Distances) Get(c *Cell) (int, error) {
	dist, ok := d.cells.Find(c)
	if !ok {
		return -1, fmt.Errorf("distance to [%v] not known", c)
	}
	return dist.(int), nil
}

// Cell defines a single cell in the grid
type Cell struct {
	column, row int
	// keep track of neighborgs
	North, South, East, West *Cell
	// keeps track of which cells this cell has a connection (no wall) to
	links SafeMap
	// distances to other cells
	distances *Distances
	// How many times has this cell been visited?
	visited int
	// Background color of the cell
	bgColor colors.Color
	// Wall color of the cell
	wallColor colors.Color
	// path color
	pathColor colors.Color
	// size of the cell
	width     int
	wallWidth int
	pathWidth int

	// config
	config *Config

	// keep track of what cells we have a path to
	pathNorth, pathSouth, pathEast, pathWest bool

	// keep track of paths to specific cells
	paths SafeMap
}

// CellInCellList returns true if cell is in cellList
func CellInCellList(cell *Cell, cellList []*Cell) bool {
	for _, c := range cellList {
		if cell == c {
			return true
		}
	}
	return false
}

// NewCell initializes a new cell
func NewCell(x, y int, c *Config) *Cell {
	cell := &Cell{
		row:       y,
		column:    x,
		links:     NewSafeMap(),
		bgColor:   c.BgColor,   // default
		wallColor: c.WallColor, // default
		pathColor: c.PathColor, //default
		width:     c.CellWidth,
		wallWidth: c.WallWidth,
		pathWidth: c.PathWidth,
		paths:     NewSafeMap(),
		config:    c,
	}
	cell.distances = NewDistances(cell)

	return cell
}

func (c *Cell) String() string {
	return fmt.Sprintf("(%v, %v)", c.column, c.row)
}

// PathTo returns the path to the toCell or nil if not available
func (c *Cell) PathTo(toCell *Cell) *Path {
	if path, ok := c.paths.Find(toCell); ok {
		return path.(*Path)
	}
	return nil
}

// SetPathTo sets the path from this cell to toCell
func (c *Cell) SetPathTo(toCell *Cell, path *Path) {
	c.paths.Insert(toCell, path)
}

// RemovePathTo removes the path from this cell to toCell
func (c *Cell) RemovePathTo(toCell *Cell, path *Path) {
	c.paths.Delete(toCell)
}

// Location returns the x,y location of the cell
func (c *Cell) Location() Location {
	return Location{c.column, c.row}
}

// Visited returns true if the cell has been visited
func (c *Cell) Visited() bool {
	return c.visited > 0
}

// VisitedTimes returns how many times a cell has been visited
func (c *Cell) VisitedTimes() int {
	return c.visited
}

// SetVisited marks the cell as visited
func (c *Cell) SetVisited() {
	c.visited++
}

// SetUnVisited marks the cell as unvisited
func (c *Cell) SetUnVisited() {
	c.visited = 0
}

// SetPaths sets the paths present in the cell
func (c *Cell) SetPaths(previous, next *Cell) {
	if c.North == previous || c.North == next {
		c.pathNorth = true
	}
	if c.South == previous || c.South == next {
		c.pathSouth = true
	}
	if c.East == previous || c.East == next {
		c.pathEast = true
	}
	if c.West == previous || c.West == next {
		c.pathWest = true
	}
}

// FurthestCell returns the cell and distance of the cell that is furthest from this one
func (c *Cell) FurthestCell() (*Cell, int) {
	var furthest *Cell = c // you are the furthest from yourself at the start
	fromDist := c.Distances()

	longest := 0
	for _, cell := range fromDist.Cells() {
		dist, _ := fromDist.Get(cell)
		if dist > longest {
			furthest = cell
			longest = dist
		}

	}
	return furthest, longest
}

// Distances finds the distances of all cells to *this* cell
// Implements simplified Dijkstra’s algorithm
// Shades the cells
func (c *Cell) Distances() *Distances {
	if c.distances.cells.Len() > 1 {
		// Already have this info
		return c.distances
	}

	frontier := []*Cell{c}

	for len(frontier) > 0 {
		newFrontier := []*Cell{}

		for _, cell := range frontier {

			for _, l := range cell.Links() {
				if _, err := c.distances.Get(l); err == nil {
					continue // already been
				}
				d, err := c.distances.Get(cell)
				if err != nil {
					log.Fatalf("error getting distance from [%v]->[%v]: %v", c, l, err)
				}

				// sets distance to new cell
				c.distances.Set(l, d+1)
				newFrontier = append(newFrontier, l)

			}
		}
		frontier = newFrontier

	}
	return c.distances
}

// Draw draws one cell on renderer.
func (c *Cell) Draw(r *sdl.Renderer) *sdl.Renderer {
	var bg *sdl.Rect
	PixelsPerCell := c.width

	// Fill in background color
	colors.SetDrawColor(c.bgColor, r)

	bg = &sdl.Rect{int32(c.column*PixelsPerCell + c.wallWidth), int32(c.row*PixelsPerCell + c.wallWidth),
		int32(PixelsPerCell), int32(PixelsPerCell)}
	r.FillRect(bg)

	// Draw walls as needed
	// East
	if !c.Linked(c.East) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column*PixelsPerCell + PixelsPerCell - c.wallWidth/2 + c.wallWidth), int32(c.row*PixelsPerCell + c.wallWidth),
			int32(c.wallWidth / 2), int32(PixelsPerCell + c.wallWidth/2)}
		r.FillRect(bg)
	}

	// West
	if !c.Linked(c.West) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column*PixelsPerCell + c.wallWidth), int32(c.row*PixelsPerCell + c.wallWidth),
			int32(c.wallWidth / 2), int32(PixelsPerCell + c.wallWidth/2)}
		r.FillRect(bg)
	}

	// North
	if !c.Linked(c.North) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column*PixelsPerCell + c.wallWidth), int32(c.row*PixelsPerCell + c.wallWidth),
			int32(PixelsPerCell), int32(c.wallWidth / 2)}
		r.FillRect(bg)
	}

	// South
	if !c.Linked(c.South) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column*PixelsPerCell + c.wallWidth), int32(c.row*PixelsPerCell + PixelsPerCell - c.wallWidth/2 + c.wallWidth),
			int32(PixelsPerCell), int32(c.wallWidth / 2)}
		r.FillRect(bg)
	}

	return r
}

// DrawVisited draws the visited marker. If num is supplied and is not 0, use that as the times visited. Used for animation.
func (c *Cell) DrawVisited(r *sdl.Renderer) *sdl.Renderer {
	PixelsPerCell := c.width

	if c.config.MarkVisitedCells && c.Visited() {
		colors.SetDrawColor(c.config.VisitedCellColor, r)

		times := c.VisitedTimes()
		factor := times * 3

		offset := int32(c.wallWidth/4 + c.wallWidth)
		h, w := int32(c.width/10+factor), int32(c.width/10+factor)

		if h > int32(PixelsPerCell-c.wallWidth)-offset {
			h = int32(PixelsPerCell-c.wallWidth) - offset
			w = int32(PixelsPerCell-c.wallWidth) - offset
		}

		// draw a small box to mark visited cells
		box := &sdl.Rect{int32(c.column*PixelsPerCell+c.wallWidth) + offset, int32(c.row*PixelsPerCell+c.wallWidth) + offset, h, w}
		r.FillRect(box)
	}

	return r
}

// DrawCurrentLocation marks the current location of the user
func (c *Cell) DrawCurrentLocation(r *sdl.Renderer) *sdl.Renderer {
	PixelsPerCell := c.width
	colors.SetDrawColor(c.config.CurrentLocationColor, r)

	avatar := &sdl.Rect{
		int32(c.column*PixelsPerCell + PixelsPerCell/2),
		int32(c.row*PixelsPerCell + PixelsPerCell/2),
		int32(c.pathWidth * 6),
		int32(c.pathWidth * 6)}
	r.FillRect(avatar)

	return r
}

// DrawPath draws the path as present in the cells
func (c *Cell) DrawPath(r *sdl.Renderer) *sdl.Renderer {
	var path *sdl.Rect
	colors.SetDrawColor(c.pathColor, r)
	pathWidth := c.pathWidth
	PixelsPerCell := c.width

	//// shift the path right, left, up, down, depending on direction moving
	//var northX, southX, eastY, westY int32
	//var shift int32 = 30
	//
	//switch c.moveNext {
	//case "north":
	//	northX = shift
	//case "south":
	//	southX = -shift
	//case "east":
	//	eastY = shift
	//case "west":
	//	westY = -shift
	//}
	//
	//switch c.movePrevious {
	//case "north":
	//	northX = -shift
	//case "south":
	//	southX = shift
	//case "east":
	//	eastY = -shift
	//case "west":
	//	westY = shift
	//}

	if c.pathEast {
		path = &sdl.Rect{
			int32(c.column*PixelsPerCell + PixelsPerCell/2),
			int32(c.row*PixelsPerCell + PixelsPerCell/2),
			int32(PixelsPerCell/2 + c.wallWidth),
			int32(pathWidth)}
		r.FillRect(path)
	}
	if c.pathWest {
		path = &sdl.Rect{
			int32(c.column*PixelsPerCell + c.wallWidth),
			int32(c.row*PixelsPerCell + PixelsPerCell/2),
			int32(PixelsPerCell/2 + pathWidth - c.wallWidth),
			int32(pathWidth)}
		r.FillRect(path)
	}
	if c.pathNorth {
		path = &sdl.Rect{
			int32(c.column*PixelsPerCell + PixelsPerCell/2),
			int32(c.row*PixelsPerCell + c.wallWidth),
			int32(pathWidth),
			int32(PixelsPerCell/2 - c.wallWidth)}
		r.FillRect(path)
	}
	if c.pathSouth {
		path = &sdl.Rect{
			int32(c.column*PixelsPerCell + PixelsPerCell/2),
			int32(c.row*PixelsPerCell + PixelsPerCell/2),
			int32(pathWidth),
			int32(PixelsPerCell/2 + c.wallWidth)}
		r.FillRect(path)
	}

	return r
}

func (c *Cell) linkOneWay(cell *Cell) {
	c.links.Insert(cell, true)
}

func (c *Cell) unLinkOneWay(cell *Cell) {
	c.links.Delete(cell)
}

// Link links a cell to its neighbor (adds passage)
func (c *Cell) Link(cell *Cell) {
	c.linkOneWay(cell)
	cell.linkOneWay(c)
}

// UnLink unlinks a cell from its neighbor (removes passage)
func (c *Cell) UnLink(cell *Cell) {
	c.unLinkOneWay(cell)
	cell.unLinkOneWay(c)
}

// Links returns a list of all cells linked (passage to) to this one
func (c *Cell) Links() []*Cell {
	var keys []*Cell
	for _, k := range c.links.Keys() {
		if c.Linked(k) {
			keys = append(keys, k)
		}
	}
	return keys
}

// RandomLink returns a random cell linked to this one
func (c *Cell) RandomLink() *Cell {
	var keys []*Cell
	for _, k := range c.links.Keys() {
		if c.Linked(k) {
			keys = append(keys, k)
		}
	}
	return keys[utils.Random(0, len(keys))]
}

// RandomUnvisitedLink returns a random cell linked to this one that has not been visited
func (c *Cell) RandomUnvisitedLink() *Cell {
	var keys []*Cell
	for _, k := range c.links.Keys() {
		linked := c.Linked(k)
		if linked && !k.Visited() {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	return keys[utils.Random(0, len(keys))]
}

// Linked returns true if the two cells are linked (joined by a passage)
func (c *Cell) Linked(cell *Cell) bool {
	linked, ok := c.links.Find(cell)
	if !ok {
		return false
	}
	return linked.(bool)
}

// Neighbors returns a list of all cells that are neighbors (weather connected by passage or not)
func (c *Cell) Neighbors() []*Cell {
	var n []*Cell

	for _, cell := range []*Cell{c.North, c.South, c.East, c.West} {
		if cell != nil {
			n = append(n, cell)
		}
	}
	return n
}

// RandomNeighbor returns a random neighbor of this cell
func (c *Cell) RandomNeighbor() *Cell {
	var n []*Cell

	for _, cell := range []*Cell{c.North, c.South, c.East, c.West} {
		if cell != nil {
			n = append(n, cell)
		}
	}
	return n[utils.Random(0, len(n))]
}

// GetFacingDirection returns the direction walker was facing when moving to toCell from this cell
// north, south, east, west
func (c *Cell) GetFacingDirection(toCell *Cell) string {
	facing := ""

	if c.North == toCell {
		facing = "north"
	}
	if c.East == toCell {
		facing = "east"
	}
	if c.West == toCell {
		facing = "west"
	}
	if c.South == toCell {
		facing = "south"
	}
	return facing
}

// Path is a path (ordered collection of cells) through the maze
type Path struct {
	segments []*PathSegment
}

func NewPath() *Path {
	return &Path{segments: make([]*PathSegment, 0)}
}

// PathSegment is one segement of a path. A cell, and metadata.
type PathSegment struct {
	cell   *Cell
	facing string // when you came in, which way were you facing (north, south, east, west)
}

func NewSegment(c *Cell, f string) *PathSegment {
	return &PathSegment{cell: c, facing: f}
}

func (ps *PathSegment) Cell() *Cell {
	return ps.cell
}

func (ps *PathSegment) Facing() string {
	return ps.facing
}

func (ps *PathSegment) UpdateFacingDirection(f string) {
	ps.facing = f
}

// DrawPath draws the path as present in the cells
func (c *PathSegment) DrawPath(r *sdl.Renderer, g *Grid, isLast, isSolution bool) *sdl.Renderer {
	pathWidth := c.Cell().pathWidth
	PixelsPerCell := c.Cell().width

	getPathRect := func(d string, inSolution bool) *sdl.Rect {
		if !inSolution {
			pathWidth = c.Cell().pathWidth / 2
		} else {
			pathWidth = c.Cell().pathWidth
		}
		// these are the path segments from the middle towards the given direction
		paths := map[string]*sdl.Rect{
			"east": {
				int32(c.Cell().column*PixelsPerCell + PixelsPerCell/2),
				int32(c.Cell().row*PixelsPerCell + PixelsPerCell/2),
				int32(PixelsPerCell/2 + c.Cell().wallWidth),
				int32(pathWidth)},
			"west": {
				int32(c.Cell().column*PixelsPerCell + c.Cell().wallWidth),
				int32(c.Cell().row*PixelsPerCell + PixelsPerCell/2),
				int32(PixelsPerCell/2 + pathWidth - c.Cell().wallWidth),
				int32(pathWidth)},
			"north": {
				int32(c.Cell().column*PixelsPerCell + PixelsPerCell/2),
				int32(c.Cell().row*PixelsPerCell + c.Cell().wallWidth),
				int32(pathWidth),
				int32(PixelsPerCell/2 - c.Cell().wallWidth)},
			"south": {
				int32(c.Cell().column*PixelsPerCell + PixelsPerCell/2),
				int32(c.Cell().row*PixelsPerCell + PixelsPerCell/2),
				int32(pathWidth),
				int32(PixelsPerCell/2 + c.Cell().wallWidth)},
		}
		return paths[d]
	}

	pathColor := c.Cell().pathColor
	if isSolution {
		pathColor = colors.SetOpacity(pathColor, 255) // solution is fully visible
	} else {
		pathColor = colors.SetOpacity(pathColor, 60) // travel path is less visible
	}

	colors.SetDrawColor(pathColor, r)
	currentCellInSolution := CellInCellList(c.Cell(), g.SolvePath.ListCells())

	if isLast && !c.Cell().Visited() {
		switch c.Facing() {
		case "east":
			r.FillRect(getPathRect("west", currentCellInSolution))
		case "west":
			r.FillRect(getPathRect("east", currentCellInSolution))
		case "north":
			r.FillRect(getPathRect("south", currentCellInSolution))
		case "south":
			r.FillRect(getPathRect("north", currentCellInSolution))
		}

	} else {
		if c.Cell().pathEast && c.Cell().East != nil {
			// if current cell and neighbor is in the solution, solid color.
			eastInSolution := CellInCellList(c.Cell().East, g.SolvePath.ListCells())
			if eastInSolution && currentCellInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("east", eastInSolution && currentCellInSolution))

		}
		if c.Cell().pathWest && c.Cell().West != nil {
			westInSolution := CellInCellList(c.Cell().West, g.SolvePath.ListCells())
			if westInSolution && currentCellInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("west", westInSolution && currentCellInSolution))

		}
		if c.Cell().pathNorth && c.Cell().North != nil {
			northInSolution := CellInCellList(c.Cell().North, g.SolvePath.ListCells())
			if northInSolution && currentCellInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("north", northInSolution && currentCellInSolution))

		}
		if c.Cell().pathSouth && c.Cell().South != nil {
			southInSolution := CellInCellList(c.Cell().South, g.SolvePath.ListCells())
			if southInSolution && currentCellInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("south", southInSolution && currentCellInSolution))

		}
	}

	return r
}

func (p *Path) AddSegement(s *PathSegment) {
	p.segments = append(p.segments, s)

}

func (p *Path) AddSegements(s []*PathSegment) {
	for _, seg := range s {
		p.segments = append(p.segments, seg)
	}
}

func (p *Path) LastSegment() *PathSegment {
	if len(p.segments) == 0 {
		return nil
	}
	return p.segments[len(p.segments)-1]
}

// DelSegement removes the last segment from the path
func (p *Path) DelSegement() {
	p.segments = p.segments[:len(p.segments)-1]
}

func (p *Path) List() []*PathSegment {
	return p.segments
}

// Length returns the length of the path
func (p *Path) Length() int {
	return len(p.segments)
}

func (p *Path) ListCells() []*Cell {
	var cells []*Cell
	for _, s := range p.segments {
		cells = append(cells, s.Cell())
	}
	return cells
}
