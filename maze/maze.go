package maze

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/sasha-s/go-deadlock"
	"github.com/veandco/go-sdl2/sdl"
	"mazes/colors"
	pb "mazes/proto"
	"mazes/utils"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano()) // need to initialize the seed
}

// Location is x,y coordinate of a cell
type Location struct {
	X, Y, Z int
}

// Grid defines the maze grid
type Maze struct {
	id               string
	config           *pb.MazeConfig
	rows             int64
	columns          int64
	cells            [][]*Cell
	mazeCells        map[*Cell]bool // cells that are in the maze, not orphaned (for caching)
	orphanCells      map[*Cell]bool // cells that are orphaned (for caching)
	cellWidth        int64
	wallWidth        int64
	pathWidth        int64
	bgColor          colors.Color
	borderColor      colors.Color
	wallColor        colors.Color
	createTime       time.Duration    // how long it took to apply the algorithm to create the grid
	fromCell, toCell map[string]*Cell // save these for proper coloring, per client
	nextClient       int              // number of the next client to connect

	genCurrentLocation *Cell // the current location of generator

	// map of client IDs to client structures
	clients map[string]*client

	avatar *sdl.Texture

	deadlock.RWMutex
}

func (m *Maze) Config() *pb.MazeConfig {
	return m.config
}

// setupMazeMask reads in the mask image and creates the maze based on it.
// The size of the maze is the size of the image, in pixels.
// Any *black* pixel in the mask image becomes an orphan square.
func setupMazeMask(f string, c *pb.MazeConfig, mask []*pb.MazeLocation) ([]*pb.MazeLocation, error) {

	addToMask := func(mask []*pb.MazeLocation, x, y int64) ([]*pb.MazeLocation, error) {
		l := &pb.MazeLocation{x, y, 0}

		if x >= c.Columns || y >= c.Rows || x < 0 || y < 0 {
			return nil, fmt.Errorf("invalid cell passed to mask: %v (grid size: %v %v)", l, c.Columns, c.Rows)
		}

		mask = append(mask, l)
		return mask, nil
	}

	// read in image
	reader, err := os.Open(f)
	if err != nil {
		return nil, fmt.Errorf("failed to open mask image file: %v", err)
	}

	m, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}

	bounds := m.Bounds()
	c.Rows = int64(bounds.Max.Y)
	c.Columns = int64(bounds.Max.X)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := m.At(x, y).RGBA()
			// this only works for black, fix my colors to use the go image package colors
			if colors.Same(colors.GetColor("black"), colors.Color{uint8(r), uint8(g), uint8(b), uint8(a), ""}) {
				if mask, err = addToMask(mask, int64(x), int64(y)); err != nil {
					return nil, err
				}
			}

		}
	}
	return mask, nil
}

// NewMazeFromImage creates a new maze from the image at file f
func NewMazeFromImage(c *pb.MazeConfig, f string) (*Maze, error) {
	mask := make([]*pb.MazeLocation, 0)
	mask, err := setupMazeMask(f, c, mask)
	if err != nil {
		return nil, err
	}
	c.OrphanMask = mask

	return NewMaze(c)
}

// NewGrid returns a new grid.
func NewMaze(c *pb.MazeConfig) (*Maze, error) {
	//if err := c.CheckConfig(); err != nil {
	//	return nil, err
	//}

	m := &Maze{
		id:          c.GetId(),
		rows:        c.GetRows(),
		columns:     c.GetColumns(),
		cells:       [][]*Cell{},
		cellWidth:   c.GetCellWidth(),
		wallWidth:   c.GetWallWidth(),
		pathWidth:   c.GetPathWidth(),
		bgColor:     colors.GetColor(c.GetBgColor()),
		borderColor: colors.GetColor(c.GetBorderColor()),
		wallColor:   colors.GetColor(c.GetWallColor()),
		fromCell:    make(map[string]*Cell),
		toCell:      make(map[string]*Cell),

		config: c,

		mazeCells:   make(map[*Cell]bool),
		orphanCells: make(map[*Cell]bool),

		clients: make(map[string]*client),
	}

	m.prepareGrid()
	m.configureCells()

	return m, nil
}

