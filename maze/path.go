package maze

import (
	"mazes/colors"

	"github.com/sasha-s/go-deadlock"
	"github.com/veandco/go-sdl2/sdl"
)

// PathSegment is one segement of a path. A cell, and metadata.
type PathSegment struct {
	cell   *Cell
	facing string // when you came in, which way were you facing (north, south, east, west)
	deadlock.RWMutex
}

func NewSegment(c *Cell, f string) *PathSegment {
	return &PathSegment{cell: c, facing: f}
}

func (ps *PathSegment) Cell() *Cell {
	// no need to lock, this is never set after creation
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

// DrawPath draws the path as present in the cells
func (p *PathSegment) DrawPath(r *sdl.Renderer, g *Maze, solveCells map[*Cell]bool, isLast, isSolution bool) *sdl.Renderer {
	cell := p.Cell()
	pathWidth := cell.pathWidth
	PixelsPerCell := cell.width
	solvePath := g.SolvePath()

	getPathRect := func(d string, inSolution bool) *sdl.Rect {
		if !inSolution {
			pathWidth = cell.pathWidth / 2
		} else {
			pathWidth = cell.pathWidth
		}
		// these are the path segments from the middle towards the given direction
		paths := map[string]*sdl.Rect{
			"east": {
				int32(cell.column*PixelsPerCell + PixelsPerCell/2),
				int32(cell.row*PixelsPerCell + PixelsPerCell/2),
				int32(PixelsPerCell/2 + cell.wallWidth),
				int32(pathWidth)},
			"west": {
				int32(cell.column*PixelsPerCell + cell.wallWidth),
				int32(cell.row*PixelsPerCell + PixelsPerCell/2),
				int32(PixelsPerCell/2 + pathWidth - cell.wallWidth),
				int32(pathWidth)},
			"north": {
				int32(cell.column*PixelsPerCell + PixelsPerCell/2),
				int32(cell.row*PixelsPerCell + cell.wallWidth),
				int32(pathWidth),
				int32(PixelsPerCell/2 - cell.wallWidth)},
			"south": {
				int32(cell.column*PixelsPerCell + PixelsPerCell/2),
				int32(cell.row*PixelsPerCell + PixelsPerCell/2),
				int32(pathWidth),
				int32(PixelsPerCell/2 + cell.wallWidth)},
		}
		return paths[d]
	}

	pathColor := p.cell.pathColor
	if isSolution {
		pathColor = colors.SetOpacity(pathColor, 255) // solution is fully visible
	} else {
		pathColor = colors.SetOpacity(pathColor, 60) // travel path is less visible
	}

	colors.SetDrawColor(pathColor, r)
	currentSegmentInSolution := solvePath.SegmentInPath(p)

	if isLast && !cell.Visited() {
		switch p.Facing() {
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
		if cell.pathEast && cell.East != nil {
			// if current cell and neighbor is in the solution, solid color.
			eastInSolution := solvePath.CellInPath(cell.East)
			if eastInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("east", eastInSolution && currentSegmentInSolution))

		}
		if cell.pathWest && cell.West != nil {
			westInSolution := solvePath.CellInPath(cell.West)
			if westInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("west", westInSolution && currentSegmentInSolution))

		}
		if cell.pathNorth && p.cell.North != nil {
			northInSolution := solvePath.CellInPath(cell.North)
			if northInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("north", northInSolution && currentSegmentInSolution))

		}
		if cell.pathSouth && cell.South != nil {
			southInSolution := solvePath.CellInPath(cell.South)
			if southInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("south", southInSolution && currentSegmentInSolution))

		}
	}

	return r
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

func (p *Path) LastSegment() *PathSegment {
	p.RLock()
	defer p.RUnlock()

	if len(p.segments) == 0 {
		return nil
	}
	return p.segments[len(p.segments)-1]
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
