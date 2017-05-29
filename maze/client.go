package maze

// Define a Client struct to keep track of the clients in the maze

type client struct {
	id         string
	solvePath  *Path
	travelPath *Path
}