func (m *Maze) configToCell(config *pb.ClientConfig, c string) (*Cell, error) {

	switch c {
	case "min":
		return m.SmallestCell(), nil
	case "max":
		return m.LargestCell(), nil
	case "random":
		return m.RandomCell(), nil
	default:
		from := strings.Split(c, ",")
		if len(from) != 2 {
			log.Fatalf("%v is not a valid coordinate", config.FromCell)
		}
		x, _ := strconv.ParseInt(from[0], 10, 64)
		y, _ := strconv.ParseInt(from[1], 10, 64)
		cell, err := m.Cell(x, y, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid fromCell: %v", err)
		}
		return cell, nil
	}

}

// AddClient adds a new client to the maze
func (m *Maze) AddClient(id string, config *pb.ClientConfig) error {
	log.Printf("adding client: %v", id)
	m.Lock()
	defer m.Unlock()

	m.clients[id] = &client{
		id: id,
		// currentLocation: m.fromCell,
		SolvePath:  NewPath(),
		TravelPath: NewPath(),
		config:     config,
		number:     m.nextClient,
	}

	m.nextClient++

	var fromCell, toCell *Cell
	var err error

	if config.GetFromCell() != "" {
		fromCell, err = m.configToCell(config, config.FromCell)
		if err != nil {
			return err
		}
	}

	if config.GetToCell() != "" {
		toCell, err = m.configToCell(config, config.ToCell)
		if err != nil {
			return err
		}
	}

	// solve the longest path
	if fromCell == nil || toCell == nil {
		log.Print("No fromCella and/or toCell set, defaulting to longestPath.")
		_, fromCell, toCell, _ = m.LongestPath()
	}

	m.SetFromCell(m.clients[id], fromCell)
	m.SetToCell(m.clients[id], toCell)

	log.Printf("Path: %v -> %v", fromCell, toCell)

	if m.clients[id].config.ShowFromToColors {
		m.SetFromToColors(m.clients[id], fromCell, toCell)
	}

	// this will color the maze based on the last client to register
	log.Printf("setting distance colors")
	if m.Config().GetShowDistanceColors() {
		m.SetDistanceInfo(m.clients[id], fromCell)
	}

	log.Printf("added client: %v", id)
	return nil
}

// Braid removes dead ends from the maze with a probability p (p = 1 means no dead ends)
func (m *Maze) Braid(p float64) {
	log.Printf("Removing dead ends with probability %v", p)

	for _, c := range m.DeadEnds() {
		if utils.Random(0, 100) >= int(p*100) {
			continue
		}

		// make sure still dead end
		if len(c.Links()) == 1 {
			n := c.RandomUnLinkPreferDeadends()
			m.Link(c, n)
		}

	}
}

// Client returns a single client
func (m *Maze) Client(id string) (*client, error) {
	if c, found := m.clients[id]; found {
		return c, nil
	}
	return nil, fmt.Errorf("client [%v] not found", id)
}

// Clients returns the clients connected to this maze
func (m *Maze) Clients() map[string]*client {
	return m.clients
}

// Link links c1 to c2 to its neighbor (adds passage)
func (m *Maze) Link(c1, c2 *Cell) {
	if c1 == nil || c2 == nil {
		debug.PrintStack()
		log.Fatalf("failure linking %v to %v!", c1, c2)
	}

	var linkCell *Cell
	// if weaving, check if we need to link through a hidden cell
	if m.config.AllowWeaving {

		// is there a cell between this one and the link to cell?
		if c1.North() != nil && c2.South() != nil && c1.North() == c2.South() {
			linkCell = NewCell(c1.North().x, c1.North().y, c1.North().z-1, m.config) // under
			c1.North().SetBelow(linkCell)
			// rework neighbor links
			c1.SetNorth(linkCell)
			c2.SetSouth(linkCell)
			linkCell.SetSouth(c1)
			linkCell.SetNorth(c2)
		} else if c1.South() != nil && c2.North() != nil && c1.South() == c2.North() {
			linkCell = NewCell(c1.South().x, c1.South().y, c1.South().z-1, m.config) // under
			c1.South().SetBelow(linkCell)
			c1.SetSouth(linkCell)
			c2.SetNorth(linkCell)
			linkCell.SetSouth(c2)
			linkCell.SetNorth(c1)
		} else if c1.East() != nil && c2.West() != nil && c1.East() == c2.West() {
			linkCell = NewCell(c1.East().x, c1.East().y, c1.East().z-1, m.config) // under
			c1.East().SetBelow(linkCell)
			c1.SetEast(linkCell)
			c2.SetWest(linkCell)
			linkCell.SetEast(c2)
			linkCell.SetWest(c1)
		} else if c1.West() != nil && c2.East() != nil && c1.West() == c2.East() {
			linkCell = NewCell(c1.West().x, c1.West().y, c1.West().z-1, m.config) // under
			c1.West().SetBelow(linkCell)
			c1.SetWest(linkCell)
			c2.SetEast(linkCell)
			linkCell.SetEast(c1)
			linkCell.SetWest(c2)
		}

		if linkCell != nil {
			c1.linkOneWay(linkCell)
			linkCell.linkOneWay(c1)

			c2.linkOneWay(linkCell)
			linkCell.linkOneWay(c2)
		} else {
			c1.linkOneWay(c2)
			c2.linkOneWay(c1)
		}

	} else {
		c1.linkOneWay(c2)
		c2.linkOneWay(c1)
	}

}

