package maze

// Define a Client struct to keep track of the clients in the maze

type client struct {
	id              string
	currentLocation *Cell
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
