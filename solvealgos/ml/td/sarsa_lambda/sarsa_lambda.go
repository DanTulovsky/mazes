package sarsa_lambda

import (
	"fmt"
	"log"
	"time"

	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/ml"
	"github.com/DanTulovsky/mazes/solvealgos"
	"github.com/DanTulovsky/mazes/utils"

	"math"

	pb "github.com/DanTulovsky/mazes/proto"
)

type MLTDSarsaLambda struct {
	solvealgos.Common
}

func printSarsaProgress(e, numEpisodes int64, epsilon float64) {
	// termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if math.Mod(float64(e), 10) == 0 {
		fmt.Printf("Episode %d of %d (epsilon = %v)\n", e, numEpisodes, epsilon)
	}
	// termbox.Flush()
}

// runEpisode runs through the maze, updating the svf and policy on each step
func (a *MLTDSarsaLambda) runSarsaEpisode(mazeID string, rows, columns int64, clientID string, svf, evf *ml.StateActionValueFunction,
	p *ml.Policy, fromCell, toCell *pb.MazeLocation, maxSteps int64, gamma, lambda, alpha, epsilon float64) (err error) {

	numStates := int(columns * rows)

	state, err := utils.StateFromLocation(rows, columns, fromCell)
	if err != nil {
		return err
	}

	solved := false
	steps := int64(0)

	// get the action, according to policy, for this state
	action := p.BestWeightedActionsForState(state)
	if err != nil {
		return err
	}

	for !solved {
		steps++

		// only actually move if we picked a valid direction, otherwise we stay in the same place
		// log.Printf("moving: %v", ml.ActionToText[action])
		reply, err := a.Move(mazeID, clientID, ml.ActionToText[action])
		//if err != nil {
		//	return err
		//}

		nextState, err := utils.StateFromLocation(rows, columns, reply.GetCurrentLocation())
		if err != nil {
			return fmt.Errorf("failed to extract state from location [%v]: %v", reply.GetCurrentLocation(), err)
		}
		reward := reply.GetReward()
		solved = reply.GetSolved()
		// log.Printf("reward: %v", reward)

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

	if solved {
		log.Printf("maze solved in %v steps!", steps)
	}

	return err
}

func (a *MLTDSarsaLambda) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
	directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	// params
	numEpisodes := int64(1000000)
	epsilon := 0.99 // chance of picking random action [0-1], used to explore
	epsilonDecay := -0.00000001
	maxSteps := int64(10000)
	alpha := 0.01 // learning rate
	gamma := 0.9  // discount rate of earlier steps
	lambda := 0.2 // trace decay parameter (1 = monte carlo, 0 = TD(0))

	numStates := int(m.Config().Columns * m.Config().Rows)

	// state,action -> value function (Q)
	svf := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// eligibility traces
	evf := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

	// policy
	p := ml.NewEpsilonGreedyPolicy(numStates, ml.DefaultActions, epsilon)

	for e := int64(0); e < numEpisodes; e++ {
		// reset client location
		reply, err := a.ResetClient(mazeID, clientID)
		if err != nil || !reply.GetSuccess() {
			return fmt.Errorf("error resetting client: %v [%v]", err, reply.GetMessage())
		}

		log.Printf("epsilon: %v", epsilon)
		log.Printf("epsilon decay: %v", epsilonDecay)
		log.Printf("learning rate (alpha): %v", alpha)
		log.Printf("discount rate of steps (gamma): %v", gamma)
		log.Printf("trace decay (lambda): %v", lambda)
		log.Println()

		// slowly decrease epsilon, do less exploration over time
		epsilon = utils.Decay(epsilon, float64(e), epsilonDecay)
		if epsilon < 0.01 {
			epsilon = 0.01
		}

		printSarsaProgress(e, numEpisodes, epsilon)

		if err := a.runSarsaEpisode(mazeID, m.Config().Rows, m.Config().Columns, clientID, svf, evf, p, fromCell, toCell,
			maxSteps, gamma, lambda, alpha, epsilon); err != nil {
			return err
		}

		// log.Printf("new client location: %v", reply.GetCurrentLocation())
	}

	a.ShowStats()

	return nil
}
