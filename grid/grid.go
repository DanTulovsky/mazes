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

// Config defines the configuration parameters passed to the Grid
type Config struct {
	Rows        int
	Columns     int
	CellWidth   int // cell width
	WallWidth   int
	PathWidth   int
	BgColor     colors.Color
	BorderColor colors.Color
	WallColor   colors.Color
	PathColor   colors.Color
}

// CheckConfig makes sure the config is valid
func (c *Config) CheckConfig() error {

	if c.Rows <= 0 || c.Columns <= 0 {
		return fmt.Errorf("rows and columns must be > 0: %#v", c)
	}
	return nil
}

// Grid defines the maze grid
type Grid struct {
	config      *Config
	rows        int
	columns     int
	cells       [][]*Cell
	cellWidth   int // cell width
	wallWidth   int
	pathWidth   int
	bgColor     colors.Color
	borderColor colors.Color
	wallColor   colors.Color
	pathColor   colors.Color
}

var (
	// set by caller
	PixelsPerCell int
	longestPath   int
)

// Fail fails the process due to an unrecoverable error
func Fail(err error) {
	fmt.Println(err)
	os.Exit(1)

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
	}

	g.prepareGrid()
	g.configureCells()
	PixelsPerCell = c.CellWidth

	return g, nil
}

// ClearDrawPresent clears the buffer, draws the maze in buffer, and displays on the screen
func (g *Grid) ClearDrawPresent(r *sdl.Renderer) {
	if r == nil {
		log.Fatal("trying to render on an uninitialied render, did you pass --gui?")
	}
	r.Clear()   // clears buffer
	g.Draw(r)   // populate buffer
	r.Present() // redraw screen
}

