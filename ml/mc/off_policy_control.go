package mc

import (
	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/ml"
	pb "github.com/DanTulovsky/mazes/proto"
)

func OffPolicyControl(m *maze.Maze, clientID string, numEpisodes int64, theta float64, df float64,
	fromCell *pb.MazeLocation, toCell *maze.Cell, maxSteps int64) (*ml.StateActionValueFunction, *ml.Policy, error) {

	numStates := int(m.Config().Columns * m.Config().Rows)

	// The final action-value function.
	Q := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// The cumulative numerator of the weighted importance sampling formula (across all episodes)
	N := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// The cumulative denominator of the weighted importance sampling formula (across all episodes)
	D := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// Our greedily policy we want to learn
	targetPolicy := ml.NewRandomPolicy(numStates, ml.DefaultActions)

	// Behavior policy we follow through the maze
	behaviorPolicy := ml.NewRandomPolicy(numStates, ml.DefaultActions)

	for e := int64(0); e < numEpisodes; e++ {
		printProgress(e, numEpisodes, 0, 0)

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

		// the latest time (step) at which behaviorPolicy(s, a) != targetPolicy(s, a)
		var tao int
		for s := len(episode.sr) - 1; s >= 0; s-- {
			state, action, _ := episode.sr[s].state, episode.sr[s].action, episode.sr[s].reward
			if action != targetPolicy.BestDeterministicActionForState(state) {
				tao = s
				break
			}
		}
		// log.Printf("  ++ tao: %v (of %v)", tao, len(episode.sr))

		//log.Printf(">> episode <<\n%v\n", episode)
		// for each (s, a) appearing in the episode at time t or later
		for s := tao; s < len(episode.sr); s++ {
			state, action, _ := episode.sr[s].state, episode.sr[s].action, episode.sr[s].reward
			//log.Printf("evaluating: state: %v, action: %v", state, action)

			// t is the time (step) of first occurance of (s,a) such that t > tao
			t, err := firstStateActionInEpisodeIdx(episode, state, action, tao)
			if err != nil {
				return nil, nil, err
			}
			//log.Printf("  first occurance of (%v, %v): %v", state, action, t)

			w := productWeightProbability(t+1, len(episode.sr)-1, state, action, behaviorPolicy)
			if err != nil {
				return nil, nil, err
			}
			//log.Printf("  weighted probability: %v", w)

			cReward := cumulativeRewardSince(episode, t)
			//log.Printf("  cReward: %v", cReward)
			currentN, _ := N.Get(state, action)
			N.Set(state, action, currentN+w*cReward)
			newN, _ := N.Get(state, action)
			//log.Printf("  newN (%v+%v*%v): %v", currentN, w, cReward, newN)

			currentD, _ := D.Get(state, action)
			D.Set(state, action, currentD+w)
			newD, _ := D.Get(state, action)
			//log.Printf("  newD (%v+%v): %v", currentD, w, newD)

			//log.Printf("  new value (%v/%v): %v", newN, newD, newN/newD)
			if err := Q.Set(state, action, newN/newD); err != nil {
				return nil, nil, err
			}
		}

		// update targetPolicy
		//log.Printf("updating targetPolicy...")
		for s := 0; s < numStates; s++ {
			// log.Printf("  state: %v", s)
			actionValues := Q.ValuesForState(s)
			//log.Printf("    actionvalues: %v", mat64.Formatted(actionValues.T(), mat64.Prefix(""), mat64.Squeeze()))

			bestAction := ml.MaxInVectorIndex(actionValues)
			//log.Printf("    bestAction: %v", bestAction)

			for a := 0; a < actionValues.Len(); a++ {
				// log.Printf("a: %v; bestAction: %v", a, bestAction)
				newValue := 0.0
				if a == bestAction {
					newValue = 1
				}
				targetPolicy.SetStateAction(s, a, newValue)
			}
		}
		//log.Printf("targetPolicy:\n%v", targetPolicy.String())

		// Sum of discounted returns
		//var G float64
		//// The importance sampling ratio (the weights of the returns)
		//W := 1.0
		//
		//// For each step in the episode, backwards
		//for s := len(episode.sr) - 1; s >= 0; s-- {
		//	state, action, reward := episode.sr[s].state, episode.sr[s].action, episode.sr[s].reward
		//	// log.Printf("step: %v; state: %v; action: %v; reward: %v", s, state, action, reward)
		//
		//	// If the action taken by the behavior policy is not the action
		//	// taken by the target policy the probability will be 0 and we can break
		//	if action != targetPolicy.BestDeterministicActionForState(state) {
		//		log.Printf("break...")
		//		break
		//	}
		//
		//	// Update the total reward since step t
		//	G = df*G + reward
		//
		//	// Update weighted importance sampling formula denominator
		//	currentC, err := D.Get(state, action)
		//	if err != nil {
		//		return nil, nil, err
		//	}
		//	if err := D.Set(state, action, currentC+W); err != nil {
		//		return nil, nil, err
		//	}
		//
		//	newC, err := D.Get(state, action)
		//	if err != nil {
		//		return nil, nil, err
		//	}
		//
		//	currentQ, err := Q.Get(state, action)
		//	if err != nil {
		//		return nil, nil, err
		//	}
		//
		//	// Update the action-value function using the incremental update formula (5.7)
		//	// Q[state][action] += (W / D[state][action]) * (G - Q[state][action])
		//	if err := Q.Set(state, action, (W/newC)*(G-currentQ)); err != nil {
		//		return nil, nil, err
		//	}
		//
		//	// Update target policy
		//	actionValues := Q.ValuesForState(state)
		//	bestAction := ml.MaxInVectorIndex(actionValues)
		//
		//	for a := 0; a < actionValues.Len(); a++ {
		//		log.Printf("a: %v; bestAction: %v", a, bestAction)
		//		newValue := 0.0
		//		if a == bestAction {
		//			newValue = 1
		//		}
		//		targetPolicy.SetStateAction(state, a, newValue)
		//	}
		//
		//	W = W * 1 / behaviorPolicy.GetStateActionValue(state, action)
		//}
		//
		//printProgress(e, numEpisodes, -1, -1)
	}

	return Q, targetPolicy, nil
}
