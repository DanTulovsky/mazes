package maze

import (
	"fmt"
	"time"

	"mazes/colors"
	"mazes/utils"

	"github.com/rcrowley/go-metrics"
	"github.com/sasha-s/go-deadlock"
	"github.com/veandco/go-sdl2/sdl"
)

// Path is a path (ordered collection of cells) through the maze
type Path struct {
	segments   []*PathSegment
	segmentMap map[*PathSegment]bool
	cellMap    map[*Cell]bool
	deadlock.RWMutex
}

func NewPath() *Path {
	return &Path{
		segments:   make([]*PathSegment, 0),
		segmentMap: make(map[*PathSegment]bool),
		cellMap:    make(map[*Cell]bool),
	}
}

func (p *Path) Segments() []*PathSegment {
	p.RLock()
	defer p.RUnlock()
	return p.segments
}

// SegmentInSegmentList returns true if segment is in path
func (p *Path) SegmentInPath(segment *PathSegment) bool {
	p.RLock()
	defer p.RUnlock()

	if _, ok := p.segmentMap[segment]; ok {
		return true
	}
	return false
}

func (p *Path) ReverseCells() {
	p.Lock()
	defer p.Unlock()

	for i, j := 0, len(p.segments)-1; i < j; i, j = i+1, j-1 {
		p.segments[i], p.segments[j] = p.segments[j], p.segments[i]
	}
}
func (p *Path) AddSegement(s *PathSegment) {
	p.Lock()
	defer p.Unlock()

	p.segments = append(p.segments, s)
	p.segmentMap[s] = true
	p.cellMap[s.Cell()] = true

}

func (p *Path) AddSegements(s []*PathSegment) {
	for _, seg := range s {
		p.Lock()
		defer p.Unlock()

		p.segments = append(p.segments, seg)
		p.segmentMap[seg] = true
		p.cellMap[seg.Cell()] = true
	}
}

func (p *Path) CellInPath(c *Cell) bool {
	p.RLock()
	defer p.RUnlock()

	if _, ok := p.cellMap[c]; ok {
		return true
	}
	return false
}

// LastNSegments returns the last N segment in the path, -1 means return all of them
func (p *Path) LastNSegments(n int64) []*PathSegment {
	p.RLock()
	defer p.RUnlock()

	if len(p.segments) == 0 {
		return nil
	}

	if n == -1 || n > int64(len(p.segments)) {
		return p.segments
	}

	return p.segments[int64(len(p.segments))-n : len(p.segments)]
}

// LastSegment returns the last segment in the path, this is the one the client is standing on
func (p *Path) LastSegment() *PathSegment {
	p.RLock()
	defer p.RUnlock()

	if len(p.segments) == 0 {
		return nil
	}
	return p.segments[len(p.segments)-1]
}

// PreviousSegmentinSolution returns the last segment that is in the solution and that is not the segment the client is current on
// this is the one the client came from
func (p *Path) PreviousSegmentinSolution() *PathSegment {
	p.RLock()
	defer p.RUnlock()

	if len(p.segments) <= 1 {
		return nil
	}

	last := p.LastSegment().Cell() // where we are at

	// walk back through the path until you hit the second solution == true
	for x := len(p.segments) - 1; x >= 0; x-- {
		if p.segments[x].solution {
			if last == p.segments[x].Cell() {
				p.segments[x].solution = false
				continue
			}
			return p.segments[x]
		}
	}

	return nil
}

// DelSegement removes the last segment from the path
func (p *Path) DelSegement() {
	p.Lock()
	defer p.Unlock()

	seg := p.segments[len(p.segments)-1]

	delete(p.segmentMap, seg)
	delete(p.cellMap, seg.Cell())
	p.segments = p.segments[:len(p.segments)-1]
}

func (p *Path) List() []*PathSegment {
	p.RLock()
	defer p.RUnlock()

	return p.segments
}

// Length returns the length of the path
func (p *Path) Length() int {
	p.RLock()
	defer p.RUnlock()

	return len(p.segments)
}

// ListCells returns a map containing all the cells in the path
func (p *Path) ListCells() map[*Cell]bool {
	p.Lock()
	defer p.Unlock()

	return p.cellMap
}

// Draw draws the path
func (p *Path) Draw(r *sdl.Renderer, client *client, avatar *sdl.Texture) {
	alreadyDone := make(map[*PathSegment]bool)

	metrics.GetOrRegisterGauge("maze.path.tavel.length", nil).Update(int64(p.Length()))

	for _, segment := range p.LastNSegments(client.config.GetDrawPathLength()) {

		if _, ok := alreadyDone[segment]; ok {
			continue
		}

		// cache state of this cell
		alreadyDone[segment] = true

		p.drawSegment(segment, r, client, false)

		if client.config.GetMarkVisitedCells() || client.config.GetNumberMarkVisitedCells() {
			segment.Cell().DrawVisited(r, client)
		}
	}

	// handle last segment
	if segment := p.LastSegment(); segment != nil {
		if client.config.GetDrawPathLength() != 0 {
			p.drawSegment(segment, r, client, true)
		}
		segment.DrawCurrentLocation(r, client, avatar)
	}

}

