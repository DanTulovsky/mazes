package grid

import (
	"fmt"
	"log"
	"math/rand"
	"mazes/utils"
	"time"

	"os"

	"mazes/colors"

	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano()) // need to initialize the seed
}

type Grid struct {
	rows        int
	columns     int
	cells       [][]*Cell
	cellWidth   int // cell width
	bgColor     colors.Color
	borderColor colors.Color
	wallColor   colors.Color
}

var (
	// set by caller
	PixelsPerCell int
)

// Fail fails the process due to an unrecoverable error
func Fail(err error) {
	fmt.Println(err)
	os.Exit(1)

}

// NewGrid returns a new grid.
func NewGrid(r, c, w int, bgColor, borderColor, wallColor colors.Color) *Grid {
	g := &Grid{
		rows:        r,
		columns:     c,
		cells:       [][]*Cell{},
		cellWidth:   w,
		bgColor:     bgColor,
		borderColor: borderColor,
		wallColor:   wallColor,
	}

	g.prepareGrid()
	g.configureCells()
	PixelsPerCell = w

	return g
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

// Draw renders the gui maze
func (g *Grid) Draw(r *sdl.Renderer) *sdl.Renderer {

	// Each cell draws the right and bottom border
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
	colors.SetDrawColor(g.borderColor, r)
	bg := &sdl.Rect{0, 0, int32(g.columns) * int32(PixelsPerCell), int32(g.rows) * int32(PixelsPerCell)}

	if err := r.DrawRect(bg); err != nil {
		Fail(fmt.Errorf("error drawing border: %v", err))
	}
	return r
}

// prepareGrid initializes the grid with cells
func (g *Grid) prepareGrid() {
	g.cells = make([][]*Cell, g.rows)

	for x := 0; x < g.columns; x++ {
		g.cells[x] = make([]*Cell, g.rows)

		for y := 0; y < g.rows; y++ {
			g.cells[x][y] = NewCell(x, y, g.cellWidth, g.bgColor, g.wallColor)
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

// Cells returns a list of cells that we have distance information for
func (d *Distances) Cells() []*Cell {
	var cells []*Cell
	for c, _ := range d.cells {
		cells = append(cells, c)
	}
	return cells

}

// Set sets the distance to the provided cell
func (d *Distances) Set(c *Cell, dist int) {
	d.cells[c] = dist
}

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
	// size of the cell
	width int
}

// NewCell initializes a new cell
func NewCell(x, y, w int, bgColor, wallColor colors.Color) *Cell {
	c := &Cell{
		row:       y,
		column:    x,
		links:     make(map[*Cell]bool),
		bgColor:   bgColor,   // default
		wallColor: wallColor, // default
		width:     w,
	}
	c.distances = NewDistances(c)

	return c
}

func (c *Cell) String() string {
	return fmt.Sprintf("(%v, %v)", c.column, c.row)
}

// Distances finds the distances of all cells to *this* cell
// Implements simplified Dijkstra’s algorithm
func (c *Cell) Distances() *Distances {
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

				// sets the color of the new cell to be slightly darker than the previous
				r, g, b := c.bgColor.R-uint8(d), c.bgColor.G-uint8(d), c.bgColor.G-uint8(d)
				l.bgColor = colors.Color{R: r, G: g, B: b}

				newFrontier = append(newFrontier, l)
			}
		}

		frontier = newFrontier

	}
	return c.distances
}

// Draw draws one cell on renderer.
func (c *Cell) Draw(r *sdl.Renderer) *sdl.Renderer {
	// Fill cell background color. The fill is one pixel in from the wall.
	colors.SetDrawColor(c.bgColor, r)
	bg := &sdl.Rect{int32(c.column * PixelsPerCell), int32(c.row * PixelsPerCell),
		int32(PixelsPerCell), int32(PixelsPerCell)}
	r.FillRect(bg)

	// Draw walls as needed
	// East
	if !c.Linked(c.East) {
		x := c.column*PixelsPerCell + PixelsPerCell
		y := c.row * PixelsPerCell
		x2 := (c.column * PixelsPerCell) + PixelsPerCell
		y2 := (c.row * PixelsPerCell) + PixelsPerCell
		colors.SetDrawColor(c.wallColor, r)
		r.DrawLine(x, y, x2, y2)
	}

	// West
	if !c.Linked(c.West) {
		x := c.column * PixelsPerCell
		y := c.row * PixelsPerCell
		x2 := (c.column * PixelsPerCell)
		y2 := (c.row * PixelsPerCell) + PixelsPerCell
		colors.SetDrawColor(c.wallColor, r)
		r.DrawLine(x, y, x2, y2)
	}

	// North
	if !c.Linked(c.North) {
		x := (c.column * PixelsPerCell)
		y := c.row * PixelsPerCell
		x2 := (c.column * PixelsPerCell) + PixelsPerCell
		y2 := c.row * PixelsPerCell
		colors.SetDrawColor(c.wallColor, r)
		r.DrawLine(x, y, x2, y2)
	}

	// South
	if !c.Linked(c.South) {
		x := (c.column * PixelsPerCell)
		y := (c.row * PixelsPerCell) + PixelsPerCell
		x2 := (c.column * PixelsPerCell) + PixelsPerCell
		y2 := (c.row * PixelsPerCell) + PixelsPerCell
		colors.SetDrawColor(c.wallColor, r)
		r.DrawLine(x, y, x2, y2)
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
