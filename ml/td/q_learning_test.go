package td

import (
	"testing"

	"mazes/maze"
	"mazes/ml"

	"github.com/tevino/abool"

	pb "mazes/proto"
)

var qlearningtests = []struct {
	actions      []int
	config       *pb.MazeConfig
	clientConfig *pb.ClientConfig
	df           float64 // prefer more recent steps when calculating value (1 = prefer all)
	theta        float64 // when delta is smaller, eval stops
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
		df:       0.99,
		theta:    0.00001,
		actions:  ml.DefaultActions,
	},
	//{
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
	//	actions:  ml.DefaultActions,
	//},
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
		df:       0.9,
		theta:    0.00001,
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
		df:       0.9,
		theta:    0.00001,
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
		df:       0.9,
		theta:    0.00001,
		actions:  ml.DefaultActions,
	},
}

func TestQLearning(t *testing.T) {
	for _, tt := range qlearningtests {
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
		svf, policy, err := QLearning(m, tt.clientID, numEpisodes, tt.theta, tt.df,
			nil, toCell, maxSteps, epsilon, epsilonDecayFactor)
		if err != nil {
			t.Fatalf("error evaluating policy: %v", err)
		}
		t.Logf("maze:\n%v\n", encoded)
		t.Logf("state-action value function (%v):\n%v", tt.clientID, svf.Reshape(int(tt.config.Rows*tt.config.Columns), len(ml.DefaultActions)))
		t.Logf("optimal policy (%v):\n%v", tt.clientID, policy)
	}
}
