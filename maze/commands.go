package maze

const (
	CommandListClients = iota
	CommandGetDirections
	CommandSetInitialClientLocation
	CommandMove
	CommandCurrentLocation
	CommandLocationInfo // current, from, to
)
