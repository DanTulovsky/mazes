package ml

import (
	"fmt"

	"mazes/maze"

	"github.com/gonum/matrix"
	"github.com/gonum/matrix/mat64"

	"math"

	pb "mazes/proto"
	"mazes/utils"
)

type Policy struct {
	M       *mat64.Dense // the policy matrix
	actions []int
	t       string  // policy type (probably redo as interface)
	epsilon float64 // The probability to select a random action. float between 0 and 1.
}

func reshape(m mat64.Matrix, rows, columns int) *mat64.Dense {
	mr, mc := m.Dims()
	if mr*mc != rows*columns {
		panic(matrix.ErrShape)
	}
	var r mat64.Dense
	r.Clone(m)
	raw := r.RawMatrix()
	raw.Rows = rows
	raw.Cols = columns
	raw.Stride = columns
	r.SetRawMatrix(raw)
	return &r
}

// NewZeroPolicy returns a policy that gives each action weight of 0
func NewZeroPolicy(numStates int, actions []int) *Policy {
	return &Policy{
		M:       mat64.NewDense(numStates, len(actions), nil),
		actions: actions,
	}
}

// NewRandomPolicy returns a policy that gives each action the same weight
func NewRandomPolicy(numStates int, actions []int) *Policy {

	m := mat64.NewDense(numStates, len(actions), nil)

	setOne := func(i, j int, v float64) float64 {
		return 1.0 / float64(len(actions))
	}
	m.Apply(setOne, m)
	return &Policy{
		M:       m,
		actions: actions,
		t:       "random",
	}
}

func NewEpsilonGreedyPolicy(numStates int, actions []int, epsilon float64) *Policy {
	m := mat64.NewDense(numStates, len(actions), nil)

	setOne := func(i, j int, v float64) float64 {
		return 1.0 / float64(len(actions))
	}
	m.Apply(setOne, m)
	return &Policy{
		M:       m,
		actions: actions,
		t:       "epsilon_greedy",
		epsilon: epsilon,
	}
}

// NewPolicyFromValueFunction returns a policy based on the provided value function
func NewPolicyFromValueFunction(m *maze.Maze, endCell *pb.MazeLocation, vf *ValueFunction, df float64, numStates int, actions []int) (*Policy, error) {
	policy := NewZeroPolicy(numStates, actions)

	for state := 0; state < numStates; state++ {
		// One step lookahead to find the best action for this state
		actionValues, err := OneStepLookAhead(m, endCell, vf, df, state, len(actions))
		if err != nil {
			return nil, err
		}

		bestAction := MaxInVectorIndex(actionValues)
		// Always take the best action
		policy.SetStateAction(state, bestAction, 1.0)
	}
	return policy, nil
}

func (p *Policy) SetState(state int, values []float64) {
	p.M.SetRow(state, values)

}

func (p *Policy) SetStateAction(state, action int, value float64) {
	p.M.Set(state, action, value)
}

// GetStateActionValue returns the value of state/action
func (p *Policy) GetStateActionValue(state, action int) float64 {
	return p.M.At(state, action)
}

func (p *Policy) ActionsForState(s int) *mat64.Vector {
	return p.M.RowView(s)
}

// SetType sets the type of the policy
func (p *Policy) SetType(t string) {
	p.t = t
}

// BestDeterministicActionForState returns the best action based on policy, ties are broken arbitrarily
func (p *Policy) BestDeterministicActionForState(s int) int {
	actions := p.M.RowView(s)
	max := math.Inf(-1)
	var bestActions []int
	var allActions []int

	for x := 0; x < actions.Len(); x++ {
		allActions = append(allActions, x)

		a := actions.At(x, 0)
		if a >= max {
			if a == max {
				bestActions = append(bestActions, x)
			} else {
				bestActions = []int{x}
			}
			max = a
		}
	}

	return bestActions[utils.Random(0, len(bestActions))]
}

// BestValidDeterministicActionForState returns the best action based on policy, ties are broken arbitrarily
// accepts a list of valid actions
func (p *Policy) BestValidDeterministicActionForState(s int, validActions []*pb.Direction) int {
	actions := p.M.RowView(s)
	max := math.Inf(-1)
	var bestActions []int
	var allActions []int

	for x := 0; x < actions.Len(); x++ {
		if !utils.DirectionInList(validActions, ActionToText[x]) {
			// skip actions that are not valid moves
			continue
		}
		allActions = append(allActions, x)

		a := actions.At(x, 0)
		if a >= max {
			if a == max {
				bestActions = append(bestActions, x)
			} else {
				bestActions = []int{x}
			}
			max = a
		}
	}

	return bestActions[utils.Random(0, len(bestActions))]
}

// BestActionsForState returns the best action based on the probabilities in the policy
// state -> [0.5, 0, 0, 0.5, 0] picks the 1st and 3rd action with 50% probability
func (p *Policy) BestWeightedActionsForState(s int) int {
	actions := p.M.RowView(s)

	// in MC, you have to pick all state/action pairs with some probability
	//for a := 0; a < actions.Len(); a++ {
	//	isValid, _ := ActionIsValid(m, s, a)
	//	if !isValid {
	//		actions.SetVec(a, 0)
	//	}
	//}

	return utils.WeightedChoice(actions)
}

func (p *Policy) String() string {
	r, c := p.M.Dims()
	excerpt := 0
	if r > 10 || c > 10 {
		excerpt = 10000
	}
	return fmt.Sprintf("%v\n\n", mat64.Formatted(p.M, mat64.Prefix(""), mat64.Excerpt(excerpt)))
}
