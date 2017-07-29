package mc

import (
	"mazes/maze"
	"mazes/ml"
)

type StateAction struct {
	state  int
	action int
}

// sort StateAction first by state, then by action
type ByStateAction []StateAction

func (sa ByStateAction) Len() int      { return len(sa) }
func (sa ByStateAction) Swap(i, j int) { sa[i], sa[j] = sa[j], sa[i] }
func (sa ByStateAction) Less(i, j int) bool {
	if sa[i].state == sa[j].state {
		return sa[i].action < sa[j].action
	}
	return sa[i].state < sa[j].state
}

// ControlEpsilonGreedy returns the optimal state-value function and policy
func ControlEpsilonGreedy(m *maze.Maze, clientID string, numEpisodes int, df float64, toCell *maze.Cell, maxSteps int, epsilon float64) (*ml.StateActionValueFunction, *ml.Policy, error) {
	// map of state, action -> sum of all returns in that state
	returnsSum := make(map[StateAction]float64)
	// map of state,action -> count of all returns
	returnsCount := make(map[StateAction]float64)

	numStates := int(m.Config().Columns * m.Config().Rows)

	// state,action -> value function (Q)
	svf := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// policy
	p := ml.NewEpsilonGreedyPolicy(numStates, ml.DefaultActions, epsilon)

	for e := 1; e <= numEpisodes; e++ {
		printProgress(e, numEpisodes)

		// generate an episode (wonder through the maze following policy)
		// An episode is an array of (state, action, reward) tuples
		episode, err := RunEpisode(m, p, clientID, toCell, maxSteps)
		if err != nil {
			return nil, nil, err
		}

		// Find all state/actions pairs the we've visited in this episode
		stateActions := stateActionsInEpisode(episode)
		//log.Printf("stateActions: %v", stateActions)

		for _, sa := range stateActions {
			// Find the first occurrence of the state,action in the episode
			stateActionIdx, err := firstStateActionInEpisodeIdx(episode, sa.state, sa.action)
			if err != nil {
				return nil, nil, err
			}

			// Sum up all rewards since the first occurrence
			sum, err := sumRewardsSinceIdx(episode, stateActionIdx, df)
			if err != nil {
				return nil, nil, err
			}

			// Calculate average return for this state over all sampled episodes
			if _, ok := returnsSum[sa]; !ok {
				returnsSum[sa] = 0
			}
			returnsSum[sa] += sum

			if _, ok := returnsCount[sa]; !ok {
				returnsCount[sa] = 0
			}
			returnsCount[sa]++

			svf.Set(sa.state, sa.action, returnsSum[sa]/returnsCount[sa])

			actionValues := svf.ValuesForState(sa.state)
			// log.Printf("actionValues: %v", actionValues)
			// improve policy for state, action
			bestAction := ml.MaxInVectorIndex(actionValues)
			if sa.action == bestAction {
				old := p.GetStateActionValue(sa.state, sa.action)
				p.SetStateAction(sa.state, sa.action, old+1)
			}
		}
	}

	return svf, p, nil
}
