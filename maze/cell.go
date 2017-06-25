package maze

import (
	"container/heap"
	"fmt"
	"log"

	"mazes/colors"
	pb "mazes/proto"
	"mazes/utils"

	"github.com/sasha-s/go-deadlock"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_gfx"
)

const (
	VisitedGenerator = "generator"
)

// Cell defines a single cell in the grid
type Cell struct {
	x, y, z int64
	// keep track of neighborgs
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

func (c *Cell) SetBelow(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.below = cell
}

func (c *Cell) SetNorth(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.north = cell
}

func (c *Cell) SetSouth(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.south = cell
}

func (c *Cell) SetEast(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.east = cell
}

func (c *Cell) SetWest(cell *Cell) {
	c.Lock()
	defer c.Unlock()
	c.west = cell
}

func (c *Cell) Below() *Cell {
	c.RLock()
	defer c.RUnlock()
	return c.below
}

func (c *Cell) North() *Cell {
	c.RLock()
	defer c.RUnlock()
	return c.north
}

func (c *Cell) South() *Cell {
	c.RLock()
	defer c.RUnlock()
	return c.south
}

func (c *Cell) East() *Cell {
	c.RLock()
	defer c.RUnlock()
	return c.east
}

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

func (c *Cell) String() string {
	c.RLock()
	defer c.RUnlock()
	return fmt.Sprintf("(%v, %v, %v)", c.x, c.y, c.z)
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

	return &pb.MazeLocation{c.x, c.y, c.z}
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

			totalWeight := int(d) + l.weight // never changes once set

			prevDistance, err := c.distances.Get(l)

			if totalWeight < int(prevDistance) || err != nil {
				heap.Push(&pending, l)
				// sets distance to new cell
				c.distances.Set(l, int(totalWeight))
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
		gfx.StringRGBA(r, int(x), int(y), fmt.Sprintf("%v", c.Distance()), 0, 0, 0, 255)
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
	//c.RLock()
	//defer c.RUnlock()

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

func (c *Cell) linkOneWay(cell *Cell) {
	c.links.Insert(cell, true)
}

func (c *Cell) unLinkOneWay(cell *Cell) {
	c.links.Delete(cell)
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

// directionTo returns the direction (north, south, east, west) of cell from c
// c and cell must be linked
// TODO(dan): raise appropriate error if cells are not linked
func (c *Cell) directionTo(cell *Cell, client string) *pb.Direction {

	switch {
	case c.North() == cell:
		return &pb.Direction{"north", cell.Visited(client)}
	case c.South() == cell:
		return &pb.Direction{"south", cell.Visited(client)}
	case c.East() == cell:
		return &pb.Direction{"east", cell.Visited(client)}
	case c.West() == cell:
		return &pb.Direction{"west", cell.Visited(client)}
	}

	return &pb.Direction{"", false}
}

// DirectionLinks returns a list of directions that have linked (passage to) cells
func (c *Cell) DirectionLinks(client string) []*pb.Direction {
	var directions []*pb.Direction
	for item := range c.links.Iter() {
		if c.Linked(item.Key) {
			directions = append(directions, c.directionTo(item.Key, client))
		}
	}
	return directions
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

// Neighbors returns a list of all cells that are neighbors (weather connected by passage or not)
func (c *Cell) Neighbors() []*Cell {
	c.RLock()
	defer c.RUnlock()

	var n []*Cell

	for _, cell := range []*Cell{c.North(), c.South(), c.East(), c.West()} {
		if cell != nil {
			n = append(n, cell)
		}
	}

	// if weaving is allowed, add additional possibilities for neighbors
	if c.config.AllowWeaving && utils.Random(0, 100) <= int(c.config.WeavingProbability*100) {
		if c.canTunnelNorth() {
			n = append(n, c.North().North())
		}
		if c.canTunnelSouth() {
			n = append(n, c.South().South())
		}
		if c.canTunnelEast() {
			n = append(n, c.East().East())
		}
		if c.canTunnelWest() {
			n = append(n, c.West().West())
		}
	}

	return n
}

// RandomNeighbor returns a random neighbor of this cell
func (c *Cell) RandomNeighbor() *Cell {
	c.RLock()
	defer c.RUnlock()

	var n []*Cell

	for _, cell := range []*Cell{c.North(), c.South(), c.East(), c.West()} {
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

func (c *Cell) SetOrphan() {
	c.Lock()
	defer c.Unlock()

	c.orphan = true
}

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
