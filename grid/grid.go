package grid

import (
	"fmt"
	"log"
	"math/rand"
	"mazes/utils"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano()) // need to initialize the seed
}

type Grid struct {
	rows    int
	columns int
	cells   [][]*Cell
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
	output := "+"
	for x := 0; x < g.columns; x++ {
		output = fmt.Sprintf("%v---+", output)
	}
	output = output + "\n"

	for r := 0; r < g.rows; r++ {
		top := "|"
		bottom := "+"

		for c := 0; c < g.columns; c++ {
			cell, err := g.Cell(r, c)
			if err != nil {
				continue
			}
			body := "   "
			east_boundary := " "
			if !cell.Linked(cell.East) {
				east_boundary = "|"
			}
			top = fmt.Sprintf("%v%v%v", top, body, east_boundary)

			south_boundary := "   "
			if !cell.Linked(cell.South) {
				south_boundary = "---"
			}
			corner := "+"
			bottom = fmt.Sprintf("%v%v%v", bottom, south_boundary, corner)
		}
		output = fmt.Sprintf("%v%v\n", output, top)
		output = fmt.Sprintf("%v%v\n", output, bottom)
	}

	return output
}

func (g *Grid) prepareGrid() {
	g.cells = make([][]*Cell, g.rows)

	for r := 0; r < g.rows; r++ {
		g.cells[r] = make([]*Cell, g.columns)

		for c := 0; c < g.columns; c++ {
			g.cells[r][c] = NewCell(r, c)
		}
	}
}

func (g *Grid) configureCells() {
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.columns; c++ {
			cell, err := g.Cell(r, c)
			if err != nil {
				log.Fatalf("failed to initialize grid: %v", err)
			}
			// error is ignored, we just set nil if there is no neighbor
			cell.North, _ = g.Cell(r-1, c)
			cell.South, _ = g.Cell(r+1, c)
			cell.West, _ = g.Cell(r, c-1)
			cell.East, _ = g.Cell(r, c+1)
		}
	}
}

// Cell returns the cell at r,c
func (g *Grid) Cell(r, c int) (*Cell, error) {
	if r < 0 || r >= g.rows || c < 0 || c >= g.columns {
		return nil, fmt.Errorf("(%v, %v) is outside the grid", r, c)
	}
	return g.cells[r][c], nil
}

// RandomCell returns a random cell
func (g *Grid) RandomCell() *Cell {
	return g.cells[utils.Random(0, g.rows)][utils.Random(0, g.columns)]
}

// Size returns the number of cells in the grid
func (g *Grid) Size() int {
	return g.rows * g.columns
}

// Rows returns a list of rows (essentially the grid
func (g *Grid) Rows() [][]*Cell {
	return g.cells
}

// Cells returns a list of cells in the grid
// TODO(dan): Make this into a generator?
func (g *Grid) Cells() []*Cell {
	cells := []*Cell{}
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.columns; c++ {
			cells = append(cells, g.cells[r][c])
		}
	}
	return cells
}

// Cell defines a single cell in the grid
type Cell struct {
	row, column int
	// keep track of neighborgs
	North, South, East, West *Cell
	// keeps track of which cells this cell has a connection (no wall) to
	links map[*Cell]bool
}

// NewCell initializes a new cell
func NewCell(r, c int) *Cell {
	return &Cell{
		row:    r,
		column: c,
		links:  make(map[*Cell]bool),
	}
}

func (c *Cell) String() string {
	return fmt.Sprintf("(%v, %v)", c.row, c.column)
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
	keys := make([]*Cell, len(c.links))
	i := 0
	for k := range c.links {
		keys[i] = k
		i++
	}
	return keys
}

// Linked returns true if the two cells are linked (joined by a passage)
func (c *Cell) Linked(cell *Cell) bool {
	_, linked := c.links[cell] // linked if in the map
	return linked
}

// Neighbors returns a list of all cells that are neighbors (weather connected by passage or not)
func (c *Cell) Neighbors() []*Cell {
	n := []*Cell{}
	for _, cell := range []*Cell{c.North, c.South, c.East, c.West} {
		if cell != nil {
			n = append(n, cell)
		}
	}
	return n
}
