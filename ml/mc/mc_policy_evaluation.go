package mc

import (
	"math"
	"mazes/maze"
	"mazes/ml/dp"

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

type episode struct {
	state  int
	action int
	reward float64
}

type Policy struct {
}

// Prediction returns the value function for the given policy
func Prediction(m *maze.Maze, policy *dp.Policy, numEpisodes int) (*dp.ValueFunction, error) {

	numStates := int(m.Config().Columns * m.Config().Rows)

	// map of state -> sum of all returns in that state
	returnsSum := make(map[int]float64)
	// map of state -> count of all returns
	returnsCount := make(map[int]float64)

	vf := dp.NewValueFunction(numStates)
	episodes := []episode{}

	// run through the policy this many times
	// each run is a walk through the maze until the end (or limit?)
	for e := 1; e <= numEpisodes; e++ {
		printProgress(e, numEpisodes)

		// generate an episode (wonder through the maze following policy)
		// An episode is an array of (state, action, reward) tuples

		// pick a random state to start at (fromCell and action)
		state := int64(utils.Random(0, numStates))
		fromCell, err := utils.LocationFromState(m.Config().Rows, m.Config().Columns, state)
		if err != nil {
			return nil, err
		}

		// get the action, according to policy, for this state
		action := policy.BestRandomActionsForState(state)

	}
	return vf
}
