package td

import (
	"testing"

	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/ml"

	"github.com/tevino/abool"

	pb "github.com/DanTulovsky/mazes/proto"
)

var tdlambdapolicytests = []struct {
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

func TestTDLambdaPolicy_Eval(t *testing.T) {
	for _, tt := range tdlambdapolicytests {
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
		fromCell, toCell, err := m.AddClient(tt.clientID, tt.clientConfig)
		if err != nil {
			t.Fatalf("failed to add client: %v", err)
		}

		t.Logf("running maze size (%v, %v): %v (%v -> %v)", tt.config.Columns, tt.config.Rows, tt.config.CreateAlgo, fromCell, toCell)

		encoded, err := m.Encode()

		p := ml.NewRandomPolicy(int(tt.config.Rows*tt.config.Columns), tt.actions)
		numEpisodes := int64(1000)
		maxSteps := int64(10000) // max steps per run through maze
		vf, err := TDLambdaEvaluate(p, m, tt.clientID, numEpisodes, tt.gamma, tt.lambda, tt.alpha, fromCell.Location(), toCell, maxSteps)
		if err != nil {
			t.Fatalf("error evaluating policy: %v", err)
		}
		t.Logf("maze:\n%v\nvalue function (%v):\n%v", encoded, tt.clientID, vf.Reshape(int(tt.config.Rows), int(tt.config.Columns)))
	}
}
