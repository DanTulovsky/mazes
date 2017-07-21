package mc

import (
	"math"
	"mazes/maze"
	"mazes/ml"

	"mazes/utils"

	"github.com/buger/goterm"
)

func printProgress(e, numEpisodes int) {
	goterm.Clear()
	if math.Mod(float64(e), 1000) == 0 {
		goterm.Printf("\nEpisode %d of %d\n", e, numEpisodes)
		goterm.Flush()
	}
}

// stateReturn is the information for a single state
// an episode is a list of stateReturn structs
type stateReturn struct {
	state  int
	action int
	reward float64
}

type episode struct {
	sr []stateReturn
}

// RunEpisode runs through the maze once, following the policy.
// Returns a list of
func RunEpisode(m *maze.Maze, p *ml.Policy, clientID string) (e episode, err error) {
	numStates := int(m.Config().Columns * m.Config().Rows)

	// pick a random state to start at (fromCell and action)
	state := int64(utils.Random(0, numStates))
	fromCell, err := utils.LocationFromState(m.Config().Rows, m.Config().Columns, state)
	if err != nil {
		return nil, err
	}

	c, err := m.Client(clientID)
	c.SetCurrentLocation(fromCell)

	solved := false

	for !solved {
		// get the action, according to policy, for this state
		action := p.BestRandomActionsForState(int(state))
	}

	return e, err
}

// Evaluate returns the value function for the given policy
func Evaluate(p *ml.Policy, m *maze.Maze, clientID string, numEpisodes int, df, theta float64) (*ml.ValueFunction, error) {

	numStates := int(m.Config().Columns * m.Config().Rows)

	// map of state -> sum of all returns in that state
	returnsSum := make(map[int]float64)
	// map of state -> count of all returns
	returnsCount := make(map[int]float64)

	vf := ml.NewValueFunction(numStates)
	episodes := []episode{}

	// run through the policy this many times
	// each run is a walk through the maze until the end (or limit?)
	for e := 1; e <= numEpisodes; e++ {
		printProgress(e, numEpisodes)

		// generate an episode (wonder through the maze following policy)
		// An episode is an array of (state, action, reward) tuples
		e, err := RunEpisode(m, p, clientID)
		if err != nil {
			return nil, err
		}
		episodes = append(episodes, e)

	}
	return vf, nil
}
