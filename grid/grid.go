package grid

import (
	"fmt"
	"log"
	"math/rand"
	"mazes/utils"
	"time"

	"os"

	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano()) // need to initialize the seed
}

type Grid struct {
	rows    int
	columns int
	cells   [][]*Cell
}

const PixelsPerCell = 15

// Fail fails the process due to an unrecoverable error
func Fail(err error) {
	fmt.Println(err)
	os.Exit(1)

}

// NewGrid returns a new grid.
func NewGrid(r, c int) *Grid {
	g := &Grid{
		rows:    r,
		columns: c,
		cells:   [][]*Cell{},
	}

	g.prepareGrid()
	g.configureCells()
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
	// border around
	border := &sdl.Rect{0, 0, int32(g.columns) * PixelsPerCell, int32(g.rows) * PixelsPerCell}

	if err := r.DrawRect(border); err != nil {
		Fail(fmt.Errorf("error drawing border: %v", err))
	}

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

	return r
}

// prepareGrid initializes the grid with cells
func (g *Grid) prepareGrid() {
	g.cells = make([][]*Cell, g.rows)

	for x := 0; x < g.columns; x++ {
		g.cells[x] = make([]*Cell, g.rows)

		for y := 0; y < g.rows; y++ {
			g.cells[x][y] = NewCell(x, y)
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

// Cell defines a single cell in the grid
type Cell struct {
	column, row int
	// keep track of neighborgs
	North, South, East, West *Cell
	// keeps track of which cells this cell has a connection (no wall) to
	links map[*Cell]bool
}

// NewCell initializes a new cell
func NewCell(x, y int) *Cell {
	return &Cell{
		row:    y,
		column: x,
		links:  make(map[*Cell]bool),
	}
}

func (c *Cell) String() string {
	return fmt.Sprintf("(%v, %v)", c.column, c.row)
}

// Draw draws one cell on renderer. Draws the east and south walls
func (c *Cell) Draw(r *sdl.Renderer) *sdl.Renderer {

	log.Printf("drawing %v\n", c)
	log.Printf("neighbors: %v\n", c.Neighbors())
	log.Printf("links: %v\n", c.Links())
	log.Printf("east: %v\n", c.East)
	log.Printf("south: %v\n", c.South)
	// East

	log.Printf("l east: %v\n", c.Linked(c.East))
	if !c.Linked(c.East) {
		x := c.column*PixelsPerCell + PixelsPerCell
		y := c.row * PixelsPerCell
		x2 := (c.column * PixelsPerCell) + PixelsPerCell
		y2 := (c.row * PixelsPerCell) + PixelsPerCell
		log.Printf(" >East: (%v, %v) -> (%v, %v)\n", x, y, x2, y2)
		r.DrawLine(x, y, x2, y2)
	}

	// South
	log.Printf("l south: %v\n", c.Linked(c.South))
	if !c.Linked(c.South) {
		x := (c.column * PixelsPerCell)
		y := (c.row * PixelsPerCell) + PixelsPerCell
		x2 := (c.column * PixelsPerCell) + PixelsPerCell
		y2 := (c.row * PixelsPerCell) + PixelsPerCell
		log.Printf(" >South: (%v, %v) -> (%v, %v)\n", x, y, x2, y2)
		r.DrawLine(x, y, x2, y2)
	}
	log.Print("\n")

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
		log.Printf("key: %v; value: %v\n", k, linked)
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