//// loadAvatar reads in the avatar image
//func (m *Maze) loadAvatar(r *sdl.Renderer) {
//	if m.avatar != nil {
//		return
//	}
//
//	var err error
//	m.avatar, err = img.LoadTexture(r, m.config.AvatarImage)
//	if err != nil {
//		Fail(err)
//	}
//	m.avatar.SetBlendMode(sdl.BLENDMODE_BLEND)
//	m.avatar.SetAlphaMod(255)
//}
//
// getAvatar returns the avatar texture
func (m *Maze) getAvatar() *sdl.Texture {
	// TODO: Fix this to be per client
	return nil
	//if m.avatar == nil && m.config.AvatarImage != "" {
	//	Fail(errors.New("calling getAvatar() before loadVatar()"))
	//}
	//return m.avatar
}

// prepareGrid initializes the grid with cells
func (m *Maze) prepareGrid() {
	m.Lock()
	defer m.Unlock()

	z := int64(0)
	m.cells = make([][]*Cell, m.columns)

	for x := int64(0); x < m.columns; x++ {
		m.cells[x] = make([]*Cell, m.rows)

		for y := int64(0); y < m.rows; y++ {
			m.cells[x][y] = NewCell(x, y, z, m.config)
		}
	}
}

// configureCells configures cells with their neighbors
func (m *Maze) configureCells() {
	m.Lock()
	defer m.Unlock()

	z := int64(0)

	for x := int64(0); x < m.columns; x++ {
		for y := int64(0); y < m.rows; y++ {
			cell, err := m.Cell(x, y, z)
			if err != nil {
				log.Fatalf("failed to initialize grid: %v", err)
			}
			var c *Cell
			// error is ignored, we just set nil if there is no neighbor
			c, _ = m.Cell(x, y-1, z)
			cell.SetNorth(c)

			c, _ = m.Cell(x, y+1, z)
			cell.SetSouth(c)

			c, _ = m.Cell(x-1, y, z)
			cell.SetWest(c)

			c, _ = m.Cell(x+1, y, z)
			cell.SetEast(c)

		}
	}

	for _, o := range m.config.GetOrphanMask() {
		cell, err := m.Cell(o.X, o.Y, 0)
		if err != nil {
			Fail(err)
		}
		cell.Orphan()
	}

}

// SetCurrentLocation sets the current cell location of the generator algorithm
func (m *Maze) SetGenCurrentLocation(cell *Cell) {
	m.Lock()
	defer m.Unlock()
	m.genCurrentLocation = cell
}

// GenCurrentLocation returns the current cell location of the generator algorithm
func (m *Maze) GenCurrentLocation() *Cell {
	m.RLock()
	defer m.RUnlock()

	return m.genCurrentLocation
}

func (m *Maze) SetCreateTime(t time.Duration) {
	m.Lock()
	defer m.Unlock()

	m.createTime = t
}

func (m *Maze) CreateTime() time.Duration {
	m.RLock()
	defer m.RUnlock()

	return m.createTime
}

// Dimensions returns the dimensions of the grid.
func (m *Maze) Dimensions() (int64, int64) {
	// No lock, does not change
	return m.columns, m.rows
}

