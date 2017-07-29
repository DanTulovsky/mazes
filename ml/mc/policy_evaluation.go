package mc

import (
	"math"
	"mazes/maze"
	"mazes/ml"

	"mazes/utils"

	"fmt"

	"sort"

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
func RunEpisode(m *maze.Maze, p *ml.Policy, clientID string, toCell *maze.Cell) (e episode, err error) {
	// log.Printf("\nStarting episode...\n")
	numStates := int(m.Config().Columns * m.Config().Rows)

	// pick a random state to start at (fromCell), toCell is always the same
	state := int(utils.Random(0, numStates))
	fromCell, err := utils.LocationFromState(m.Config().Rows, m.Config().Columns, int64(state))
	if err != nil {
		return e, err
	}

	c, err := m.Client(clientID)
	cell, err := m.CellFromLocation(fromCell)
	if err != nil {
		return e, err
	}
	c.SetCurrentLocation(cell)
	// log.Printf("initial state: %v, initial location: %v", state, c.CurrentLocation().Location())

	solved := false

	for !solved {
		// get the action, according to policy, for this state
		action := p.BestRandomActionsForState(state)
		// log.Printf("state: %v; action: %v", state, ml.ActionToText[action])

		// get the next state
		nextState, reward, valid, err := ml.NextState(m, toCell.Location(), state, action)
		if err != nil {
			return e, err
		}
		// log.Printf("nextState: %v, reward: %v, valid: %v, err: %v", nextState, reward, valid, err)

		sr := stateReturn{
			state:  state,
			action: action,
			reward: reward,
		}
		e.sr = append(e.sr, sr)

		if utils.LocsSame(c.CurrentLocation().Location(), toCell.Location()) {
			solved = true
		}

		if valid && action != ml.None && !solved {
			// only actually move if we picked a valid direction, otherwise we stay in the same place
			// log.Printf("moving: %v", ml.ActionToText[action])
			c, err = m.MoveClient(clientID, ml.ActionToText[action])
			if err != nil {
				return e, err
			}
		}

		state = nextState
		// log.Printf("current location: %v, current state: %v", c.CurrentLocation().Location(), state)

	}

	return e, err
}

// statesInEpisode returns a list of all the states found in the episode
func statesInEpisode(e episode) []int {
	statesMap := make(map[int]bool)
	for _, sr := range e.sr {
		statesMap[sr.state] = true
	}

	states := make([]int, 0, len(statesMap))
	for k := range statesMap {
		states = append(states, k)
	}

	sort.Ints(states)
	return states
}

// firstStateInEpisodeIdx returns the first index of the state in the episode (first time we reached the state)
func firstStateInEpisodeIdx(e episode, state int) (int, error) {
	for idx, sr := range e.sr {
		if sr.state == state {
			return idx, nil
		}
	}
	return -1, fmt.Errorf("unable to find state: %v in episode: %v", state, e)
}

// sumRewardsSinceIdx returns the sume of all rewards since state at index idx
func sumRewardsSinceIdx(e episode, idx int, df float64) (float64, error) {
	if len(e.sr) <= idx {
		return 0, fmt.Errorf("idx (%v) is too large for size of e.sr (%v)", idx, len(e.sr))
	}
	var sum float64

	for i, sr := range e.sr[idx:] {
		sum = sum + sr.reward*math.Pow(df, float64(i))
	}
	return sum, nil
}

// Evaluate returns the value function for the given policy
func Evaluate(p *ml.Policy, m *maze.Maze, clientID string, numEpisodes int, df, theta float64, toCell *maze.Cell) (*ml.ValueFunction, error) {

	numStates := int(m.Config().Columns * m.Config().Rows)

	// map of state -> sum of all returns in that state
	returnsSum := make(map[int]float64)
	// map of state -> count of all returns
	returnsCount := make(map[int]float64)

	vf := ml.NewValueFunction(numStates)
	// episodes := []episode{}

	// run through the policy this many times
	// each run is a walk through the maze until the end (or limit?)
	for e := 1; e <= numEpisodes; e++ {
		printProgress(e, numEpisodes)

		// generate an episode (wonder through the maze following policy)
		// An episode is an array of (state, action, reward) tuples
		episode, err := RunEpisode(m, p, clientID, toCell)
		if err != nil {
			return nil, err
		}
		// log.Printf("e: %v", episode)
		// episodes = append(episodes, e)

		// Find all states the we've visited in this episode
		states := statesInEpisode(episode)
		// log.Printf("states in episode (%v): %v", e, states)

		for _, s := range states {
			// Find the first occurrence of the state in the episode
			stateIdx, err := firstStateInEpisodeIdx(episode, s)
			if err != nil {
				return nil, err
			}

			// Sum up all rewards since the first occurrence
			sum, err := sumRewardsSinceIdx(episode, stateIdx, df)
			if err != nil {
				return nil, err
			}

			// Calculate average return for this state over all sampled episodes
			if _, ok := returnsSum[s]; !ok {
				returnsSum[s] = 0
			}
			returnsSum[s] += sum

			if _, ok := returnsCount[s]; !ok {
				returnsCount[s] = 0
			}
			returnsCount[s]++

			vf.Set(s, returnsSum[s]/returnsCount[s])
		}

	}

	// .Printf("returnsSum: %v; returnsCount: %v", returnsSum, returnsCount)

	return vf, nil
}
