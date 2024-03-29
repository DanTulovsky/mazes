package maze

import (
	"container/heap"
	"fmt"
	"log"
	"strconv"

	"github.com/DanTulovsky/mazes/colors"
	pb "github.com/DanTulovsky/mazes/proto"
	"github.com/DanTulovsky/mazes/utils"

	"github.com/sasha-s/go-deadlock"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	// VisitedGenerator ...
	VisitedGenerator = "generator"
)

// Cell defines a single cell in the grid
type Cell struct {
	x, y, z int64
	// keep track of neighbors
	north, south, east, west, below *Cell
	// keeps track of which cells this cell has a connection (no wall) to
	links *safeMap2
	// distances to other cells
	distances *Distances
	// How many times has this cell been visited?
	// per client and 'generator'
	visited map[string]int64
	// Background color of the cell
	bgColor colors.Color
	// Wall color of the cell
	wallColor colors.Color
	// size of the cell
	width     int64
	wallWidth int64
	pathWidth int64

	// config
	config *pb.MazeConfig

	// keep track of what cells we have a path to for each client
	pathNorth, pathSouth, pathEast, pathWest map[string]bool

	// keep track of paths to specific cells
	paths *safeMap2

	// cell is isolated
	orphan bool

	// havePath cache; previous, next, per client
	havePath map[string]map[*Cell]*Cell

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
func NewCell(x, y, z int64, c *pb.MazeConfig) *Cell {
	cell := &Cell{
		y:         y,
		x:         x,
		z:         z,
		links:     NewSafeMap2(),
		bgColor:   colors.GetColor(c.BgColor),   // default
		wallColor: colors.GetColor(c.WallColor), // default
		width:     c.CellWidth,
		wallWidth: c.WallWidth,
		pathWidth: c.PathWidth,
		paths:     NewSafeMap2(),
		config:    c,
		orphan:    false,
		havePath:  make(map[string]map[*Cell]*Cell),
		weight:    1,
		visited:   make(map[string]int64),
		pathNorth: make(map[string]bool),
		pathSouth: make(map[string]bool),
		pathEast:  make(map[string]bool),
		pathWest:  make(map[string]bool),
	}
	cell.distances = NewDistances(cell)

	return cell
}

// Encode encodes the cell (shape and cells/passages) to ascii
// Each character is created by encoding the passages present in the cell into one of the 4 bits
// north, south, east, west
// e.g. 0000 = no passages = 0
// 1000 = passage north = 1
// 1100 = passage north and south = C
func (c *Cell) Encode() string {
	var e int

	switch 1 {
	case 1:
		if c.North() != nil && c.Linked(c.North()) {
			e = utils.SetBit(e, 3)
		}
		fallthrough
	case 2:
		if c.South() != nil && c.Linked(c.South()) {
			e = utils.SetBit(e, 2)
		}
		fallthrough
	case 3:
		if c.East() != nil && c.Linked(c.East()) {
			e = utils.SetBit(e, 1)
		}
		fallthrough
	case 4:
		if c.West() != nil && c.Linked(c.West()) {
			e = utils.SetBit(e, 0)
		}
	}

	return fmt.Sprintf("%X", e)
}

// Decode decodes the neighbors of a cell from the encoded string and sets them
func (c *Cell) Decode(e string) error {
	i, err := strconv.ParseInt(e, 16, 0)
	if err != nil {
		return err
	}
	enc := int(i)

	switch 1 {
	case 1:
		if utils.HasBit(enc, 3) {
			if err := c.Link(c.North()); err != nil {
				return fmt.Errorf("encNorth: %b (%X); cell: %v, err: %v", enc, enc, c, err)
			}
		}
		fallthrough
	case 2:
		if utils.HasBit(enc, 2) {
			if err := c.Link(c.South()); err != nil {
				return fmt.Errorf("encSouth: %b (%X); cell: %v, err: %v", enc, enc, c, err)
			}
		}
		fallthrough
	case 3:
		if utils.HasBit(enc, 1) {
			if err := c.Link(c.East()); err != nil {
				return fmt.Errorf("encEast: %b (%X); cell: %v, err: %v", enc, enc, c, err)
			}
		}
		fallthrough
	case 4:
		if utils.HasBit(enc, 0) {
			if err := c.Link(c.West()); err != nil {
				return fmt.Errorf("encWest: %b (%X); cell: %v, err: %v", enc, enc, c, err)
			}
		}
	}

	return nil
}

// SetBelow ...
func (c *Cell) SetBelow(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.below = cell
}

// SetNorth ...
func (c *Cell) SetNorth(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.north = cell
}

// SetSouth ...
func (c *Cell) SetSouth(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.south = cell
}

// SetEast ...
func (c *Cell) SetEast(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.east = cell
}

// SetWest ...
func (c *Cell) SetWest(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.west = cell
}

// Below ...
func (c *Cell) Below() *Cell {
	c.RLock()
	defer c.RUnlock()
	return c.below
}

// North ...
func (c *Cell) North() *Cell {
	c.RLock()
	defer c.RUnlock()
	return c.north
}

// South ...
func (c *Cell) South() *Cell {
	c.RLock()
	defer c.RUnlock()
	return c.south
}

// East ...
func (c *Cell) East() *Cell {
	c.RLock()
	defer c.RUnlock()
	return c.east
}

// West ...
func (c *Cell) West() *Cell {
	c.RLock()
	defer c.RUnlock()
	return c.west
}

// HavePath returns true if there is a path to s (north, south, east, west)
func (c *Cell) HavePath(client *client, s string) (have bool) {
	c.RLock()
	defer c.RUnlock()

	switch s {
	case "north":
		have = c.pathNorth[client.id]
	case "south":
		have = c.pathSouth[client.id]
	case "east":
		have = c.pathEast[client.id]
	case "west":
		have = c.pathWest[client.id]
	}
	return have
}

// SetHavePath ...
func (c *Cell) SetHavePath(client *client, s string) {
	c.Lock()
	defer c.Unlock()

	switch s {
	case "north":
		c.pathNorth[client.id] = true
	case "south":
		c.pathSouth[client.id] = true
	case "east":
		c.pathEast[client.id] = true
	case "west":
		c.pathWest[client.id] = true
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

// IncDistance increments the distance of the cell by one
func (c *Cell) IncDistance() {
	c.Lock()
	defer c.Unlock()
	c.distance++
}

func (c *Cell) String() string {
	c.RLock()
	defer c.RUnlock()
	return fmt.Sprintf("(%v, %v, %v)", c.x, c.y, c.z)
}

// StringXY returns the cell coordinates as x,y pair
func (c *Cell) StringXY() string {
	c.RLock()
	defer c.RUnlock()
	return fmt.Sprintf("%v,%v", c.x, c.y)
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
func (c *Cell) Location() *pb.MazeLocation {
	//c.RLock()
	//defer c.RUnlock()
	// no lock needed, never changes after cell creation

	return &pb.MazeLocation{X: c.x, Y: c.y, Z: c.z}
}

// Visited returns true if the cell has been visited
func (c *Cell) Visited(client string) bool {
	c.RLock()
	defer c.RUnlock()

	if t, ok := c.visited[client]; ok {
		return t > 0
	}
	return false
}

// VisitedTimes returns how many times a cell has been visited
func (c *Cell) VisitedTimes(client string) int64 {
	c.RLock()
	defer c.RUnlock()

	if t, ok := c.visited[client]; ok {
		return t
	}
	return 0
}

// SetVisited marks the cell as visited
func (c *Cell) SetVisited(client string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.visited[client]; !ok {
		c.visited[client] = 0
	}
	c.visited[client]++
}

// SetUnVisited marks the cell as unvisited
func (c *Cell) SetUnVisited(client string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.visited[client]; !ok {
		c.visited[client] = 0
	}

	c.visited[client] = 0
}

// SetPaths sets the paths present in the cell
func (c *Cell) SetPaths(client *client, previous, next *Cell) {
	// no lock needed, only ever called from one thread (for now)

	if _, ok := c.havePath[client.id]; !ok {
		c.havePath[client.id] = make(map[*Cell]*Cell)
	}

	if n, ok := c.havePath[client.id][previous]; ok {
		if n == next {
			return
		}
	}
	if c.North() == previous || c.North() == next {
		c.SetHavePath(client, "north")
	}
	if c.South() == previous || c.South() == next {
		c.SetHavePath(client, "south")
	}
	if c.East() == previous || c.East() == next {
		c.SetHavePath(client, "east")
	}
	if c.West() == previous || c.West() == next {
		c.SetHavePath(client, "west")
	}

	c.havePath[client.id][previous] = next
}

// FurthestCell returns the cell and distance of the cell that is furthest from this one
func (c *Cell) FurthestCell() (*Cell, int) {
	return c.Distances().Furthest()
}

// Distances finds the distances of all cells to *this* cell
// Includes weight information
// Shades the cells
func (c *Cell) Distances() *Distances {
	if c.distances.cells.Len() > 1 {
		// Already have this info
		return c.distances
	}

	pending := make(CellPriorityQueue, 0)
	heap.Init(&pending)
	heap.Push(&pending, c)

	for pending.Len() > 0 {
		cell := heap.Pop(&pending).(*Cell)

		for _, l := range cell.Links() {
			d, err := c.distances.Get(cell)
			if err != nil {
				log.Fatalf("error getting distance from [%v]->[%v]: %v", c, l, err)
			}

			totalWeight := d + l.weight // never changes once set

			prevDistance, err := c.distances.Get(l)

			if totalWeight < prevDistance || err != nil {
				heap.Push(&pending, l)
				// sets distance to new cell
				c.distances.Set(l, totalWeight)
			}
		}

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
	// defer utils.TimeTrack(time.Now(), "CellDraw")
	wallSpace := c.config.WallSpace / 2

	// Fill in background color
	colors.SetDrawColor(c.BGColor(), r)

	var x, y, w, h int64

	if c.z >= 0 { // don't color below cells
		x = c.x*c.width + c.wallWidth + wallSpace + c.wallWidth/2
		y = c.y*c.width + c.wallWidth + wallSpace + c.wallWidth/2
		w = c.width - wallSpace*2 - c.wallWidth/2 - c.wallWidth/2
		h = c.width - wallSpace*2 - c.wallWidth/2 - c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}
	linkEast, linkWest, linkSouth, linkNorth := c.Linked(c.East()), c.Linked(c.West()), c.Linked(c.South()), c.Linked(c.North())

	// Draw walls as needed

	// draw stubs
	if linkNorth {
		// background
		colors.SetDrawColor(c.BGColor(), r)
		x = c.x*c.width + c.wallWidth + wallSpace + c.wallWidth/2
		y = c.y*c.width + c.wallWidth
		w = c.width - wallSpace*2 - c.wallWidth
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		colors.SetDrawColor(c.wallColor, r)
		// colors.SetDrawColor(colors.GetColor("red"), r)

		// east
		x = c.x*c.width + c.width - wallSpace + c.wallWidth/2
		y = c.y*c.width + c.wallWidth
		w = c.wallWidth / 2
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		// west
		x = c.x*c.width + c.wallWidth + wallSpace
		y = c.y*c.width + c.wallWidth
		w = c.wallWidth / 2
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	if linkSouth {
		// background
		colors.SetDrawColor(c.BGColor(), r)
		x = c.x*c.width + c.wallWidth + wallSpace + c.wallWidth/2
		y = c.y*c.width + c.width - wallSpace + c.wallWidth/2
		w = c.width - wallSpace*2 - c.wallWidth
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		colors.SetDrawColor(c.wallColor, r)
		// east
		x = c.x*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		y = c.y*c.width + c.width - wallSpace + c.wallWidth/2
		w = c.wallWidth / 2
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		// west
		x = c.x*c.width + c.wallWidth + wallSpace
		y = c.y*c.width + wallSpace + c.width + c.wallWidth/2 - wallSpace*2
		w = c.wallWidth / 2
		h = wallSpace + c.wallWidth/2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	if linkEast {
		// background
		colors.SetDrawColor(c.BGColor(), r)
		// colors.SetDrawColor(colors.GetColor("blue"), r)
		x = c.x*c.width + c.wallWidth/2 + wallSpace + c.width - wallSpace*2
		y = c.y*c.width + c.wallWidth + wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.width - wallSpace*2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		colors.SetDrawColor(c.wallColor, r)

		// north
		x = c.x*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		y = c.y*c.width + c.wallWidth + wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		// south
		x = c.x*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		y = c.y*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	if linkWest {
		// background
		colors.SetDrawColor(c.BGColor(), r)
		x = c.x*c.width + c.wallWidth
		y = c.y*c.width + c.wallWidth + wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.width - wallSpace*2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		colors.SetDrawColor(c.wallColor, r)

		// north
		x = c.x*c.width + c.wallWidth
		y = c.y*c.width + c.wallWidth + wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})

		// south
		x = c.x*c.width + c.wallWidth
		y = c.y*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		w = wallSpace + c.wallWidth/2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// Don't draw anything below here for cells below other cells
	if c.z < 0 {
		return r
	}

	// walls
	colors.SetDrawColor(c.wallColor, r)

	// East
	if !linkEast {
		x = c.x*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		y = c.y*c.width + c.wallWidth + wallSpace
		w = c.wallWidth / 2
		h = c.width - wallSpace*2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// West
	if !linkWest {
		x = c.x*c.width + c.wallWidth + wallSpace
		y = c.y*c.width + c.wallWidth + wallSpace
		w = c.wallWidth / 2
		h = c.width - wallSpace*2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// North
	if !linkNorth {
		x = c.x*c.width + c.wallWidth + wallSpace
		y = c.y*c.width + c.wallWidth + wallSpace
		w = c.width - wallSpace*2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// South
	if !linkSouth {
		x = c.x*c.width + c.wallWidth + wallSpace
		y = c.y*c.width + c.width - c.wallWidth/2 + c.wallWidth - wallSpace
		w = c.width - wallSpace*2
		h = c.wallWidth / 2

		r.FillRect(&sdl.Rect{int32(x), int32(y), int32(w), int32(h)})
	}

	// Display distance value
	if c.config.GetShowDistanceValues() {
		x := c.x*c.width + c.wallWidth + 1 + wallSpace
		y := c.y*c.width + c.wallWidth + 1 + wallSpace

		if e := gfx.StringRGBA(r, int32(x), int32(y), fmt.Sprintf("%v", c.Distance()), 0, 0, 0, 255); e != true {
			log.Printf("error: %v", sdl.GetError())
		}
		gfx.SetFont(nil, 0, 0)
	}

	if c.config.GetShowWeightValues() {
		x := c.x*c.width + c.wallWidth + 1 + wallSpace
		y := c.y*c.width + c.wallWidth + 1 + wallSpace

		if e := gfx.StringRGBA(r, int32(x), int32(y), fmt.Sprintf("%v", c.Weight()), 0, 0, 0, 255); e != true {
			log.Printf("error: %v", sdl.GetError())
		}
		gfx.SetFont(nil, 0, 0)
	}

	return r
}

// DrawCurrentLocation marks the current location of the user
func (c *Cell) DrawCurrentLocation(r *sdl.Renderer, client *client, avatar *sdl.Texture, facing string) {

	PixelsPerCell := c.width

	// rotateAngle returns the angle of rotation based on facing direction
	// the texture used for the avatar is assumed to be "facing" "west"
	rotateAngle := func(f string) (angle float64, flip sdl.RendererFlip) {

		switch f {
		case "north":
			angle = 90
			flip = sdl.FLIP_NONE

		case "east":
			angle = 180
			flip = sdl.FLIP_VERTICAL

		case "south":
			angle = -90
			flip = sdl.FLIP_NONE

		case "west":
			angle = 0
			flip = sdl.FLIP_NONE
		}

		return angle, flip
	}

	if avatar == nil {
		colors.SetDrawColor(colors.GetColor(client.config.CurrentLocationColor), r)
		// draw a standard box
		sq := &sdl.Rect{
			int32(c.x*PixelsPerCell + PixelsPerCell/4),
			int32(c.y*PixelsPerCell + PixelsPerCell/4),
			int32(PixelsPerCell/2 - c.wallWidth/2),
			int32(PixelsPerCell/2 - c.wallWidth/2)}
		r.FillRect(sq)
	} else {
		angle, flip := rotateAngle(facing)

		sq := &sdl.Rect{
			int32(c.x*PixelsPerCell + PixelsPerCell/4),
			int32(c.y*PixelsPerCell + PixelsPerCell/4),
			int32(c.pathWidth * 15),
			int32(c.pathWidth * 15)}

		r.CopyEx(avatar, nil, sq, angle, nil, flip)
	}
}

// DrawVisited draws the visited marker.
func (c *Cell) DrawVisited(r *sdl.Renderer, client *client) {
	if client.config.NumberMarkVisitedCells {
		wallSpace := c.config.WallSpace / 2
		x := c.x*c.width + c.wallWidth + 1 + wallSpace
		y := c.y*c.width + c.wallWidth + 1 + wallSpace

		if e := gfx.StringRGBA(r, int32(x), int32(y), fmt.Sprint(c.VisitedTimes(client.id)), 0, 0, 0, 255); e != true {
			log.Printf("error: %v", sdl.GetError())
		}
		gfx.SetFont(nil, 0, 0)
	}

	if client.config.MarkVisitedCells {
		PixelsPerCell := c.width

		// don't mark cells under other cell
		if client.config.MarkVisitedCells && c.Visited(client.id) && c.z >= 0 {
			colors.SetDrawColor(colors.GetColor(client.config.VisitedCellColor), r)

			times := c.VisitedTimes(client.id)
			factor := times * 3

			wallSpace := c.config.WallSpace / 2

			offset := int32(c.wallWidth/4 + c.wallWidth + wallSpace)
			h, w := int32(c.width/10+factor), int32(c.width/10+factor)

			if h > int32(PixelsPerCell-c.wallWidth)-offset {
				h = int32(PixelsPerCell-c.wallWidth) - offset
				w = int32(PixelsPerCell-c.wallWidth) - offset
			}

			// draw a small box to mark visited cells
			box := &sdl.Rect{int32(c.x*PixelsPerCell+c.wallWidth) + offset, int32(c.y*PixelsPerCell+c.wallWidth) + offset, h, w}
			r.FillRect(box)
		}
	}
}

func (c *Cell) linkOneWay(cell *Cell) error {
	if cell == nil {
		return fmt.Errorf("linkOneWay: cannot link %v to nil", c)
	}
	c.links.Insert(cell, true)
	return nil
}

func (c *Cell) unLinkOneWay(cell *Cell) error {
	if cell == nil {
		return fmt.Errorf("unLinkOneWay: cannot link %v to nil", c)
	}
	c.links.Delete(cell)
	return nil
}

// Link unlinks a cell from its neighbor (removes passage)
func (c *Cell) Link(cell *Cell) error {
	if cell == nil {
		return fmt.Errorf("error in Link: cannot link %v to nil", c)
	}
	c.linkOneWay(cell)
	cell.linkOneWay(c)

	return nil
}

// UnLink unlinks a cell from its neighbor (removes passage)
func (c *Cell) UnLink(cell *Cell) error {
	if cell == nil {
		return fmt.Errorf("error in UnLink: cannot link %v to nil", c)
	}
	c.unLinkOneWay(cell)
	cell.unLinkOneWay(c)
	return nil
}

// Links returns a list of all cells linked (passage to) to this one
func (c *Cell) Links() []*Cell {
	var keys []*Cell
	if c.links == nil {
		return keys
	}
	for item := range c.links.Iter() {
		if c.Linked(item.Key) {
			keys = append(keys, item.Key)
		}
	}
	return keys
}

// DirectionTo returns the direction (north, south, east, west) of cell from c
// c and cell must be linked
// TODO(dan): raise appropriate error if cells are not linked
func (c *Cell) DirectionTo(cell *Cell, client string) (*pb.Direction, error) {

	switch {
	case c.North() == cell:
		return &pb.Direction{Name: "north", Visited: cell.Visited(client)}, nil
	case c.South() == cell:
		return &pb.Direction{Name: "south", Visited: cell.Visited(client)}, nil
	case c.East() == cell:
		return &pb.Direction{Name: "east", Visited: cell.Visited(client)}, nil
	case c.West() == cell:
		return &pb.Direction{Name: "west", Visited: cell.Visited(client)}, nil
	}

	return &pb.Direction{Name: "", Visited: false}, fmt.Errorf("error: cell [%v] not linked to [%v]", c, cell)
}

// DirectionLinks returns a list of directions that have linked (passage to) cells
func (c *Cell) DirectionLinks(client string) []*pb.Direction {
	var directions []*pb.Direction
	for item := range c.links.Iter() {
		if c.Linked(item.Key) {
			d, err := c.DirectionTo(item.Key, client)
			if err != nil {
				return directions
			}

			directions = append(directions, d)
		}
	}
	return directions
}

// RandomLink returns a random cell linked to this one
// an error is return if there are no such cells
func (c *Cell) RandomLink() (*Cell, error) {
	var keys []*Cell
	for item := range c.links.Iter() {
		if c.Linked(item.Key) {
			keys = append(keys, item.Key)
		}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no cells linked to %v", c)
	}
	return keys[utils.Random(0, len(keys))], nil
}

// RandomUnLink returns a random cell not linked to this one, but one that is a neighbor
// an error is return if there are no such cells
func (c *Cell) RandomUnLink() (*Cell, error) {
	var keys []*Cell
	for _, k := range c.Neighbors() {
		if !c.Linked(k) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no cells unlinked from %v", c)
	}
	return keys[utils.Random(0, len(keys))], nil
}

// UnLinked returns all cells not linked anywhere, but ones that are neighbors
func (c *Cell) UnLinked() []*Cell {
	var keys []*Cell
	for _, k := range c.Neighbors() {
		if !c.Linked(k) {
			if len(k.Links()) == 0 {
				keys = append(keys, k)
			}
		}
	}
	return keys
}

// RandomUnLinkPreferDeadEnds returns a random cell not linked to this one, but one that is a neighbor
// It prefers returning a cell that is itself a dead end
func (c *Cell) RandomUnLinkPreferDeadEnds() *Cell {
	var keys []*Cell
	var deadEnds []*Cell

	for _, k := range c.Neighbors() {
		if !c.Linked(k) {
			keys = append(keys, k)
		}
		if len(k.Links()) == 1 {
			deadEnds = append(deadEnds, k)
		}
	}
	if len(deadEnds) > 0 {
		return keys[utils.Random(0, len(deadEnds))]
	}
	return keys[utils.Random(0, len(keys))]
}

// RandomUnvisitedLink returns a random cell linked to this one that has not been visited
func (c *Cell) RandomUnvisitedLink(client string) *Cell {
	var keys []*Cell
	for item := range c.links.Iter() {
		linked := c.Linked(item.Key)
		if linked && !item.Key.Visited(client) {
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

// AllNeighbors returns a list of all cells that are neighbors (includes diagonals)
// Used for game of life only
func (c *Cell) AllNeighbors() []*Cell {
	c.RLock()
	defer c.RUnlock()

	var n []*Cell

	for _, cell := range []*Cell{c.North(), c.South(), c.East(), c.West()} {
		if cell != nil {
			n = append(n, cell)
		}
	}

	if c.North() != nil {
		for _, cell := range []*Cell{c.North().East(), c.North().West()} {
			if cell != nil {
				n = append(n, cell)
			}
		}
	}

	if c.South() != nil {
		for _, cell := range []*Cell{c.South().East(), c.South().West()} {
			if cell != nil {
				n = append(n, cell)
			}
		}
	}
	return n
}

// Neighbors returns a list of all cells that are neighbors (weather connected by passage or not)
func (c *Cell) Neighbors() []*Cell {
	c.RLock()
	defer c.RUnlock()

	var n []*Cell

	for _, cell := range []*Cell{c.north, c.south, c.east, c.west} {
		if cell != nil {
			n = append(n, cell)
		}
	}

	// if weaving is allowed, add additional possibilities for neighbors
	if c.config.AllowWeaving && utils.Random(0, 100) <= int(c.config.WeavingProbability*100) {
		if c.canTunnelNorth() {
			n = append(n, c.north.North())
		}
		if c.canTunnelSouth() {
			n = append(n, c.south.South())
		}
		if c.canTunnelEast() {
			n = append(n, c.east.East())
		}
		if c.canTunnelWest() {
			n = append(n, c.west.West())
		}
	}

	return n
}

// RandomNeighbor returns a random neighbor of this cell
func (c *Cell) RandomNeighbor() *Cell {
	c.RLock()
	defer c.RUnlock()

	n := c.Neighbors()

	return n[utils.Random(0, len(n))]
}

// RandomAllNeighbor returns a random neighbor of this cell (including diagonals)
func (c *Cell) RandomAllNeighbor() *Cell {
	c.RLock()
	defer c.RUnlock()

	n := c.AllNeighbors()

	return n[utils.Random(0, len(n))]
}

// GetFacingDirection returns the direction walker was facing when moving to toCell from this cell
// north, south, east, west
func (c *Cell) GetFacingDirection(toCell *Cell) string {
	c.RLock()
	defer c.RUnlock()

	facing := ""

	if c.North() == toCell {
		facing = "north"
	}
	if c.East() == toCell {
		facing = "east"
	}
	if c.West() == toCell {
		facing = "west"
	}
	if c.South() == toCell {
		facing = "south"
	}
	return facing
}

// Orphan isolates the cell from all of its neighbors
func (c *Cell) Orphan() {
	if c.East() != nil {
		c.East().SetWest(nil)
	}
	if c.West() != nil {
		c.West().SetEast(nil)
	}
	if c.North() != nil {
		c.North().SetSouth(nil)
	}
	if c.South() != nil {
		c.South().SetNorth(nil)
	}

	c.SetOrphan()
}

// SetOrphan ...
func (c *Cell) SetOrphan() {
	c.Lock()
	defer c.Unlock()

	c.orphan = true
}

// IsOrphan ...
func (c *Cell) IsOrphan() bool {
	c.RLock()
	defer c.RUnlock()

	return c.orphan
}

func (c *Cell) isHorizontalPassage() bool {
	return c.Linked(c.East()) && c.Linked(c.West()) && !c.Linked(c.North()) && !c.Linked(c.South())
}

func (c *Cell) isVerticalPassage() bool {
	return !c.Linked(c.East()) && !c.Linked(c.West()) && c.Linked(c.North()) && c.Linked(c.South())
}

func (c *Cell) canTunnelNorth() bool {
	return c.North() != nil && c.North().North() != nil && c.North().isHorizontalPassage()
}

func (c *Cell) canTunnelSouth() bool {
	return c.South() != nil && c.South().South() != nil && c.South().isHorizontalPassage()
}

func (c *Cell) canTunnelEast() bool {
	return c.East() != nil && c.East().East() != nil && c.East().isVerticalPassage()
}

func (c *Cell) canTunnelWest() bool {
	return c.West() != nil && c.West().West() != nil && c.West().isVerticalPassage()
}