func (m *Maze) String() string {
	m.RLock()
	defer m.RUnlock()

	output := "  "
	for x := int64(0); x < m.columns; x++ {
		output = fmt.Sprintf("%v%4v", output, x)
	}

	output = fmt.Sprintf("\n%v\n   ┌", output)
	for x := int64(0); x < m.columns-1; x++ {
		output = fmt.Sprintf("%v───┬", output)
	}
	output = output + "───┐" + "\n"

	for y := int64(0); y < m.rows; y++ {
		top := fmt.Sprintf("%-3v│", y)
		bottom := "   ├"

		for x := int64(0); x < m.columns; x++ {
			cell, err := m.Cell(x, y, 0)
			if err != nil {
				continue
			}
			body := "   "
			east_boundary := " "
			if !cell.Linked(cell.East()) {
				east_boundary = "│"
			}
			top = fmt.Sprintf("%v%v%v", top, body, east_boundary)

			south_boundary := "   "
			if !cell.Linked(cell.South()) {
				south_boundary = "───"
			}
			corner := "┼"
			if x == m.columns-1 {
				corner = "┤" // right wall
			}
			if x == m.columns-1 && y == m.rows-1 {
				corner = "┘"
			}
			if x == 0 && y == m.rows-1 {
				bottom = "   └"
			}
			if x < m.columns-1 && y == m.rows-1 {
				corner = "┴"
			}
			bottom = fmt.Sprintf("%v%v%v", bottom, south_boundary, corner)
		}
		output = fmt.Sprintf("%v%v\n", output, top)
		output = fmt.Sprintf("%v%v\n", output, bottom)
	}

	return output
}

func (m *Maze) FromCell(client *client) *Cell {
	m.RLock()
	defer m.RUnlock()

	if c, ok := m.fromCell[client.id]; ok {
		return c
	}
	return nil
}

func (m *Maze) SetFromCell(client *client, c *Cell) {
	// TODO: switch to safemap
	m.fromCell[client.id] = c
}

func (m *Maze) ToCell(client *client) *Cell {
	m.RLock()
	defer m.RUnlock()

	if c, ok := m.toCell[client.id]; ok {
		return c
	}

	return nil
}

func (m *Maze) SetToCell(client *client, c *Cell) {
	// TODO: switch to safe map
	m.toCell[client.id] = c
}

// DrawMazeBackground renders the gui maze background in memory
func (m *Maze) DrawMazeBackground(r *sdl.Renderer) *sdl.Renderer {
	// defer utils.TimeTrack(time.Now(), "DrawMaze")

	//fromCell := m.FromCell()
	//
	//// If saved, draw distance colors
	//if fromCell != nil && m.config.ShowDistanceColors {
	//	m.SetDistanceInfo(fromCell)
	//}

	// Each cell draws its background, half the wall as well as anything inside it
	for x := int64(0); x < m.columns; x++ {
		for y := int64(0); y < m.rows; y++ {
			cell, err := m.Cell(x, y, 0)
			if err != nil {
				Fail(fmt.Errorf("Error drawing cell (%v, %v): %v", x, y, err))
			}

			if cell.IsOrphan() {
				// these are cells not connected to the maze
				continue
			}

			// draw the below cell if it exists
			if cell.Below() != nil {
				// cell exists
				cell.Below().Draw(r)
			}

			cell.Draw(r)

		}
	}

	// Draw outside border
	// m.drawBorder(r)

	// Load avatar if needed
	//if m.config.AvatarImage != "" {
	//	m.loadAvatar(r)
	//}

	return r
}

// Draw renders the gui maze in memory, display by calling Present
func (m *Maze) DrawMaze(r *sdl.Renderer, bg *sdl.Texture) *sdl.Renderer {
	// defer utils.TimeTrack(time.Now(), "DrawMaze")

	if bg != nil {
		r.Copy(bg, nil, nil) // copy the background texture
	}

	// Draw location of the generator algorithm
	m.drawGenCurrentLocation(r)

	// Draw the path and location of solver
	clients := m.Clients()
	for _, client := range clients {
		m.drawPath(r, client, m.config.MarkVisitedCells)
	}

	return r
}

// DrawBorder renders the maze border in memory, display by calling Present
func (m *Maze) drawBorder(r *sdl.Renderer) *sdl.Renderer {
	// defer utils.TimeTrack(time.Now(), "drawBorder")

	colors.SetDrawColor(m.borderColor, r)

	var bg sdl.Rect
	var rects []sdl.Rect
	winWidth := int32(m.columns*m.cellWidth + m.wallWidth*2)
	winHeight := int32(m.rows*m.cellWidth + m.wallWidth*2)
	wallWidth := int32(m.wallWidth)

	// top
	bg = sdl.Rect{0, 0, winWidth, wallWidth}
	rects = append(rects, bg)

	// left
	bg = sdl.Rect{0, 0, wallWidth, winHeight}
	rects = append(rects, bg)

	// bottom
	bg = sdl.Rect{0, winHeight - wallWidth, winWidth, wallWidth}
	rects = append(rects, bg)

	// right
	bg = sdl.Rect{winWidth - wallWidth, 0, wallWidth, winHeight}
	rects = append(rects, bg)

	if err := r.FillRects(rects); err != nil {
		Fail(fmt.Errorf("error drawing border: %v", err))
	}
	return r
}

