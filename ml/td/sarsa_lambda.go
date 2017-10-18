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

func printSarsaLambdaProgress(e, numEpisodes int64, epsilon float64) {
	if math.Mod(float64(e), 100) == 0 {
		fmt.Printf("Episode %d of %d (epsilon = %v)\n", e, numEpisodes, epsilon)
	}
}

// runEpisode runs through the maze, updating the svf and policy on each step
func runSarsaLambdaEpisode(m *maze.Maze, clientID string, svf, evf *ml.StateActionValueFunction,
	p *ml.Policy, fromCell *pb.MazeLocation, toCell *maze.Cell, maxSteps int64, gamma, lambda, alpha, epsilon float64) (err error) {
	if fromCell == nil {
		numStates := int(m.Config().Columns * m.Config().Rows)
		// pick a random state to start at (fromCell), toCell is always the same
		s := int(utils.Random(0, numStates))
		fromCell, err = utils.LocationFromState(m.Config().Rows, m.Config().Columns, int64(s))
		if err != nil {
			return err
		}
	}
	numStates := int(m.Config().Columns * m.Config().Rows)

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

	// get the action, according to policy, for this state
	action := p.BestWeightedActionsForState(state)
	if err != nil {
		return err
	}

	for !solved {
		steps++

		// get the next state
		nextState, reward, valid, err := ml.NextState(m, toCell.Location(), state, action)
		if err != nil {
			return err
		}

		if utils.LocsSame(c.CurrentLocation().Location(), toCell.Location()) {
			// log.Printf("+++ solved in %v steps!", steps)
			solved = true
			return nil
		}

		if valid && action != ml.None && !solved {
			// only actually move if we picked a valid direction, otherwise we stay in the same place
			// log.Printf("moving: %v", ml.ActionToText[action])
			c, err = m.MoveClient(clientID, ml.ActionToText[action])
			if err != nil {
				return err
			}
		}

		nextAction := p.BestWeightedActionsForState(nextState)

		// TD Update
		q, err := svf.Get(state, action)
		if err != nil {
			return err
		}

		nextQ, err := svf.Get(nextState, nextAction)
		if err != nil {
			return err
		}

		etState, err := evf.Get(state, action)
		if err != nil {
			return nil
		}

		// sigma = r + gamma*Q(s', a') - Q(s, a)
		sigma := reward + gamma*nextQ - q

		// eligibility trace
		// e(s,a) = e(s, a) + 1
		evf.Set(state, action, etState+1)

		// for all s, a:
		for s := 0; s < numStates; s++ {
			for a := 0; a < len(ml.DefaultActions); a++ {
				// Q(s,a) = Q(s,a) + alpha*sigma*e(s,a)
				qState, err := svf.Get(s, a)
				if err != nil {
					return err
				}
				eState, err := evf.Get(s, a)
				if err != nil {
					return err
				}
				svf.Set(s, a, qState+alpha*sigma*eState)

				// e(s,a) = gamma*lambda*e(s,a)
				evf.Set(s, a, gamma*lambda*eState)
			}
		}

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
		action = nextAction
		state = nextState

		if steps > maxSteps {
			log.Printf("--- not solved in %v steps!", steps)
			break
		}

	}

	return err
}

// ControlEpsilonGreedy returns the optimal state-value function and policy
func SarsaLambda(m *maze.Maze, clientID string, numEpisodes int64, gamma, lambda, alpha, epsilon, epsilonDecay float64,
	fromCell *pb.MazeLocation, toCell *maze.Cell, maxSteps int64) (*ml.StateActionValueFunction, *ml.Policy, error) {

	numStates := int(m.Config().Columns * m.Config().Rows)

	// state,action -> value function (Q)
	svf := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// eligibility traces
	evf := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// policy
	p := ml.NewEpsilonGreedyPolicy(numStates, ml.DefaultActions, epsilon)

	for e := int64(0); e < numEpisodes; e++ {
		if err := m.ResetClient(clientID); err != nil {
			return nil, nil, err
		}

		// slowly decrease epsilon, do less exploration over time
		epsilon = utils.Decay(epsilon, float64(e), epsilonDecay)
		// epsilon = epsilon - epsilon/float64(numEpisodes)*(float64(e))
		if epsilon < 0.01 { // always do *some* exploration
			epsilon = 0.01
		}

		printSarsaLambdaProgress(e, numEpisodes, epsilon)

		if err := runSarsaLambdaEpisode(m, clientID, svf, evf, p, fromCell, toCell, maxSteps, gamma, lambda, alpha, epsilon); err != nil {
			return nil, nil, err
		}

	}
	return svf, p, nil
}
