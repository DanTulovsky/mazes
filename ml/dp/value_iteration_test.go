package dp

import (
	"mazes/maze"
	"testing"

	pb "mazes/proto"

	"github.com/tevino/abool"
)

var valueiterationtests = []struct {
	actions      []int
	config       *pb.MazeConfig
	clientConfig *pb.ClientConfig
	df           float64 // prefer more recent steps when calculating value (1 = prefer all)
	theta        float64 // when delta is smaller, eval stops
	clientID     string
}{
	{
		config: &pb.MazeConfig{
			Columns:    8,
			Rows:       6,
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
		df:       0.99,
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
		df:       0.99,
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
		df:       0.99,
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
		df:       0.99,
		theta:    0.00001,
		actions:  DefaultActions,
	},
}

func TestValueIteration(t *testing.T) {
	for _, tt := range valueiterationtests {
		// create empty maze
		m, err := maze.NewMaze(tt.config, nil)
		if err != nil {
			t.Fatalf("error creating maze: %v", err)
		}

		// apply any algorithm to it
		algo := Algorithms[tt.config.CreateAlgo]
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

		policy, vf, err := ValueIteration(m, tt.clientID, tt.df, tt.theta, tt.actions)
		if err != nil {
			t.Fatalf("%v", err)
		}

		encoded, err := m.Encode()
		t.Logf("maze:\n%v\npolicy(df=%v):\n%v\nvalue function (%v):\n%v", encoded, tt.df, policy, tt.clientID, vf.Reshape(int(tt.config.Rows), int(tt.config.Columns)))

	}
}