func (m *Maze) drawGenCurrentLocation(r *sdl.Renderer) *sdl.Renderer {

	current_location := m.GenCurrentLocation()

	if current_location != nil {
		for cell := range m.Cells() {
			if cell != nil {
				// reset all colors to default
				cell.SetBGColor(colors.GetColor("white"))
			}
		}

		current_location.SetBGColor(colors.GetColor("yellow"))
	}
	return r
}

// DrawPath renders the gui maze path in memory, display by calling Present
func (m *Maze) drawPath(r *sdl.Renderer, client *client, markVisited bool) *sdl.Renderer {
	// defer utils.TimeTrack(time.Now(), "drawPath")
	travelPath := client.TravelPath
	solvePath := client.SolvePath

	alreadyDone := make(map[*PathSegment]bool)

	var isSolution bool
	var isLast bool

	pathLength := travelPath.Length()

	travelPath.RLock()
	defer travelPath.RUnlock()

	for x, segment := range travelPath.segments {
		cell := segment.Cell()

		if x == pathLength-1 {
			isLast = true // last segment is drawn slightly different
		}

		if _, ok := alreadyDone[segment]; ok {
			if isLast {
				segment.DrawCurrentLocation(r, m.getAvatar())
			}
			continue
		}

		// cache state of this cell
		alreadyDone[segment] = true

		if solvePath.SegmentInPath(segment) {
			isSolution = true
		} else {
			isSolution = false
		}

		// TODO(dan): Change this to draw path between any two cells
		segment.DrawPath(r, m, client, solvePath, isLast, isSolution) // solution is colored by a different color

		if markVisited {
			cell.DrawVisited(r, client)
		}

		if isLast {
			segment.DrawCurrentLocation(r, m.getAvatar())
		}

	}

	return r
}

// Cell returns the cell at r,c
func (m *Maze) Cell(column, row, z int64) (*Cell, error) {
	if column < 0 || column >= m.columns || row < 0 || row >= m.rows {
		return nil, fmt.Errorf("(%v, %v) is outside the grid", column, row)
	}
	return m.cells[column][row], nil
}