func (g *Grid) String() string {
	output := "┌"
	for x := 0; x < g.columns-1; x++ {
		output = fmt.Sprintf("%v───┬", output)
	}
	output = output + "───┐" + "\n"

	for y := 0; y < g.rows; y++ {
		top := "│"
		bottom := "├"

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
				bottom = "└"
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
func (g *Grid) Draw(r *sdl.Renderer) *sdl.Renderer {

	// Each cell draws its background, half the wall and the path, as well as anything inside it
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
	// TODO(dan): This should be based on wallWidth, and the maze should offset itself away from the outside border
	colors.SetDrawColor(g.borderColor, r)
	bg := &sdl.Rect{0, 0, int32(g.columns) * int32(PixelsPerCell), int32(g.rows) * int32(PixelsPerCell)}
	if err := r.DrawRect(bg); err != nil {
		Fail(fmt.Errorf("error drawing border: %v", err))
	}

	return r
}

// DrawPath renders the gui maze path in memory, display by calling Present
func (g *Grid) DrawPath(r *sdl.Renderer) *sdl.Renderer {

	// TODO(dan): Figure out how to animate this
	r.Clear()
	g.Draw(r)

	// Each cell draws its background, half the wall and the path, as well as anything inside it
	for x := 0; x < g.columns; x++ {
		for y := 0; y < g.rows; y++ {
			cell, err := g.Cell(x, y)
			if err != nil {
				Fail(fmt.Errorf("Error drawing cell (%v, %v): %v", x, y, err))
			}
			cell.DrawPath(r)
		}
	}

	// Draw outside border
	// TODO(dan): This should be based on wallWidth, and the maze should offset itself away from the outside border
	colors.SetDrawColor(g.borderColor, r)
	bg := &sdl.Rect{0, 0, int32(g.columns) * int32(PixelsPerCell), int32(g.rows) * int32(PixelsPerCell)}
	if err := r.DrawRect(bg); err != nil {
		Fail(fmt.Errorf("error drawing border: %v", err))
	}

	r.Present()

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

// Size returns the number of cells in the grid
func (g *Grid) Size() int {
	return g.columns * g.rows
}

// Rows returns a list of rows (essentially the grid
func (g *Grid) Rows() [][]*Cell {
	return g.cells
}

// Cells returns a list of cells in the grid
func (g *Grid) Cells() []*Cell {
	cells := []*Cell{}
	for x := 0; x < g.columns; x++ {
		for y := 0; y < g.rows; y++ {
			cells = append(cells, g.cells[x][y])
		}
	}
	return cells
}

// LongestPath returns the longest path through the maze
func (g *Grid) LongestPath() (dist int, fromCell, toCell *Cell, path []*Cell) {

	// pick random starting point
	fromCell = g.RandomCell()

	// find furthest point
	furthest := fromCell.FurthestCell()

	// now find the furthest point from that
	toCell = furthest.FurthestCell()

	// now get the path
	dist, path = g.ShortestPath(furthest, toCell)

	// update longest path for colors
	longestPath = dist

	return dist, furthest, toCell, path
}

func reverseCells(cells []*Cell) {
	for i, j := 0, len(cells)-1; i < j; i, j = i+1, j-1 {
		cells[i], cells[j] = cells[j], cells[i]
	}
}

// SetPath draws the shortest path from fromCell to toCell
func (g *Grid) SetPath(fromCell, toCell *Cell) {
	_, path := g.ShortestPath(fromCell, toCell)

	var prev, next *Cell
	for x := 0; x < len(path); x++ {
		if x > 0 {
			prev = path[x-1]
		} else {
			prev = path[x]
		}

		if x < len(path)-1 {
			next = path[x+1]
		}
		path[x].SetPaths(prev, next)
	}
}

// ShortestPath finds the shortest path from fromCell to toCell
func (g *Grid) ShortestPath(fromCell, toCell *Cell) (int, []*Cell) {
	var path []*Cell
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
		path = append(path, next)
		current = next
	}

	// add toCell to path
	reverseCells(path)
	path = append(path, toCell)

	return toCellDist, path
}

// SetDistanceColors colors the graph based on distances from c
func (g *Grid) SetDistanceColors(c *Cell) {
	// sets the color of the new cell to be slightly darker than the previous

	// always start at white, d is the distance from the source cell
	// l.bgColor = colors.Darker(colors.GetColor("white"), d)

	// figure out the distances if needed
	c.Distances()

	// use alpha blending, works for any color
	for _, cell := range g.Cells() {
		d, err := c.distances.Get(cell)
		if err != nil {
			log.Fatalf("failed to get distance from %v to %v", c, cell)
		}
		// decrease the last parameter to make the longest cells brighter. max = 255 (good = 228)
		adjustedColor := utils.AffineTransform(float32(d), 0, float32(longestPath), 0, 228)
		cell.bgColor = colors.OpacityAdjust(g.bgColor, adjustedColor)

	}
}

type Distances struct {
	root  *Cell         // the root cell
	cells map[*Cell]int // Distance to this cell
}

func NewDistances(c *Cell) *Distances {
	return &Distances{
		root:  c,
		cells: map[*Cell]int{c: 0},
	}
}

func (d *Distances) String() string {
	// TODO(dan): Implenet
	return "TODO"
}

// Cells returns a list of cells that we have distance information for
func (d *Distances) Cells() []*Cell {
	var cells []*Cell
	for c := range d.cells {
		cells = append(cells, c)
	}
	return cells

}

// Set sets the distance to the provided cell
func (d *Distances) Set(c *Cell, dist int) {
	d.cells[c] = dist
}

// Get returns the distance to c
func (d *Distances) Get(c *Cell) (int, error) {
	dist, ok := d.cells[c]
	if !ok {
		return -1, fmt.Errorf("distance to [%v] not known", c)
	}
	return dist, nil
}

// Cell defines a single cell in the grid
type Cell struct {
	column, row int
	// keep track of neighborgs
	North, South, East, West *Cell
	// keeps track of which cells this cell has a connection (no wall) to
	links map[*Cell]bool
	// distances to other cells
	distances *Distances
	// Has this cell been visited?
	visited bool
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

	// keep track of what cells we have a path to
	pathNorth, pathSouth, pathEast, pathWest bool
}

// NewCell initializes a new cell
func NewCell(x, y int, c *Config) *Cell {
	cell := &Cell{
		row:       y,
		column:    x,
		links:     make(map[*Cell]bool),
		bgColor:   c.BgColor,   // default
		wallColor: c.WallColor, // default
		pathColor: c.PathColor, //default
		width:     c.CellWidth,
		wallWidth: c.WallWidth,
		pathWidth: c.PathWidth,
	}
	cell.distances = NewDistances(cell)

	return cell
}

func (c *Cell) String() string {
	return fmt.Sprintf("(%v, %v)", c.column, c.row)
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

// FurthestCell returns the cell that is furthest from this one
func (c *Cell) FurthestCell() *Cell {
	var furthest *Cell
	fromDist := c.Distances()

	longest := 0
	for _, c := range fromDist.Cells() {
		dist, _ := fromDist.Get(c)
		if dist > longest {
			furthest = c
			longest = dist
		}

	}
	return furthest
}

// Distances finds the distances of all cells to *this* cell
// Implements simplified Dijkstra’s algorithm
// Shades the cells
func (c *Cell) Distances() *Distances {
	if len(c.distances.cells) > 1 {
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

	// Fill in background color
	colors.SetDrawColor(c.bgColor, r)
	bg = &sdl.Rect{int32(c.column * PixelsPerCell), int32(c.row * PixelsPerCell),
		int32(PixelsPerCell), int32(PixelsPerCell)}
	r.FillRect(bg)

	// Draw walls as needed
	// East
	if !c.Linked(c.East) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column*PixelsPerCell + PixelsPerCell - c.wallWidth/2), int32(c.row * PixelsPerCell),
			int32(c.wallWidth / 2), int32(PixelsPerCell + c.wallWidth/2)}
		r.FillRect(bg)
	}

	// West
	if !c.Linked(c.West) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column * PixelsPerCell), int32(c.row * PixelsPerCell),
			int32(c.wallWidth / 2), int32(PixelsPerCell + c.wallWidth/2)}
		r.FillRect(bg)
	}

	// North
	if !c.Linked(c.North) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column * PixelsPerCell), int32(c.row * PixelsPerCell),
			int32(PixelsPerCell), int32(c.wallWidth / 2)}
		r.FillRect(bg)
	}

	// South
	if !c.Linked(c.South) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column * PixelsPerCell), int32(c.row*PixelsPerCell + PixelsPerCell - c.wallWidth/2),
			int32(PixelsPerCell), int32(c.wallWidth / 2)}
		r.FillRect(bg)
	}

	return r
}

