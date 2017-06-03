package maze

const (
	CommandListClients = iota
	CommandGetDirections
	CommandSetInitialClientLocation
	CommandMove
	CommandMoveBack
	CommandCurrentLocation
	CommandLocationInfo // current, from, to
)
