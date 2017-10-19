package one_step_sarsa

import (
	"fmt"
	"log"
	"time"

	"mazes/maze"
	"mazes/ml"
	"mazes/solvealgos"
	"mazes/utils"

	"math"

	pb "mazes/proto"
)

type MLTDOneStepSarsa struct {
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
func (a *MLTDOneStepSarsa) runSarsaEpisode(mazeID string, rows, columns int64, clientID string, svf *ml.StateActionValueFunction,
	p *ml.Policy, fromCell, toCell *pb.MazeLocation, maxSteps int64, df, alpha, epsilon float64) (err error) {

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

		// TD update
		// td_target = reward + discount_factor * Q[next_state][next_action]
		// td_delta = td_target - Q[state][action]
		// Q[state][action] += alpha * td_delta
		q, err := svf.Get(state, action)
		if err != nil {
			return err
		}

		nextQ, err := svf.Get(nextState, nextAction)
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
		action = nextAction
		state = nextState

		if steps > maxSteps {
			log.Printf("--- not solved in %v steps!", steps)
			break
		}

	}

	log.Printf("maze solved in %v steps!", steps)

	return err
}

func (a *MLTDOneStepSarsa) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
	directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	// params
	numEpisodes := int64(1000)
	epsilon := 0.99 // chance of picking random action [0-1], used to explore
	epsilonDecay := -0.001
	maxSteps := int64(10000)
	df := 1.0    // discount factor
	alpha := 0.1 // learning rate

	numStates := int(m.Config().Columns * m.Config().Rows)

	// state,action -> value function (Q)
	svf := ml.NewStateActionValueFunction(numStates, len(ml.DefaultActions))

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
		log.Printf("discount factor: %v", df)
		log.Printf("learning rate (alpha): %v", alpha)
		log.Println()

		// slowly decrease epsilon, do less exploration over time
		epsilon = utils.Decay(epsilon, float64(e), epsilonDecay)
		if epsilon < 0.01 {
			epsilon = 0.01
		}

		printSarsaProgress(e, numEpisodes, epsilon)

		if err := a.runSarsaEpisode(mazeID, m.Config().Rows, m.Config().Columns, clientID, svf, p, fromCell, toCell, maxSteps,
			df, alpha, epsilon); err != nil {
			return err
		}

		// log.Printf("new client location: %v", reply.GetCurrentLocation())
	}

	a.ShowStats()

	return nil
}
