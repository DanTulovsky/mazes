package mc

import (
	"mazes/maze"
	"mazes/ml"
	pb "mazes/proto"
)

func OffPolicyControlImportanceSampling(m *maze.Maze, clientID string, numEpisodes int64, theta float64, df float64,
	fromCell *pb.MazeLocation, toCell *maze.Cell, maxSteps int64) (*ml.StateActionValueFunction, *ml.Policy, error) {

	numStates := int(m.Config().Columns * m.Config().Rows)

	// The final action-value function.
	Q := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// The cumulative denominator of the weighted importance sampling formula (across all episodes)
	C := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// Our greedily policy we want to learn
	targetPolicy := ml.NewRandomPolicy(numStates, ml.DefaultActions)

	// Behavior policy we follow through the maze
	behaviorPolicy := ml.NewRandomPolicy(numStates, ml.DefaultActions)

	for e := int64(0); e < numEpisodes; e++ {

		if err := m.ResetClient(clientID); err != nil {
			return nil, nil, err
		}

		// generate an episode (wonder through the maze following policy)
		// An episode is an array of (state, action, reward) tuples
		// log.Printf("Generating episode %v...", e)
		episode, err := RunEpisode(m, behaviorPolicy, clientID, fromCell, toCell, maxSteps)
		if err != nil {
			return nil, nil, err
		}

		// Sum of discounted returns
		var G float64
		// The importance sampling ratio (the weights of the returns)
		W := 1.0

		// For each step in the episode, backwards
		for s := len(episode.sr) - 1; s >= 0; s-- {
			state, action, reward := episode.sr[s].state, episode.sr[s].action, episode.sr[s].reward
			// log.Printf("step: %v; state: %v; action: %v; reward: %v", s, state, action, reward)

			// Update the total reward since step t
			G = df*G + reward

			// Update weighted importance sampling formula denominator
			currentC, err := C.Get(state, action)
			if err != nil {
				return nil, nil, err
			}
			if err := C.Set(state, action, currentC+W); err != nil {
				return nil, nil, err
			}

			newC, err := C.Get(state, action)
			if err != nil {
				return nil, nil, err
			}

			currentQ, err := Q.Get(state, action)
			if err != nil {
				return nil, nil, err
			}

			// Update the action-value function using the incremental update formula (5.7)
			// Q[state][action] += (W / C[state][action]) * (G - Q[state][action])
			if err := Q.Set(state, action, (W/newC)*(G-currentQ)); err != nil {
				return nil, nil, err
			}

			// Update target policy
			actionValues := Q.ValuesForState(state)
			bestAction := ml.MaxInVectorIndex(actionValues)

			for a := 0; a < actionValues.Len(); a++ {
				newValue := 0.0
				if a == bestAction {
					newValue = 1
				}
				targetPolicy.SetStateAction(state, a, newValue)
			}

			// If the action taken by the behavior policy is not the action
			// taken by the target policy the probability will be 0 and we can break
			if action != targetPolicy.BestDeterministicActionForState(state) {
				// log.Printf("break...")
				break
			}

			W = W * 1 / behaviorPolicy.GetStateActionValue(state, action)
		}

		printProgress(e, numEpisodes, -1, -1)
	}

	return Q, targetPolicy, nil
}
