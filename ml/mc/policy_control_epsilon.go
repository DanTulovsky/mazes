package mc

import (
	"log"
	"mazes/maze"
	"mazes/ml"
	pb "mazes/proto"

	"math"

	"github.com/gonum/matrix/mat64"
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
func ControlEpsilonGreedy(m *maze.Maze, clientID string, numEpisodes int, theta float64, df float64, fromCell *pb.MazeLocation, toCell *maze.Cell, maxSteps int, epsilon float64) (*ml.StateActionValueFunction, *ml.Policy, error) {
	// map of state, action -> sum of all returns in that state
	returnsSum := make(map[StateAction]float64)
	// map of state,action -> count of all returns
	returnsCount := make(map[StateAction]float64)

	numStates := int(m.Config().Columns * m.Config().Rows)

	// state,action -> value function (Q)
	svf := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// policy
	p := ml.NewEpsilonGreedyPolicy(numStates, ml.DefaultActions, epsilon)

	// Run this many epsidoes, or until delta < theta
	for e := 0; e < numEpisodes; e++ {
		delta := 0.0

		// generate an episode (wonder through the maze following policy)
		// An episode is an array of (state, action, reward) tuples
		episode, err := RunEpisode(m, p, clientID, fromCell, toCell, maxSteps)
		if err != nil {
			return nil, nil, err
		}

		// Find all state/actions pairs the we've visited in this episode
		stateActions := stateActionsInEpisode(episode)
		//log.Printf("stateActions: %v", stateActions)

		// log.Printf("Processing...")
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
			// log.Printf("processing state-action: %v; all_rewards: %v", sa, sum)

			// Calculate average return for this state/action over all sampled episodes
			if _, ok := returnsSum[sa]; !ok {
				returnsSum[sa] = 0
			}
			returnsSum[sa] += sum

			if _, ok := returnsCount[sa]; !ok {
				returnsCount[sa] = 0
			}
			returnsCount[sa]++

			svf.Set(sa.state, sa.action, returnsSum[sa]/returnsCount[sa])
		}

		for _, s := range statesInEpisode(episode) {
			actionValues := svf.ValuesForState(s)
			//log.Printf("state: %v, actionValues: %v", s, actionValues)
			// improve policy for state, action
			bestAction := ml.MaxInVectorIndex(actionValues)
			bestActionValue := mat64.Max(actionValues)

			var newValue float64

			for a := 0; a < actionValues.Len(); a++ {
				prevActionValue, err := svf.Get(s, a)
				if err != nil {
					return nil, nil, err
				}
				if a == bestAction {
					// log.Printf("found best action: %v", ml.ActionToText[bestAction])
					newValue = 1 - epsilon + epsilon/float64(actionValues.Len())
				} else {
					newValue = epsilon / float64(actionValues.Len())
				}
				p.SetStateAction(s, a, newValue)

				delta = math.Max(delta, math.Abs(bestActionValue-prevActionValue))
			}
		}
		// log.Printf("...done processing")

		// slowly decrease epsilon, do less exploration over time
		epsilon = epsilon - epsilon/float64(numEpisodes)*(float64(e))
		if epsilon < 0.1 {
			epsilon = 0.1
		}

		printProgress(e, numEpisodes, epsilon, delta)

		if delta < theta {
			log.Printf("stopping, change in value function (%v) less than %v", delta, theta)
			break
		}
	}

	return svf, p, nil
}
