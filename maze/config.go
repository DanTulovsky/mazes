package maze

import (
	"fmt"
	"mazes/colors"
)

// Config defines the configuration parameters passed to the Grid
type Config struct {
	Rows                 int
	Columns              int
	CellWidth            int // cell width
	WallWidth            int
	PathWidth            int
	MarkVisitedCells     bool
	DarkMode             bool       // only show cells the solver has seen
	OrphanMask           []Location // these cells are turned off and are not part of the grid
	VisitedCellColor     colors.Color
	BgColor              colors.Color
	BorderColor          colors.Color
	WallColor            colors.Color
	PathColor            colors.Color
	CurrentLocationColor colors.Color
}

// CheckConfig makes sure the config is valid
func (c *Config) CheckConfig() error {

	if c.Rows <= 0 || c.Columns <= 0 {
		return fmt.Errorf("rows and columns must be > 0: %#v", c)
	}
	return nil
}