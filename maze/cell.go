package maze

import (
	"fmt"
	"log"
	"mazes/colors"
	"mazes/utils"

	"github.com/sasha-s/go-deadlock"
	"github.com/veandco/go-sdl2/sdl"
)

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

	// cell is isolated
	orphan bool

	deadlock.RWMutex
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
		orphan:    false,
	}
	cell.distances = NewDistances(cell)

	return cell
}

func (c *Cell) String() string {
	c.RLock()
	defer c.RUnlock()
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
	c.RLock()
	defer c.RUnlock()

	return Location{c.column, c.row}
}

// Visited returns true if the cell has been visited
func (c *Cell) Visited() bool {
	c.RLock()
	defer c.RUnlock()

	return c.visited > 0
}

// VisitedTimes returns how many times a cell has been visited
func (c *Cell) VisitedTimes() int {
	c.RLock()
	defer c.RUnlock()

	return c.visited
}

// SetVisited marks the cell as visited
func (c *Cell) SetVisited() {
	c.Lock()
	defer c.Unlock()

	c.visited++
}

// SetUnVisited marks the cell as unvisited
func (c *Cell) SetUnVisited() {
	c.Lock()
	defer c.Unlock()

	c.visited = 0
}

// SetPaths sets the paths present in the cell
func (c *Cell) SetPaths(previous, next *Cell) {
	c.Lock()
	defer c.Unlock()

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
	return c.Distances().Furthest()
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

// BGColor returns the cell's background color
func (c *Cell) BGColor() colors.Color {
	c.RLock()
	defer c.RUnlock()

	return c.bgColor
}

// SetBGColor returns the cell's background color
func (c *Cell) SetBGColor(color colors.Color) {
	c.Lock()
	defer c.Unlock()

	c.bgColor = color
}

// Draw draws one cell on renderer.
func (c *Cell) Draw(r *sdl.Renderer) *sdl.Renderer {
	var bg *sdl.Rect

	// Fill in background color
	colors.SetDrawColor(c.BGColor(), r)

	bg = &sdl.Rect{int32(c.column*c.width + c.wallWidth), int32(c.row*c.width + c.wallWidth),
		int32(c.width), int32(c.width)}
	r.FillRect(bg)

	// Draw walls as needed
	// East
	if !c.Linked(c.East) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column*c.width + c.width - c.wallWidth/2 + c.wallWidth), int32(c.row*c.width + c.wallWidth),
			int32(c.wallWidth / 2), int32(c.width + c.wallWidth/2)}
		r.FillRect(bg)
	}

	// West
	if !c.Linked(c.West) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column*c.width + c.wallWidth), int32(c.row*c.width + c.wallWidth),
			int32(c.wallWidth / 2), int32(c.width + c.wallWidth/2)}
		r.FillRect(bg)
	}

	// North
	if !c.Linked(c.North) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column*c.width + c.wallWidth), int32(c.row*c.width + c.wallWidth),
			int32(c.width), int32(c.wallWidth / 2)}
		r.FillRect(bg)
	}

	// South
	if !c.Linked(c.South) {
		colors.SetDrawColor(c.wallColor, r)
		bg = &sdl.Rect{int32(c.column*c.width + c.wallWidth), int32(c.row*c.width + c.width - c.wallWidth/2 + c.wallWidth),
			int32(c.width), int32(c.wallWidth / 2)}
		r.FillRect(bg)
	}

	return r
}

// DrawVisited draws the visited marker. If num is supplied and is not 0, use that as the times visited. Used for animation.
func (c *Cell) DrawVisited(r *sdl.Renderer) *sdl.Renderer {
	c.RLock()
	defer c.RUnlock()

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
	c.RLock()
	defer c.RUnlock()

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
	c.RLock()
	defer c.RUnlock()

	var path *sdl.Rect
	colors.SetDrawColor(c.pathColor, r)
	pathWidth := c.pathWidth
	PixelsPerCell := c.width

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
	c.RLock()
	defer c.RUnlock()

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
	c.RLock()
	defer c.RUnlock()

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
	c.RLock()
	defer c.RUnlock()

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

// Orphan isolates the cell from all of its neighbors
func (c *Cell) Orphan() {
	c.Lock()
	defer c.Unlock()

	if c.East != nil {
		c.East.West = nil
	}
	if c.West != nil {
		c.West.East = nil
	}
	if c.North != nil {
		c.North.South = nil
	}
	if c.South != nil {
		c.South.North = nil
	}

	c.orphan = true
}

func (c *Cell) IsOrphan() bool {
	c.RLock()
	defer c.RUnlock()

	return c.orphan
}
