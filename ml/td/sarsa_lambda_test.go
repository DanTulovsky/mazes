package td

import (
	"testing"

	"mazes/maze"
	"mazes/ml"

	"github.com/tevino/abool"

	pb "mazes/proto"
)

var sarsalambdatests = []struct {
	actions      []int
	config       *pb.MazeConfig
	clientConfig *pb.ClientConfig
	gamma        float64 // discount rate of earlier steps
	lambda       float64 // trace decay parameter (1 = monte carlo, 0 = TD(0)), used for eligibility traces
	alpha        float64 // step size for learning, smaller leads to slower learning
	clientID     string
}{
	{
		config: &pb.MazeConfig{
			Columns:    3,
			Rows:       3,
			CreateAlgo: "empty",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",               // doesn't matter
			ToCell:    "1,1",
		},
		clientID: "client-empty-dp-eval-only",
		gamma:    0.99,
		lambda:   0,
		alpha:    0.001,
		actions:  ml.DefaultActions,
	},
	{
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
		gamma:    0.99,
		lambda:   0,
		alpha:    0.001,
		actions:  ml.DefaultActions,
	},
	{
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
		gamma:    0.99,
		lambda:   0,
		alpha:    0.001,
		actions:  ml.DefaultActions,
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
		gamma:    0.99,
		lambda:   0,
		alpha:    0.001,
		actions:  ml.DefaultActions,
	},
}

func TestSarsaLambda(t *testing.T) {
	for _, tt := range sarsalambdatests {
		t.Logf("running maze size (%v, %v): %v; (-> (%v))", tt.config.Columns, tt.config.Rows, tt.config.CreateAlgo, tt.clientConfig.ToCell)
		// create empty maze
		m, err := maze.NewMaze(tt.config, nil)
		if err != nil {
			t.Fatalf("error creating maze: %v", err)
		}

		// apply any algorithm to it
		algo := ml.Algorithms[tt.config.CreateAlgo]
		generating := abool.New()
		generating.Set()
		if err := algo.Apply(m, 0, generating); err != nil {
			generating.UnSet()
			t.Fatalf("error applying algorithm: %v", err)
		}
		// required to get the toCell
		_, toCell, err := m.AddClient(tt.clientID, tt.clientConfig)
		if err != nil {
			t.Fatalf("failed to add client: %v", err)
		}

		encoded, err := m.Encode()
		epsilon := 0.1             // chance of picking random action, to explore
		numEpisodes := int64(1000) // number of times to run through maze
		maxSteps := int64(10000)   // max steps per run through maze
		epsilonDecayFactor := -0.001
		svf, policy, err := SarsaLambda(m, tt.clientID, numEpisodes, tt.gamma, tt.lambda, tt.alpha, epsilon, epsilonDecayFactor,
			nil, toCell, maxSteps)
		if err != nil {
			t.Fatalf("error evaluating policy: %v", err)
		}
		t.Logf("maze:\n%v\n", encoded)
		t.Logf("state-action value function (%v):\n%v", tt.clientID, svf.Reshape(int(tt.config.Rows*tt.config.Columns), len(ml.DefaultActions)))
		t.Logf("optimal policy (%v):\n%v", tt.clientID, policy)
	}
}
