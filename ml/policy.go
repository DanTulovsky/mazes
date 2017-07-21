package ml

import (
	"fmt"
	"mazes/maze"

	"math"

	"mazes/utils"

	"github.com/gonum/matrix"
	"github.com/gonum/matrix/mat64"

	pb "mazes/proto"
)

type Policy struct {
	M       *mat64.Dense // the policy matrix
	actions []int
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
	}
}

// NewPolicyFromValuFunction returns a policy based on the provided value function
func NewPolicyFromValuFunction(m *maze.Maze, endCell *pb.MazeLocation, vf *ValueFunction, df float64, numStates int, actions []int) (*Policy, error) {
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

func (p *Policy) ActionsForState(s int) *mat64.Vector {
	return p.M.RowView(s)
}

func (p *Policy) BestRandomActionsForState(s int) int {
	actions := p.M.RowView(s)
	max := math.Inf(-1)
	var bestActions []int

	for x := 0; x < actions.Len(); x++ {
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

func (p *Policy) String() string {
	r, c := p.M.Dims()
	excerpt := 0
	if r > 10 || c > 10 {
		excerpt = 10000
	}
	return fmt.Sprintf("%v\n\n", mat64.Formatted(p.M, mat64.Prefix(""), mat64.Excerpt(excerpt)))
}
