package maze

import pb "mazes/proto"

// Define a Client struct to keep track of the clients in the maze

type client struct {
	id              string
	number          int // the number of this client for the maze
	currentLocation *Cell
	config          *pb.ClientConfig
	TravelPath      *Path
	fromCell        *Cell
	toCell          *Cell
}

// UpdateClientViewAndLocation sets the client's current location
func (c *client) SetCurrentLocation(cell *Cell) {
	c.currentLocation = cell
}

// CurrentLocation returns the client's current location
func (c *client) CurrentLocation() *Cell {
	return c.currentLocation
}

func (c *client) FromCell() *Cell {
	return c.fromCell
}

func (c *client) ToCell() *Cell {
	return c.toCell
}