// drawSegment draws one segment of the path
func (p *Path) drawSegment(ps *PathSegment, r *sdl.Renderer, client *client, isLast bool) {
	t := metrics.GetOrRegisterTimer("maze.draw.path.segment.latency", nil)
	defer t.UpdateSince(time.Now())

	cell := ps.Cell()

	pathWidth := cell.pathWidth
	PixelsPerCell := cell.width

	var offset int32
	// offset client path based on the client.id
	if !client.config.GetDisableDrawOffset() {
		offset = int32(utils.DrawOffset(client.number)) * int32(pathWidth)
	}
	// TODO: Limit offset to fit inside cell

	getPathRect := func(d string, inSolution bool) *sdl.Rect {
		if !inSolution {
			pathWidth = cell.pathWidth / 2
		} else {
			pathWidth = cell.pathWidth
		}

		// these are the path segments from the middle towards the given direction
		paths := map[string]*sdl.Rect{
			"east": {
				int32(cell.x*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(cell.y*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(PixelsPerCell/2+cell.wallWidth) - offset,
				int32(pathWidth),
			},
			"west": {
				int32(cell.x*PixelsPerCell + cell.wallWidth),
				int32(cell.y*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(PixelsPerCell/2+pathWidth-cell.wallWidth) + offset,
				int32(pathWidth),
			},
			"north": {
				int32(cell.x*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(cell.y*PixelsPerCell + cell.wallWidth),
				int32(pathWidth),
				int32(PixelsPerCell/2-cell.wallWidth) + offset,
			},
			"south": {
				int32(cell.x*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(cell.y*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(pathWidth),
				int32(PixelsPerCell/2+cell.wallWidth) - offset,
			},
		}

		// stubs are for cells below other cells, we only draw a small part of the path
		stubs := map[string]*sdl.Rect{
			"east": {
				int32(cell.x*PixelsPerCell + PixelsPerCell + cell.wallWidth - cell.config.WallSpace/2),
				int32(cell.y*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(cell.config.WallSpace / 2),
				int32(pathWidth),
			},
			"west": {
				int32(cell.x*PixelsPerCell + cell.wallWidth),
				int32(cell.y*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(cell.config.WallSpace / 2),
				int32(pathWidth),
			},
			"north": {
				int32(cell.x*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(cell.y*PixelsPerCell + cell.wallWidth),
				int32(pathWidth),
				int32(cell.config.WallSpace / 2),
			},
			"south": {
				int32(cell.x*PixelsPerCell+PixelsPerCell/2) + offset,
				int32(cell.y*PixelsPerCell + PixelsPerCell + cell.wallWidth - cell.config.WallSpace/2),
				int32(pathWidth),
				int32(cell.config.WallSpace / 2),
			},
		}

		if l := cell.Location(); l.Z == -1 {
			return stubs[d]
		}

		return paths[d]
	}

	currentSegmentInSolution := ps.Solution()
	pathColor := colors.GetColor(client.config.GetPathColor())

	if currentSegmentInSolution {
		pathColor = colors.SetOpacity(pathColor, 255) // solution is fully visible
	} else {
		pathColor = colors.SetOpacity(pathColor, 60) // travel path is less visible
	}

	colors.SetDrawColor(pathColor, r)

	if isLast && !cell.Visited(client.id) {
		switch ps.Facing() {
		case "east":
			r.FillRect(getPathRect("west", currentSegmentInSolution))
		case "west":
			r.FillRect(getPathRect("east", currentSegmentInSolution))
		case "north":
			r.FillRect(getPathRect("south", currentSegmentInSolution))
		case "south":
			r.FillRect(getPathRect("north", currentSegmentInSolution))
		}

	} else {
		if cell.HavePath(client, "east") && cell.East() != nil {
			// if current cell and neighbor is in the solution, solid color.
			eastInSolution := p.CellInPath(cell.East())
			if eastInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("east", eastInSolution && currentSegmentInSolution))

		}
		if cell.HavePath(client, "west") && cell.West() != nil {
			westInSolution := p.CellInPath(cell.West())
			if westInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("west", westInSolution && currentSegmentInSolution))

		}
		if cell.HavePath(client, "north") && ps.cell.North() != nil {
			northInSolution := p.CellInPath(cell.North())
			if northInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("north", northInSolution && currentSegmentInSolution))

		}
		if cell.HavePath(client, "south") && cell.South() != nil {
			southInSolution := p.CellInPath(cell.South())
			if southInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("south", southInSolution && currentSegmentInSolution))

		}
	}

}

// PathSegment is one segement of a path. A cell, and metadata.
type PathSegment struct {
	cell     *Cell
	facing   string // when you came in, which way were you facing (north, south, east, west)
	solution bool
	deadlock.RWMutex
}

func NewSegment(c *Cell, f string, s bool) *PathSegment {
	return &PathSegment{cell: c, facing: f, solution: s}
}

func (ps *PathSegment) String() string {
	ps.RLock()
	defer ps.RUnlock()
	return fmt.Sprintf("%v (solution=%v; facing=%v)", ps.cell, ps.solution, ps.facing)
}

func (ps *PathSegment) Solution() bool {
	ps.RLock()
	defer ps.RUnlock()
	return ps.solution
}

func (ps *PathSegment) AddToSolution() {
	ps.Lock()
	defer ps.Unlock()
	ps.solution = true
}

func (ps *PathSegment) RemoveFromSolution() {
	ps.Lock()
	defer ps.Unlock()
	ps.solution = false
}

func (ps *PathSegment) Cell() *Cell {
	ps.RLock()
	defer ps.RUnlock()
	return ps.cell
}

func (ps *PathSegment) Facing() string {
	ps.RLock()
	defer ps.RUnlock()
	return ps.facing
}

func (ps *PathSegment) UpdateFacingDirection(f string) {
	ps.Lock()
	defer ps.Unlock()
	ps.facing = f
}

// DrawCurrentLocation marks the current location of the user
func (ps *PathSegment) DrawCurrentLocation(r *sdl.Renderer, client *client, avatar *sdl.Texture) {
	ps.Cell().DrawCurrentLocation(r, client, avatar, ps.Facing())
}