func CellMapKeys(m map[*Cell]bool) []*Cell {
	var keys []*Cell
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// RandomCell returns a random cell out of all non-orphaned cells
func (m *Maze) RandomCell() *Cell {
	cells := CellMapKeys(m.Cells())
	// cells := m.OrderedCells()
	// cells := CellMapKeys(m.getMazeCells())

	return cells[utils.Random(0, len(cells))]
}

// RandomCellFromList returns a random cell from the provided list of cells
func (g *Maze) RandomCellFromList(cells []*Cell) *Cell {
	return cells[utils.Random(0, len(cells))]
}

// Size returns the number of cells in the grid
func (m *Maze) Size() int64 {
	// No lock, does not change
	return m.columns * m.rows
}

// Rows returns a list of rows (essentially the grid) - excluding the orphaned cells
func (m *Maze) Rows() [][]*Cell {
	rows := [][]*Cell{}

	for y := m.rows - 1; y >= 0; y-- {
		cells := []*Cell{}
		for x := m.columns - 1; x >= 0; x-- {
			cell, _ := m.Cell(x, y, 0)
			if !cell.IsOrphan() {
				cells = append(cells, cell)
			}
		}
		rows = append(rows, cells)
	}
	return rows
}

// OrderedCells returns the cells in the maze in a deterministic order, excludes orphaned cells
func (m *Maze) OrderedCells() []*Cell {
	cells := make([]*Cell, 0)

	for y := m.rows - 1; y >= 0; y-- {
		for x := m.columns - 1; x >= 0; x-- {
			cell, _ := m.Cell(x, y, 0)
			if !cell.IsOrphan() {
				cells = append(cells, cell)

				if cell.Below() != nil {
					cells = append(cells, cell.Below())
				}
			}
		}
	}
	return cells
}

// LargestCell returns the "largest" cell that is in the grid (not orphaned)
// Used as "max" value in genmaze.go
func (m *Maze) LargestCell() *Cell {
	for y := m.rows - 1; y >= 0; y-- {
		for x := m.columns - 1; x >= 0; x-- {
			cell := m.cells[x][y]
			if !cell.IsOrphan() {
				return cell
			}
		}
	}

	return nil
}

// SmallestCell returns the "smallest" cell that is in the grid (not orphaned)
// Used as "min" value in genmaze.go
func (m *Maze) SmallestCell() *Cell {
	var c *Cell

	for y := m.rows - 1; y >= 0; y-- {
		for x := m.columns - 1; x >= 0; x-- {
			cell := m.cells[x][y]
			if !cell.IsOrphan() {
				c = cell

			}
		}
	}

	return c
}

// Cells returns a list of un-orphaned cells in the grid
func (m *Maze) Cells() map[*Cell]bool {
	cells := make(map[*Cell]bool)
	for y := m.rows - 1; y >= 0; y-- {
		for x := m.columns - 1; x >= 0; x-- {
			cell := m.cells[x][y]
			if !cell.IsOrphan() {
				cells[cell] = true

				if cell.Below() != nil {
					if !cell.Below().IsOrphan() {
						cells[cell.Below()] = true
					}
				}
			}

		}
	}

	// cache
	m.setMazeCells(cells)
	return cells
}

func (m *Maze) getMazeCells() map[*Cell]bool {
	m.RLock()
	defer m.RUnlock()

	return m.mazeCells
}

func (m *Maze) setMazeCells(cells map[*Cell]bool) {
	m.Lock()
	defer m.Unlock()

	m.mazeCells = cells
}

// OrphanCells returns a list of orphan cells in the grid
func (m *Maze) OrphanCells() map[*Cell]bool {
	orphanCells := m.getOrphanMazeCells()

	if len(orphanCells) != 0 {
		return orphanCells
	}

	cells := make(map[*Cell]bool)
	for y := int64(0); y < m.rows; y++ {
		for x := int64(0); x < m.columns; x++ {
			cell := m.cells[x][y]
			if cell.IsOrphan() {
				cells[cell] = true

				if cell.Below() != nil {
					if cell.Below().IsOrphan() {
						cells[cell.Below()] = true
					}
				}
			}
		}
	}

	m.setOrphanMazeCells(cells)
	return cells
}

func (m *Maze) getOrphanMazeCells() map[*Cell]bool {
	m.RLock()
	defer m.RUnlock()

	return m.orphanCells
}

func (m *Maze) setOrphanMazeCells(cells map[*Cell]bool) {
	m.Lock()
	defer m.Unlock()
	m.orphanCells = cells
}

// UnvisitedCells returns a list of unvisited cells in the grid
func (m *Maze) UnvisitedCells(client string) []*Cell {
	cells := []*Cell{}

	for cell := range m.Cells() {
		if !cell.Visited(client) {
			cells = append(cells, cell)
		}
	}

	return cells
}

// ConnectCells connects the list of cells in order by passageways
func (m *Maze) ConnectCells(cells []*Cell) {

	for x := 0; x < len(cells)-1; x++ {
		cell := cells[x]
		// no lock, does not change
		for _, n := range []*Cell{cell.North(), cell.South(), cell.East(), cell.West()} {
			if n == cells[x+1] {
				m.Link(cell, n)
				break
			}
		}
	}
}

// LongestPath returns the longest path through the maze
func (m *Maze) LongestPath() (dist int, fromCell, toCell *Cell, path *Path) {
	defer utils.TimeTrack(time.Now(), "LongestPath")

	// pick random starting point
	fromCell = m.RandomCell()

	// find furthest point
	furthest, _ := fromCell.FurthestCell()

	// now find the furthest point from that
	toCell, _ = furthest.FurthestCell()

	// now get the path
	dist, path = m.ShortestPath(furthest, toCell)

	return dist, furthest, toCell, path
}

// SetFromToColors sets the opacity of the from and to cells to be highly visible
func (m *Maze) SetFromToColors(client *client, fromCell, toCell *Cell) {
	// defer utils.TimeTrack(time.Now(), "SetFromToColors")

	if fromCell == nil || toCell == nil {
		log.Printf("not setting fromToColors on nil: from: %v, to: %v", fromCell, toCell)
		return
	}

	log.Printf("Setting fromToColors colors for %v", client.id)
	// Set path start and end colors
	fromCell.SetBGColor(colors.GetColor(client.config.GetFromCellColor()))
	toCell.SetBGColor(colors.GetColor(client.config.GetToCellColor()))

}

// SetPathFromTo sets the given path in the cells from fromCell to toCell
func (m *Maze) SetPathFromTo(fromCell *Cell, client *client) {

	path := client.TravelPath

	var prev, next *Cell
	for x := 0; x < path.Length(); x++ {
		if x > 0 {
			prev = path.segments[x-1].Cell()
		}

		if x < path.Length()-1 {
			next = path.segments[x+1].Cell()
		}

		path.segments[x].Cell().SetPaths(client, prev, next)
	}
}

// ShortestPath finds the shortest path from fromCell to toCell
func (m *Maze) ShortestPath(fromCell, toCell *Cell) (int, *Path) {
	defer utils.TimeTrack(time.Now(), "ShortestPath")

	if path := fromCell.PathTo(toCell); path != nil {
		return path.Length(), path
	}

	var path = NewPath()

	// Get all distances from this cell
	d := fromCell.Distances()
	toCellDist, _ := d.Get(toCell)

	current := toCell

	for current != d.root {
		var smallest int = math.MaxInt64
		var next *Cell
		for _, link := range current.Links() {
			dist, err := d.Get(link)
			if err != nil {
				continue
			}
			if dist < smallest {
				smallest = dist
				next = link
			}
		}
		segment := NewSegment(next, "north") // arbitrary facing
		path.AddSegement(segment)
		if next == nil {
			log.Fatalf("failed to find next cell from: %v", current)
		}
		current = next
	}

	// add toCell to path
	path.ReverseCells()
	segment := NewSegment(toCell, "north") // arbitrary facing
	path.AddSegement(segment)

	// record path for caching
	fromCell.SetPathTo(toCell, path)

	return toCellDist, path
}

// SetDistanceInfo sets distance info based on fromCell
func (m *Maze) SetDistanceInfo(client *client, c *Cell) {
	// defer utils.TimeTrack(time.Now(), "SetDistanceColors")
	// figure out the distances if needed
	c.Distances()

	_, longestPath := c.FurthestCell()
	log.Printf("longest: %v", longestPath)

	// use alpha blending, works for any color
	for cell := range m.Cells() {
		d, err := c.distances.Get(cell)
		if err != nil {
			log.Printf("failed to get distance from %v to %v: %v", c, cell, err)
			return
		}
		// dColor := d - int(cell.Weight()) // ignore weights when coloring distance

		if m.config.ShowDistanceColors {
			// decrease bridghtnessAdjustto make the longest cells brighter. max = 255 (good = 228)
			bridghtnessAdjust := 228
			adjustedColor := bridghtnessAdjust - utils.AffineTransform(float64(d), 0, float64(longestPath), 0, float64(bridghtnessAdjust))
			cell.SetBGColor(colors.OpacityAdjust(m.bgColor, adjustedColor))
		}

		cell.SetDistance(d)
	}

	m.SetFromCell(client, c)
}

// DeadEnds returns a list of cells that are deadends (only linked to one neighbor)
func (m *Maze) DeadEnds() []*Cell {
	var deadends []*Cell

	for cell := range m.Cells() {
		if len(cell.Links()) == 1 {
			deadends = append(deadends, cell)
		}
	}

	return deadends
}

// Reset resets vital maze stats for a new solver run
func (m *Maze) Reset() {
	m.resetVisited()
	m.resetDistances()
}

// resetVisited sets all cells to be unvisited by the generator
func (m *Maze) resetVisited() {
	for c := range m.Cells() {
		c.SetUnVisited(VisitedGenerator)
	}
}

// resetDistances resets distances on all cells
func (m *Maze) resetDistances() {
	for c := range m.Cells() {
		c.SetDistance(0)
	}
}

// GetFacingDirection returns the direction walker was facing when moving fromCell -> toCell
// north, south, east, west
func (m *Maze) GetFacingDirection(fromCell, toCell *Cell) string {
	facing := ""

	if fromCell.North() == toCell {
		facing = "north"
	}
	if fromCell.East() == toCell {
		facing = "east"
	}
	if fromCell.West() == toCell {
		facing = "west"
	}
	if fromCell.South() == toCell {
		facing = "south"
	}
	return facing
}
