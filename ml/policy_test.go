package ml

import (
	pb "gogs.wetsnow.com/dant/mazes/proto"
	"testing"

	"github.com/gonum/matrix/mat64"
)

var policytests = []struct {
	actions      []int
	config       *pb.MazeConfig
	clientConfig *pb.ClientConfig
	df           float64 // prefer more recent steps when calculating value (1 = prefer all)
	theta        float64 // when delta is smaller, eval stops
	clientID     string
}{
	{
		config: &pb.MazeConfig{
			Columns:    6,
			Rows:       5,
			CreateAlgo: "empty",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",               // doesn't matter
			ToCell:    "1,1",
		},
		clientID: "client-empty-dp-eval-only",
		df:       0.99,
		theta:    0.00001,
		actions:  DefaultActions,
	},
	{
		config: &pb.MazeConfig{
			Columns:    3,
			Rows:       2,
			CreateAlgo: "full", // no passages, with df=1 does not converge
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",               // doesn't matter
			ToCell:    "2,1",
		},
		clientID: "client-full-dp-eval-only",
		df:       0.5,
		theta:    0.00001,
		actions:  DefaultActions,
	}, {
		config: &pb.MazeConfig{
			Columns:    5,
			Rows:       4,
			CreateAlgo: "ellers",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",
			ToCell:    "2,1",
		},
		clientID: "client-ellers-dp-eval-only",
		df:       0.9,
		theta:    0.00001,
		actions:  DefaultActions,
	}, {
		config: &pb.MazeConfig{
			Columns:    6,
			Rows:       3,
			CreateAlgo: "bintree",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",
			ToCell:    "2,1",
		},
		clientID: "client-bintree-dp-eval-only",
		df:       0.9,
		theta:    0.00001,
		actions:  DefaultActions,
	}, {
		config: &pb.MazeConfig{
			Columns:    4,
			Rows:       5,
			CreateAlgo: "prim",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",
			ToCell:    "3,4",
		},
		clientID: "client-prim-dp-eval-only",
		df:       0.9,
		theta:    0.00001,
		actions:  DefaultActions,
	},
}

func TestNewRandomPolicy(t *testing.T) {
	for _, tt := range policytests {
		p := NewRandomPolicy(int(tt.config.Rows*tt.config.Columns), tt.actions)
		t.Logf("policy:\n%v", p)
	}
}

func TestNewValueFunction(t *testing.T) {
	for _, tt := range policytests {
		vf := NewValueFunction(int(tt.config.Rows * tt.config.Columns))
		t.Logf("vf reshaped:\n%v", vf.Reshape(int(tt.config.Rows), int(tt.config.Columns)))
	}
}

func TestValueFunction_Set(t *testing.T) {
	for _, tt := range policytests {
		vf := NewValueFunction(int(tt.config.Rows * tt.config.Columns))
		vf.Set(1, 0.33)
		t.Logf("value function:\n%v", vf.Reshape(int(tt.config.Rows), int(tt.config.Columns)))

	}
}

var maxinvectortests = []struct {
	v        *mat64.Vector
	expected int
}{
	{v: mat64.NewVector(3, []float64{1, 2, 3}), expected: 2},   // location of '3'
	{v: mat64.NewVector(3, []float64{52, 2, 11}), expected: 0}, // location of '52'
}

func TestMaxInVector(t *testing.T) {
	for _, tt := range maxinvectortests {
		best := MaxInVectorIndex(tt.v)
		if best != tt.expected {
			t.Errorf("expected: %v; got: %v; vector:\n%v", tt.expected, best, mat64.Formatted(tt.v, mat64.Prefix(""), mat64.Excerpt(0)))
		}

	}
}
