package maze

import (
	"fmt"
	"mazes/colors"
)

// Config defines the configuration parameters passed to the Grid
type Config struct {
	Rows                 int
	Columns              int
	AllowWeaving         bool
	CellWidth            int // cell width
	WallWidth            int
	WallSpace            int
	PathWidth            int
	MarkVisitedCells     bool
	DarkMode             bool // only show cells the solver has seen
	ShowDistanceValues   bool
	ShowDistanceColors   bool
	OrphanMask           []Location // these cells are turned off and are not part of the grid
	AvatarImage          string
	VisitedCellColor     colors.Color
	BgColor              colors.Color
	BorderColor          colors.Color
	WallColor            colors.Color
	PathColor            colors.Color
	CurrentLocationColor colors.Color
	FromCellColor        colors.Color
	ToCellColor          colors.Color
}

// CheckConfig makes sure the config is valid
func (c *Config) CheckConfig() error {

	if c.Rows <= 0 || c.Columns <= 0 {
		return fmt.Errorf("rows and columns must be > 0: %#v", c)
	}
	return nil
}
