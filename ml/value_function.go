package ml

import (
	"fmt"

	"encoding/csv"
	"github.com/gonum/matrix/mat64"
	"os"
)

// ValueFunction is a vector mapping int state to value
type ValueFunction struct {
	// should be a vector, but required to be interface for T() to work properly
	v *mat64.Vector
}

func NewValueFunction(states int) *ValueFunction {
	v := mat64.NewVector(states, nil)

	return &ValueFunction{
		v: v,
	}
}

func (vf *ValueFunction) String() string {
	r, c := vf.v.Dims()
	excerpt := 0
	if r > 10 || c > 10 {
		excerpt = 5
	}
	return fmt.Sprintf("%v\n\n", mat64.Formatted(vf.v, mat64.Prefix(""), mat64.Excerpt(excerpt)))
}

// Vector returns the underlying vector
func (vf *ValueFunction) Vector() *mat64.Vector {
	return vf.v
}

func (vf *ValueFunction) Reshape(rows, columns int) string {
	reshaped := reshape(vf.v, rows, columns)
	return fmt.Sprintf("%v\n\n", mat64.Formatted(reshaped, mat64.Prefix(""), mat64.Excerpt(0)))

}

// Set sets the value at location l to v.
func (vf *ValueFunction) Set(l int, v float64) error {
	if l > vf.v.Len() || l < 0 {
		return fmt.Errorf("(ValueFunction.set) invalid value for l (%v), must be between: [0,%v)", l, vf.v.Len())
	}
	vf.v.SetVec(l, v)
	return nil
}

// Get retrieves the value at index l
func (vf *ValueFunction) Get(l int) (float64, error) {
	if l > vf.v.Len() || l < 0 {
		return 0, fmt.Errorf("(ValueFunction.get) invalid value for l (%v), must be between: [0,%v)", l, vf.v.Len())
	}

	return vf.v.At(l, 0), nil
}

func (vf *ValueFunction) ToCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// output titles
	// TODO: write this if needed

	return nil
}
