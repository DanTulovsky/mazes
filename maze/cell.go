package maze

import (
	"fmt"
	"log"
	"mazes/colors"
	"mazes/utils"

	"container/heap"

	"github.com/sasha-s/go-deadlock"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_gfx"
)

// Cell defines a single cell in the grid
type Cell struct {
	column, row int
	// keep track of neighborgs
	North, South, East, West *Cell
	// keeps track of which cells this cell has a connection (no wall) to
	links *safeMap2
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
	paths *safeMap2

	// cell is isolated
	orphan bool

	// havePath cache; previous, next
	havePath map[*Cell]*Cell

	// weight of the cell, how expensive it is to traverse it
	weight int

	// distance of this cell from the beginning
	distance int

	deadlock.RWMutex
}

// CellInCellMap returns true if cell is in cellMap
func CellInCellMap(cell *Cell, cellMap map[*Cell]bool) bool {
	if _, ok := cellMap[cell]; ok {
		return true
	}
	return false
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
		links:     NewSafeMap2(),
		bgColor:   c.BgColor,   // default
		wallColor: c.WallColor, // default
		pathColor: c.PathColor, //default
		width:     c.CellWidth,
		wallWidth: c.WallWidth,
		pathWidth: c.PathWidth,
		paths:     NewSafeMap2(),
		config:    c,
		orphan:    false,
		havePath:  make(map[*Cell]*Cell),
		weight:    1,
	}
	cell.distances = NewDistances(cell)

	return cell
}

// HavePath returns true if there is a path to s (north, south, east, west)
func (c *Cell) HavePath(s string) (have bool) {
	c.RLock()
	defer c.RUnlock()

	switch s {
	case "north":
		have = c.pathNorth
	case "south":
		have = c.pathSouth
	case "east":
		have = c.pathEast
	case "west":
		have = c.pathWest
	}
	return have
}

func (c *Cell) SetHavePath(s string) {
	c.Lock()
	defer c.Unlock()

	switch s {
	case "north":
		c.pathNorth = true
	case "south":
		c.pathSouth = true
	case "east":
		c.pathEast = true
	case "west":
		c.pathWest = true
	}
}

// Weight returns the weight of the cell
func (c *Cell) Weight() int {
	c.RLock()
	defer c.RUnlock()
	return c.weight
}

// SetWeight returns the weight of the cell
func (c *Cell) SetWeight(w int) {
	c.Lock()
	defer c.Unlock()
	c.weight = w
}

// Distance returns the distance of the cell
func (c *Cell) Distance() int {
	c.RLock()
	defer c.RUnlock()
	return c.distance
}

// SetDistance sets the distance of the cell
func (c *Cell) SetDistance(d int) {
	c.Lock()
	defer c.Unlock()
	c.distance = d
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
	// no lock needed, only ever called from one thread (for now)
	if n, ok := c.havePath[previous]; ok {
		if n == next {
			return
		}
	}
	if c.North == previous || c.North == next {
		c.SetHavePath("north")
	}
	if c.South == previous || c.South == next {
		c.SetHavePath("south")
	}
	if c.East == previous || c.East == next {
		c.SetHavePath("east")
	}
	if c.West == previous || c.West == next {
		c.SetHavePath("west")
	}

	c.havePath[previous] = next
}

// FurthestCell returns the cell and distance of the cell that is furthest from this one
func (c *Cell) FurthestCell() (*Cell, int) {
	return c.Distances().Furthest()
}

// Distances finds the distances of all cells to *this* cell
// Implements simplified Dijkstra’s algorithm
// Shades the cells
//func (c *Cell) Distances() *Distances {
//	if c.distances.cells.Len() > 1 {
//		// Already have this info
//		return c.distances
//	}
//
//	frontier := []*Cell{c}
//
//	for len(frontier) > 0 {
//		newFrontier := []*Cell{}
//
//		for _, cell := range frontier {
//
//			for _, l := range cell.Links() {
//				if _, err := c.distances.Get(l); err == nil {
//					continue // already been
//				}
//				d, err := c.distances.Get(cell)
//				if err != nil {
//					log.Fatalf("error getting distance from [%v]->[%v]: %v", c, l, err)
//				}
//
//				// sets distance to new cell
//				c.distances.Set(l, d+1)
//				newFrontier = append(newFrontier, l)
//
//			}
//		}
//		frontier = newFrontier
//
//	}
//	return c.distances
//}

