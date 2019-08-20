package mc

import (
	"math"

	"gogs.wetsnow.com/dant/mazes/maze"
	"gogs.wetsnow.com/dant/mazes/ml"

	"fmt"

	pb "gogs.wetsnow.com/dant/mazes/proto"
	"gogs.wetsnow.com/dant/mazes/utils"

	"sort"
)

func printProgress(e, numEpisodes int64, epsilon, delta float64) {
	// termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if math.Mod(float64(e), 500) == 0 {
		fmt.Printf("Episode %d of %d (epsilon = %v; delta = %v)\n", e, numEpisodes, epsilon, delta)
	}
	// termbox.Flush()
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
func RunEpisode(m *maze.Maze, p *ml.Policy, clientID string, fromCell *pb.MazeLocation,
	toCell *maze.Cell, maxSteps int64) (e episode, err error) {
	if fromCell == nil {
		numStates := int(m.Config().Columns * m.Config().Rows)
		// pick a random state to start at (fromCell), toCell is always the same
		s := int(utils.Random(0, numStates))
		fromCell, err = utils.LocationFromState(m.Config().Rows, m.Config().Columns, int64(s))
		if err != nil {
			return e, err
		}
	}
	state, err := utils.StateFromLocation(m.Config().Rows, m.Config().Columns, fromCell)
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
	steps := int64(0)

	// log.Printf("Solving...")
	// log.Printf("policy:\n%v", p)
	for !solved {
		steps++
		// get the action, according to policy, for this state
		action := p.BestWeightedActionsForState(state)
		//log.Printf("state: %v; action: %v", state, ml.ActionToText[action])

		// get the next state
		nextState, reward, valid, err := ml.NextState(m, toCell.Location(), state, action)
		if err != nil {
			return e, err
		}
		//log.Printf("nextState: %v, reward: %v, valid: %v, err: %v", nextState, reward, valid, err)

		sr := stateReturn{
			state:  state,
			action: action,
			reward: reward,
		}

		e.sr = append(e.sr, sr)

		if utils.LocsSame(c.CurrentLocation().Location(), toCell.Location()) {
			// log.Printf("+++ solved in %v steps!", steps)
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

		if steps > maxSteps {
			// log.Printf("--- not solved in %v steps!", steps)
			break
		}
	}

	return e, err
}

// cumulativeRewardSince returns the cumulative reward since step s
func cumulativeRewardSince(e episode, s int) float64 {
	var r float64
	for x := s; x < len(e.sr); x++ {
		r = r + e.sr[s].reward
	}
	return r
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

// stateActionsInEpisode returns a list of all state,action pairs in an episode
func stateActionsInEpisode(e episode) []StateAction {
	stateActionMap := make(map[StateAction]bool)
	for _, sr := range e.sr {
		sa := StateAction{sr.state, sr.action}
		stateActionMap[sa] = true
	}

	stateActions := make([]StateAction, 0, len(stateActionMap))
	for k := range stateActionMap {
		stateActions = append(stateActions, k)
	}
	sort.Sort(ByStateAction(stateActions))

	return stateActions
}

// firstStateActionInEpisodeIdx returns the first index of the state,action in the episode (first time we reached the state)
// minIndex controls the min location of the return value
func firstStateActionInEpisodeIdx(e episode, state, action, minIndex int) (int, error) {
	for idx, sr := range e.sr {
		if sr.state == state && sr.action == action && idx >= minIndex {
			return idx, nil
		}
	}
	return -1, fmt.Errorf("unable to find state/action: %v/%v in episode: %v", state, action, e)
}

// sumRewardsSinceIdx returns the sum of all rewards since state at index idx
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

func productWeightProbability(min, max int, state, action int, p *ml.Policy) float64 {
	w := 1.0
	for x := min; x <= max; x++ {
		v := p.GetStateActionValue(state, action)
		w = w * (1 / v)
	}
	return w
}

// Evaluate returns the value function for the given policy
func Evaluate(p *ml.Policy, m *maze.Maze, clientID string, numEpisodes int64, df float64,
	toCell *maze.Cell, maxSteps int64) (*ml.ValueFunction, error) {

	// termbox.Init()
	numStates := int(m.Config().Columns * m.Config().Rows)

	// map of state -> sum of all returns in that state
	returnsSum := make(map[int]float64)
	// map of state -> count of all returns
	returnsCount := make(map[int]float64)

	vf := ml.NewValueFunction(numStates)

	// run through the policy this many times
	// each run is a walk through the maze until the end (or limit?)
	for e := int64(0); e < numEpisodes; e++ {
		printProgress(e, numEpisodes, -1, -1)

		// generate an episode (wonder through the maze following policy)
		// An episode is an array of (state, action, reward) tuples
		episode, err := RunEpisode(m, p, clientID, nil, toCell, maxSteps)
		if err != nil {
			return nil, err
		}

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
