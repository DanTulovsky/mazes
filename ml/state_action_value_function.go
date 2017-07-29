package ml

import (
	"fmt"

	"github.com/gonum/matrix/mat64"
)

// StateActionValueFunction is a matrix mapping state, action (row, column) to value
type StateActionValueFunction struct {
	m *mat64.Dense
}

func NewStateActionValueFunction(r, c int) *StateActionValueFunction {
	m := mat64.NewDense(r, c, nil)

	return &StateActionValueFunction{
		m: m,
	}
}

func (svf *StateActionValueFunction) String() string {
	r, c := svf.m.Dims()
	excerpt := 0
	if r > 10 || c > 10 {
		excerpt = 5
	}
	return fmt.Sprintf("%v\n\n", mat64.Formatted(svf.Matrix(), mat64.Prefix(""), mat64.Excerpt(excerpt)))
}

// Vector returns the underlying vector
func (svf *StateActionValueFunction) Matrix() *mat64.Dense {
	return svf.m
}

func (vf *StateActionValueFunction) Reshape(rows, columns int) string {
	reshaped := reshape(vf.m, rows, columns)
	return fmt.Sprintf("%v\n\n", mat64.Formatted(reshaped, mat64.Prefix(""), mat64.Excerpt(0)))

}

// Set sets the value at location l to v.
func (svf *StateActionValueFunction) Set(r, c int, v float64) error {
	rows, columns := svf.m.Dims()
	if r >= rows || c >= columns {
		return fmt.Errorf("(%v, %v) invalid, matrix dims are (%v, %v)", r, c, rows, columns)
	}

	svf.m.Set(r, c, v)
	return nil
}

// Get retrieves the value at index l
func (svf *StateActionValueFunction) Get(r, c int) (float64, error) {
	rows, columns := svf.m.Dims()
	if r >= rows || c >= columns {
		return 0, fmt.Errorf("(%v, %v) invalid, matrix dims are (%v, %v)", r, c, rows, columns)
	}

	return svf.m.At(r, c), nil
}

func (svf *StateActionValueFunction) ValuesForState(s int) *mat64.Vector {
	return svf.m.RowView(s)
}
