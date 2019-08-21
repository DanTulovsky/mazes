package td

import (
	"log"
	"math"

	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/ml"

	pb "github.com/DanTulovsky/mazes/proto"
	"github.com/DanTulovsky/mazes/utils"
)

func printTDLambdaEvalProgress(e, numEpisodes int64) {
	if math.Mod(float64(e), 30) == 0 {
		log.Printf("Episode %d of %d\n", e, numEpisodes)
	}
}

// RunEpisode runs through the maze once, following the policy.
func RunEpisode(m *maze.Maze, p *ml.Policy, vf, et *ml.ValueFunction, clientID string, fromCell *pb.MazeLocation,
	toCell *maze.Cell, maxSteps int64, gamma, lambda, alpha float64) (err error) {

	numStates := int(m.Config().Columns * m.Config().Rows)

	// initial starting state
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
	// log.Printf("initial state: %v, initial location: %v", state, c.CurrentLocation().Location())

	solved := false
	steps := int64(0)

	// log.Printf("Solving...")
	// log.Printf("policy:\n%v", p)
	for !solved {
		steps++
		// get the action, according to policy, for this state
		action := p.BestWeightedActionsForState(state)
		// log.Printf("state: %v; action: %v", state, ml.ActionToText[action])

		// get the next state
		nextState, reward, valid, err := ml.NextState(m, toCell.Location(), state, action)
		if err != nil {
			return err
		}
		//log.Printf("nextState: %v, reward: %v, valid: %v, err: %v", nextState, reward, valid, err)

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

		vnext, err := vf.Get(nextState)
		if err != nil {
			return err
		}
		vstate, err := vf.Get(state)
		if err != nil {
			return err
		}

		// sigma (TD error)
		// sigma = r + gamma*V(s')-V(s)
		sigma := reward + gamma*vnext - vstate

		// update eligibility trace for 'state'
		// e(s) = e(s) + 1
		etState, err := et.Get(state)
		if err != nil {
			return err
		}
		et.Set(state, etState+1)

		// for all states:
		// V(s) = V(s) + alpha*sigma*e(s)
		// e(s) = gamma*lamba*e(s)
		for s := 0; s < numStates; s++ {
			vs, err := vf.Get(s)
			if err != nil {
				return err
			}
			es, err := et.Get(s)
			if err != nil {
				return err
			}
			vf.Set(s, vs+alpha*sigma*es)
			et.Set(s, gamma*lambda*es)
		}

		state = nextState
		// log.Printf("current location: %v, current state: %v", c.CurrentLocation().Location(), state)

		if steps > maxSteps {
			log.Printf("--- not solved in %v steps!", steps)
			break
		}
		// log.Printf("\n%v", et.Reshape(int(m.Config().Rows), int(m.Config().Columns)))
	}

	return err
}

// Evaluate returns the value function for the given policy
func TDLambdaEvaluate(p *ml.Policy, m *maze.Maze, clientID string, numEpisodes int64, gamma, lambda, alpha float64,
	fromCell *pb.MazeLocation, toCell *maze.Cell, maxSteps int64) (*ml.ValueFunction, error) {

	if fromCell == nil {
		var err error
		numStates := int(m.Config().Columns * m.Config().Rows)
		// pick a random state to start at (fromCell), toCell is always the same
		s := int(utils.Random(0, numStates))
		fromCell, err = utils.LocationFromState(m.Config().Rows, m.Config().Columns, int64(s))
		if err != nil {
			return nil, err
		}
	}

	numStates := int(m.Config().Columns * m.Config().Rows)

	// value function we are solving for
	vf := ml.NewValueFunction(numStates)

	// eligibility traces for each state
	et := ml.NewValueFunction(numStates)

	// run through the policy this many times
	for e := int64(0); e < numEpisodes; e++ {
		printTDLambdaEvalProgress(e, numEpisodes)

		// generate an episode (wonder through the maze following policy)
		// An episode is an array of (state, action, reward) tuples
		err := RunEpisode(m, p, vf, et, clientID, fromCell, toCell, maxSteps, gamma, lambda, alpha)
		if err != nil {
			return nil, err
		}

	}

	return vf, nil
}
