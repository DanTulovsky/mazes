package dp

import (
	"mazes/algos"
	"mazes/maze"
	pb "mazes/proto"
	"testing"

	"github.com/tevino/abool"
)

var policytests = []struct {
	actions      []int
	config       *pb.MazeConfig
	clientConfig *pb.ClientConfig
	df           float64
	theta        float64 // when delta is smaller, eval stops
	clientID     string
}{
	{
		config: &pb.MazeConfig{
			Columns:    3,
			Rows:       2,
			CreateAlgo: "empty",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",               // doesn't matter
			ToCell:    "2,1",
		},
		clientID: "client-empty-dp-eval-only",
		df:       0.99,
		theta:    0.00001,
		actions:  []int{North, South, East, West},
	},
	// {
	//	config: &pb.MazeConfig{
	//		Columns:    3,
	//		Rows:       2,
	//		CreateAlgo: "full", // no passages, with df=1 does not converge
	//	},
	//	clientConfig: &pb.ClientConfig{
	//		SolveAlgo: "ml_dp_policy_eval", // no op yet
	//		FromCell:  "0,0",               // doesn't matter
	//		ToCell:    "2,1",
	//	},
	//	clientID: "client-full-dp-eval-only",
	//	df:       0.5,
	//	theta:    0.00001,
	//	actions:  []int{North, South, East, West},
	//}, {
	//	config: &pb.MazeConfig{
	//		Columns:    5,
	//		Rows:       4,
	//		CreateAlgo: "ellers",
	//	},
	//	clientConfig: &pb.ClientConfig{
	//		SolveAlgo: "ml_dp_policy_eval", // no op yet
	//		FromCell:  "0,0",
	//		ToCell:    "2,1",
	//	},
	//	clientID: "client-ellers-dp-eval-only",
	//	df:       0.9,
	//	theta:    0.00001,
	//	actions:  []int{North, South, East, West},
	//}, {
	//	config: &pb.MazeConfig{
	//		Columns:    6,
	//		Rows:       3,
	//		CreateAlgo: "bintree",
	//	},
	//	clientConfig: &pb.ClientConfig{
	//		SolveAlgo: "ml_dp_policy_eval", // no op yet
	//		FromCell:  "0,0",
	//		ToCell:    "2,1",
	//	},
	//	clientID: "client-bintree-dp-eval-only",
	//	df:       0.9,
	//	theta:    0.00001,
	//	actions:  []int{North, South, East, West},
	//}, {
	//	config: &pb.MazeConfig{
	//		Columns:    8,
	//		Rows:       9,
	//		CreateAlgo: "prim",
	//	},
	//	clientConfig: &pb.ClientConfig{
	//		SolveAlgo: "ml_dp_policy_eval", // no op yet
	//		FromCell:  "0,0",
	//		ToCell:    "7,8",
	//	},
	//	clientID: "client-prim-dp-eval-only",
	//	df:       0.9,
	//	theta:    0.00001,
	//	actions:  []int{North, South, East, West},
	//},
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

func TestPolicy_Eval(t *testing.T) {
	for _, tt := range policytests {
		// create empty maze
		m, err := maze.NewMaze(tt.config, nil)
		if err != nil {
			t.Fatalf("error creating maze: %v", err)
		}

		// apply any algorithm to it
		algo := algos.Algorithms[tt.config.CreateAlgo]
		generating := abool.New()
		generating.Set()
		if err := algo.Apply(m, 0, generating); err != nil {
			generating.UnSet()
			t.Fatalf("error applying algorithm: %v", err)
		}
		// required to get the toCell
		_, _, err = m.AddClient(tt.clientID, tt.clientConfig)
		if err != nil {
			t.Fatalf("failed to add client: %v", err)
		}

		p := NewRandomPolicy(int(tt.config.Rows*tt.config.Columns), tt.actions)
		vf, err := p.Eval(m, tt.clientID, tt.df, tt.theta)
		if err != nil {
			t.Fatalf("error evaluating policy: %v", err)
		}
		encoded, err := m.Encode()
		t.Logf("maze:\n%v\nvalue function (%v):\n%v", encoded, tt.clientID, vf.Reshape(int(tt.config.Rows), int(tt.config.Columns)))

	}
}

func TestValueFunction_Set(t *testing.T) {
	for _, tt := range policytests {
		vf := NewValueFunction(int(tt.config.Rows * tt.config.Columns))
		vf.Set(1, 0.33)
		t.Logf("value function:\n%v", vf.Reshape(int(tt.config.Rows), int(tt.config.Columns)))

	}
}
