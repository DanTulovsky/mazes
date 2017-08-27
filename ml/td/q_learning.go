package td

import (
	"log"

	"mazes/maze"
	"mazes/ml"

	"fmt"
	"math"

	pb "mazes/proto"
	"mazes/utils"
)

func printQProgress(e, numEpisodes int64, epsilon float64) {
	// termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if math.Mod(float64(e), 100) == 0 {
		fmt.Printf("Episode %d of %d (epsilon = %v)\n", e, numEpisodes, epsilon)
	}
	// termbox.Flush()
}

// runQEpisode runs through the maze, updating the svf and policy on each step
func runQEpisode(m *maze.Maze, clientID string, svf *ml.StateActionValueFunction,
	p *ml.Policy, fromCell *pb.MazeLocation, toCell *maze.Cell, maxSteps int64, df, alpha, epsilon float64) (err error) {
	if fromCell == nil {
		numStates := int(m.Config().Columns * m.Config().Rows)
		// pick a random state to start at (fromCell), toCell is always the same
		s := int(utils.Random(0, numStates))
		fromCell, err = utils.LocationFromState(m.Config().Rows, m.Config().Columns, int64(s))
		if err != nil {
			return err
		}
	}
	state, err := utils.StateFromLocation(m.Config().Rows, m.Config().Columns, fromCell)
	if err != nil {
		return err
	}

	c, err := m.Client(clientID)
	cell, err := m.CellFromLocation(fromCell)
	if err != nil {
		return err
	}

	c.SetCurrentLocation(cell)

	solved := false
	steps := int64(0)

	for !solved {
		steps++

		action := p.BestWeightedActionsForState(m, state)

		// get the next state
		nextState, reward, valid, err := ml.NextState(m, toCell.Location(), state, action)
		if err != nil {
			return err
		}

		if utils.LocsSame(c.CurrentLocation().Location(), toCell.Location()) {
			// log.Printf("+++ solved in %v steps!", steps)
			solved = true
		}

		if valid && action != ml.None && !solved {
			// only actually move if we picked a valid direction, otherwise we stay in the same place
			// log.Printf("moving: %v", ml.ActionToText[action])
			c, err = m.MoveClient(clientID, ml.ActionToText[action])
			if err != nil {
				return err
			}
		}

		// TD update
		// best_next_action = np.argmax(Q[next_state])
		// td_target = reward + discount_factor * Q[next_state][best_next_action]
		// td_delta = td_target - Q[state][action]
		// Q[state][action] += alpha * td_delta
		bestNextAction := ml.MaxInVectorIndex(svf.ValuesForState(nextState))

		q, err := svf.Get(state, action)
		if err != nil {
			return err
		}

		nextQ, err := svf.Get(nextState, bestNextAction)
		if err != nil {
			return err
		}

		TDTarget := reward + df*nextQ
		TDDelta := TDTarget - q
		svf.Set(state, action, q+alpha*TDDelta)

		// update policy (but change this so policy is retrieved from value function directly)
		actionValues := svf.ValuesForState(state)
		bestAction := ml.MaxInVectorIndex(actionValues)

		var newValue float64

		for a := 0; a < actionValues.Len(); a++ {
			if a == bestAction {
				// log.Printf("found best action: %v", ml.ActionToText[bestAction])
				newValue = 1 - epsilon + epsilon/float64(actionValues.Len())
			} else {
				newValue = epsilon / float64(actionValues.Len())
			}
			p.SetStateAction(state, a, newValue)
		}

		// next move
		state = nextState

		if steps > maxSteps {
			log.Printf("--- not solved in %v steps!", steps)
			break
		}

	}

	return err
}

// ControlEpsilonGreedy returns the optimal state-value function and policy
func QLearning(m *maze.Maze, clientID string, numEpisodes int64, alpha float64, df float64,
	fromCell *pb.MazeLocation, toCell *maze.Cell, maxSteps int64, epsilon float64, epsilonDecay float64) (*ml.StateActionValueFunction, *ml.Policy, error) {

	numStates := int(m.Config().Columns * m.Config().Rows)

	// state,action -> value function (Q)
	svf := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// policy
	p := ml.NewEpsilonGreedyPolicy(numStates, ml.DefaultActions, epsilon)

	for e := int64(0); e < numEpisodes; e++ {
		if err := m.ResetClient(clientID); err != nil {
			return nil, nil, err
		}

		// slowly decrease epsilon, do less exploration over time
		epsilon = utils.Decay(epsilon, float64(e), epsilonDecay)
		// epsilon = epsilon - epsilon/float64(numEpisodes)*(float64(e))
		if epsilon < 0.01 {
			epsilon = 0.01
		}

		printQProgress(e, numEpisodes, epsilon)

		if err := runQEpisode(m, clientID, svf, p, fromCell, toCell, maxSteps, df, alpha, epsilon); err != nil {
			return nil, nil, err
		}

	}
	return svf, p, nil
}
