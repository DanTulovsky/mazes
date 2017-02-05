package grid

import "fmt"

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