// DrawPath draws the path as present in the cells
func (c *Cell) DrawPath(r *sdl.Renderer) *sdl.Renderer {
	var path *sdl.Rect
	colors.SetDrawColor(c.pathColor, r)
	pathWidth := c.pathWidth

	if c.pathEast {
		path = &sdl.Rect{int32(c.column*PixelsPerCell + PixelsPerCell/2),
			int32(c.row*PixelsPerCell + PixelsPerCell/2),
			int32(PixelsPerCell / 2), int32(pathWidth)}
		r.FillRect(path)
	}
	if c.pathWest {
		path = &sdl.Rect{int32(c.column * PixelsPerCell),
			int32(c.row*PixelsPerCell + PixelsPerCell/2),
			int32(PixelsPerCell/2 + pathWidth), int32(pathWidth)}
		r.FillRect(path)
	}
	if c.pathNorth {
		path = &sdl.Rect{int32(c.column*PixelsPerCell + PixelsPerCell/2),
			int32(c.row * PixelsPerCell),
			int32(pathWidth), int32(PixelsPerCell / 2)}
		r.FillRect(path)
	}
	if c.pathSouth {
		path = &sdl.Rect{int32(c.column*PixelsPerCell + PixelsPerCell/2),
			int32(c.row*PixelsPerCell + PixelsPerCell/2),
			int32(pathWidth), int32(PixelsPerCell / 2)}
		r.FillRect(path)
	}

	return r
}

func (c *Cell) linkOneWay(cell *Cell) {
	c.links[cell] = true
}

func (c *Cell) unLinkOneWay(cell *Cell) {
	delete(c.links, cell)
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

// Links returns a list of all cells linked to this one
func (c *Cell) Links() []*Cell {
	var keys []*Cell
	for k, linked := range c.links {
		if linked {
			keys = append(keys, k)
		}
	}
	return keys
}

// Linked returns true if the two cells are linked (joined by a passage)
func (c *Cell) Linked(cell *Cell) bool {
	linked, ok := c.links[cell]
	if !ok {
		return false
	}
	return linked
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
