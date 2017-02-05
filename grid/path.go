package grid

import (
	"mazes/colors"

	"github.com/veandco/go-sdl2/sdl"
)

// Path is a path (ordered collection of cells) through the maze
type Path struct {
	segments []*PathSegment
}

func NewPath() *Path {
	return &Path{segments: make([]*PathSegment, 0)}
}

// PathSegment is one segement of a path. A cell, and metadata.
type PathSegment struct {
	cell   *Cell
	facing string // when you came in, which way were you facing (north, south, east, west)
}

func NewSegment(c *Cell, f string) *PathSegment {
	return &PathSegment{cell: c, facing: f}
}

func (ps *PathSegment) Cell() *Cell {
	return ps.cell
}

func (ps *PathSegment) Facing() string {
	return ps.facing
}

func (ps *PathSegment) UpdateFacingDirection(f string) {
	ps.facing = f
}

// DrawPath draws the path as present in the cells
func (p *PathSegment) DrawPath(r *sdl.Renderer, g *Grid, isLast, isSolution bool) *sdl.Renderer {
	pathWidth := p.Cell().pathWidth
	PixelsPerCell := p.Cell().width

	getPathRect := func(d string, inSolution bool) *sdl.Rect {
		if !inSolution {
			pathWidth = p.Cell().pathWidth / 2
		} else {
			pathWidth = p.Cell().pathWidth
		}
		// these are the path segments from the middle towards the given direction
		paths := map[string]*sdl.Rect{
			"east": {
				int32(p.Cell().column*PixelsPerCell + PixelsPerCell/2),
				int32(p.Cell().row*PixelsPerCell + PixelsPerCell/2),
				int32(PixelsPerCell/2 + p.Cell().wallWidth),
				int32(pathWidth)},
			"west": {
				int32(p.Cell().column*PixelsPerCell + p.Cell().wallWidth),
				int32(p.Cell().row*PixelsPerCell + PixelsPerCell/2),
				int32(PixelsPerCell/2 + pathWidth - p.Cell().wallWidth),
				int32(pathWidth)},
			"north": {
				int32(p.Cell().column*PixelsPerCell + PixelsPerCell/2),
				int32(p.Cell().row*PixelsPerCell + p.Cell().wallWidth),
				int32(pathWidth),
				int32(PixelsPerCell/2 - p.Cell().wallWidth)},
			"south": {
				int32(p.Cell().column*PixelsPerCell + PixelsPerCell/2),
				int32(p.Cell().row*PixelsPerCell + PixelsPerCell/2),
				int32(pathWidth),
				int32(PixelsPerCell/2 + p.Cell().wallWidth)},
		}
		return paths[d]
	}

	pathColor := p.Cell().pathColor
	if isSolution {
		pathColor = colors.SetOpacity(pathColor, 255) // solution is fully visible
	} else {
		pathColor = colors.SetOpacity(pathColor, 60) // travel path is less visible
	}

	colors.SetDrawColor(pathColor, r)
	currentSegmentInSolution := SegmentInPath(p, g.SolvePath)

	if isLast && !p.Cell().Visited() {
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
		if p.Cell().pathEast && p.Cell().East != nil {
			// if current cell and neighbor is in the solution, solid color.
			eastInSolution := CellInCellList(p.Cell().East, g.SolvePath.ListCells())
			if eastInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("east", eastInSolution && currentSegmentInSolution))

		}
		if p.Cell().pathWest && p.Cell().West != nil {
			westInSolution := CellInCellList(p.Cell().West, g.SolvePath.ListCells())
			if westInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("west", westInSolution && currentSegmentInSolution))

		}
		if p.Cell().pathNorth && p.Cell().North != nil {
			northInSolution := CellInCellList(p.Cell().North, g.SolvePath.ListCells())
			if northInSolution && currentSegmentInSolution {
				pathColor = colors.SetOpacity(pathColor, 255)
			} else {
				pathColor = colors.SetOpacity(pathColor, 60)
			}
			colors.SetDrawColor(pathColor, r)
			r.FillRect(getPathRect("north", northInSolution && currentSegmentInSolution))

		}
		if p.Cell().pathSouth && p.Cell().South != nil {
			southInSolution := CellInCellList(p.Cell().South, g.SolvePath.ListCells())
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
	p.segments = append(p.segments, s)

}

func (p *Path) AddSegements(s []*PathSegment) {
	for _, seg := range s {
		p.segments = append(p.segments, seg)
	}
}

func (p *Path) LastSegment() *PathSegment {
	if len(p.segments) == 0 {
		return nil
	}
	return p.segments[len(p.segments)-1]
}

// DelSegement removes the last segment from the path
func (p *Path) DelSegement() {
	p.segments = p.segments[:len(p.segments)-1]
}

func (p *Path) List() []*PathSegment {
	return p.segments
}

// Length returns the length of the path
func (p *Path) Length() int {
	return len(p.segments)
}

func (p *Path) ListCells() []*Cell {
	var cells []*Cell
	for _, s := range p.segments {
		cells = append(cells, s.Cell())
	}
	return cells
}