// Distances finds the distances of all cells to *this* cell
// Includes weight information
// Shades the cells
func (c *Cell) Distances() *Distances {
	log.Printf("getting distances for %v", c)
	if c.distances.cells.Len() > 1 {
		// Already have this info
		log.Println("cached")
		return c.distances
	}

	pending := make(CellPriorityQueue, 0)
	heap.Init(&pending)
	heap.Push(&pending, c)

	for pending.Len() > 0 {
		cell := heap.Pop(&pending).(*Cell)

		for _, l := range cell.Links() {

			d, err := c.distances.Get(cell)
			// log.Printf("d: %v", d)
			if err != nil {
				log.Fatalf("error getting distance from [%v]->[%v]: %v", c, l, err)
			}

			totalWeight := int(d) + l.weight // never changes once set

			prevDistance, err := c.distances.Get(l)

			if totalWeight < int(prevDistance) || err != nil {
				heap.Push(&pending, l)
				// sets distance to new cell
				// log.Printf("totalWeight: %v", totalWeight)
				c.distances.Set(l, int(totalWeight))
			}
		}

	}
	log.Printf("got distances: %v", c.distances)
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
	// defer utils.TimeTrack(time.Now(), "CellDraw")
	wallSpace := c.config.WallSpace / 2

	// Fill in background color
	colors.SetDrawColor(c.BGColor(), r)

	var x, y, w, h int

	x = c.column*c.width + c.wallWidth + wallSpace + c.wallWidth/2
	y = c.row*c.width + c.wallWidth + wallSpace + c.wallWidth/2
	w = c.width - wallSpace*2 - c.wallWidth/2 - c.wallWidth/2
	h = c.width - wallSpace*2 - c.wallWidth/2 - c.wallWidth/2

	r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	linkEast, linkWest, linkSouth, linkNorth := c.Linked(c.East), c.Linked(c.West), c.Linked(c.South), c.Linked(c.North)

	// Draw walls as needed

	// draw stubs
	if linkNorth {
		// background
		colors.SetDrawColor(c.BGColor(), r)
		// colors.SetDrawColor(colors.GetColor("yellow"), r)
		x = c.column*c.width + c.wallWidth + wallSpace + c.wallWidth/2
		y = c.row*c.width + c.wallWidth
		w = c.width - wallSpace*2 - c.wallWidth
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		colors.SetDrawColor(c.wallColor, r)
		// colors.SetDrawColor(colors.GetColor("red"), r)

		// east
		x = c.column*c.width + c.width - wallSpace + c.wallWidth/2
		// log.Printf("c.width: %v", c.width)
		y = c.row*c.width + c.wallWidth
		w = c.wallWidth / 2
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		// west
		x = c.column*c.width + c.wallWidth + wallSpace
		y = c.row*c.width + c.wallWidth
		w = c.wallWidth / 2
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	if linkSouth {
		// background
		colors.SetDrawColor(c.BGColor(), r)
		// colors.SetDrawColor(colors.GetColor("red"), r)
		x = c.column*c.width + c.wallWidth + wallSpace + c.wallWidth/2
		y = c.row*c.width + c.width - wallSpace + c.wallWidth/2
		w = c.width - wallSpace*2 - c.wallWidth
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		colors.SetDrawColor(c.wallColor, r)
		// colors.SetDrawColor(colors.GetColor("green"), r)
		// east
		x = c.column*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		y = c.row*c.width + c.width - wallSpace + c.wallWidth/2
		w = c.wallWidth / 2
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		// west
		x = c.column*c.width + c.wallWidth + wallSpace
		y = c.row*c.width + wallSpace + c.width + c.wallWidth/2 - wallSpace*2
		w = c.wallWidth / 2
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	if linkEast {
		// background
		colors.SetDrawColor(c.BGColor(), r)
		// colors.SetDrawColor(colors.GetColor("blue"), r)
		x = c.column*c.width + c.wallWidth/2 + wallSpace + c.width - wallSpace*2
		y = c.row*c.width + c.wallWidth + wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.width - wallSpace*2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		colors.SetDrawColor(c.wallColor, r)
		//colors.SetDrawColor(colors.GetColor("green"), r)

		// north
		x = c.column*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		y = c.row*c.width + c.wallWidth + wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		// south
		x = c.column*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		y = c.row*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	if linkWest {
		// background
		colors.SetDrawColor(c.BGColor(), r)
		// colors.SetDrawColor(colors.GetColor("pink"), r)
		x = c.column*c.width + c.wallWidth
		y = c.row*c.width + c.wallWidth + wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.width - wallSpace*2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		colors.SetDrawColor(c.wallColor, r)
		//colors.SetDrawColor(colors.GetColor("red"), r)

		// north
		x = c.column*c.width + c.wallWidth
		y = c.row*c.width + c.wallWidth + wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		// south
		x = c.column*c.width + c.wallWidth
		y = c.row*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// walls
	colors.SetDrawColor(c.wallColor, r)

	// East
	if !linkEast {
		x = c.column*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		y = c.row*c.width + c.wallWidth + wallSpace
		w = c.wallWidth / 2
		h = c.width - wallSpace*2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// West
	if !linkWest {
		x = c.column*c.width + c.wallWidth + wallSpace
		y = c.row*c.width + c.wallWidth + wallSpace
		w = c.wallWidth / 2
		h = c.width - wallSpace*2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// North
	if !linkNorth {
		x = c.column*c.width + c.wallWidth + wallSpace
		y = c.row*c.width + c.wallWidth + wallSpace
		w = c.width - wallSpace*2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// South
	if !linkSouth {
		x = c.column*c.width + c.wallWidth + wallSpace
		y = c.row*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		w = c.width - wallSpace*2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// Display distance value
	if c.config.ShowDistanceValues {
		x := c.column*c.width + c.wallWidth + 1 + wallSpace
		y := c.row*c.width + c.wallWidth + 1 + wallSpace
		gfx.StringRGBA(r, x, y, fmt.Sprintf("%v", c.Distance()), 0, 0, 0, 255)
	}

	return r
}

// DrawVisited draws the visited marker.
func (c *Cell) DrawVisited(r *sdl.Renderer) *sdl.Renderer {
	c.RLock()
	defer c.RUnlock()

	PixelsPerCell := c.width

	if c.config.MarkVisitedCells && c.Visited() {
		colors.SetDrawColor(c.config.VisitedCellColor, r)

		times := c.VisitedTimes()
		factor := times * 3

		wallSpace := c.config.WallSpace / 2

		offset := int32(c.wallWidth/4 + c.wallWidth + wallSpace)
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

func (c *Cell) linkOneWay(cell *Cell) {
	c.links.Insert(cell, true)
}

func (c *Cell) unLinkOneWay(cell *Cell) {
	c.links.Delete(cell)
}

// Link links a cell to its neighbor (adds passage)
func (c *Cell) Link(cell *Cell) {
	if cell == nil {
		log.Fatalf("linking %v to nil!", c)
	}
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
	for item := range c.links.Iter() {
		if c.Linked(item.Key) {
			keys = append(keys, item.Key)
		}
	}
	return keys
}

// RandomLink returns a random cell linked to this one
func (c *Cell) RandomLink() *Cell {
	var keys []*Cell
	for item := range c.links.Iter() {
		if c.Linked(item.Key) {
			keys = append(keys, item.Key)
		}
	}
	return keys[utils.Random(0, len(keys))]
}

// RandomUnLink returns a random cell not linked to this one, but one that is a neighbor
func (c *Cell) RandomUnLink() *Cell {
	var keys []*Cell
	for _, k := range c.Neighbors() {
		if !c.Linked(k) {
			keys = append(keys, k)
		}
	}
	return keys[utils.Random(0, len(keys))]
}

// RandomUnLinkPreferDeadends returns a random cell not linked to this one, but one that is a neighbor
// It prefers returning a cell that is itself a deadend
func (c *Cell) RandomUnLinkPreferDeadends() *Cell {
	var keys []*Cell
	var deadends []*Cell

	for _, k := range c.Neighbors() {
		if !c.Linked(k) {
			keys = append(keys, k)
		}
		if len(k.Links()) == 1 {
			deadends = append(deadends, k)
		}
	}
	if len(deadends) > 0 {
		return keys[utils.Random(0, len(deadends))]
	}
	return keys[utils.Random(0, len(keys))]
}

// RandomUnvisitedLink returns a random cell linked to this one that has not been visited
func (c *Cell) RandomUnvisitedLink() *Cell {
	var keys []*Cell
	for item := range c.links.Iter() {
		linked := c.Linked(item.Key)
		if linked && !item.Key.Visited() {
			keys = append(keys, item.Key)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	return keys[utils.Random(0, len(keys))]
}

// Linked returns true if the two cells are linked (joined by a passage)
func (c *Cell) Linked(cell *Cell) bool {
	if c == nil || cell == nil {
		return false
	}
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
