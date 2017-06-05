package maze

import "mazes/colors"

// Define a Client struct to keep track of the clients in the maze

type client struct {
	id              string
	number          int // the number of this client for the maze
	currentLocation *Cell
	pathColor       colors.Color
	SolvePath       *Path
	TravelPath      *Path
}

// SetCurrentLocation sets the client's current location
func (c *client) SetCurrentLocation(cell *Cell) {
	c.currentLocation = cell
}

// CurrentLocation returns the client's current location
func (c *client) CurrentLocation() *Cell {
	return c.currentLocation
}
