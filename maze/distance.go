package maze

import "fmt"

type Distances struct {
	root                 *Cell // the root cell
	cells                *safeMap2
	furthestCell         *Cell
	furthestCellDistance int
}

func NewDistances(c *Cell) *Distances {
	sm := NewSafeMap2()
	sm.Insert(c, 0)

	return &Distances{
		root:  c,
		cells: sm,
	}
}

func (d *Distances) String() string {
	// TODO(dan): Implement
	return "TODO"
}

// Root returns the root cell that distances are calculated from
func (d *Distances) Root() *Cell {
	return d.root
}

// Cells returns a list of cells that we have distance information for
func (d *Distances) Cells() []*Cell {
	var cells []*Cell
	for item := range d.cells.Iter() {
		cells = append(cells, item.Key)
	}
	return cells
}

// Set sets the distance to the provided cell
func (d *Distances) Set(c *Cell, dist int) {
	d.cells.Update(c, dist)
}

// Get returns the distance to c
func (d *Distances) Get(c *Cell) (int, error) {
	dist, ok := d.cells.Find(c)
	if !ok {
		return -1, fmt.Errorf("distance to [%v] not known", c)
	}
	return dist.(int), nil
}

// Furthest returns one of the cells that is furthest from this one, and the distance
func (d *Distances) Furthest() (*Cell, int) {
	if d.furthestCell != nil {
		return d.furthestCell, d.furthestCellDistance
	}

	var furthest *Cell = d.root

	var longest int
	for _, cell := range d.Cells() {
		dist, _ := d.Get(cell)
		//dist = dist - int(cell.Weight()) // ignore weights

		if dist > longest {
			furthest = cell
			longest = dist
		}

	}

	// cache
	d.furthestCell = furthest
	d.furthestCellDistance = longest

	return furthest, longest
}
