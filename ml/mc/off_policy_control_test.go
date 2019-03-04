package mc

import (
	"testing"

	"gogs.wetsnow.com/dant/mazes/maze"
	"gogs.wetsnow.com/dant/mazes/ml"

	pb "gogs.wetsnow.com/dant/mazes/proto"

	"github.com/tevino/abool"
)

var offpolicytests = []struct {
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
	//	actions:  ml.DefaultActions,
	//},
	//{
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
	//	actions:  ml.DefaultActions,
	//}, {
	//	config: &pb.MazeConfig{
	//		Columns:    4,
	//		Rows:       5,
	//		CreateAlgo: "prim",
	//	},
	//	clientConfig: &pb.ClientConfig{
	//		SolveAlgo: "ml_dp_policy_eval", // no op yet
	//		FromCell:  "0,0",
	//		ToCell:    "3,4",
	//	},
	//	clientID: "client-prim-dp-eval-only",
	//	df:       0.9,
	//	theta:    0.00001,
	//	actions:  ml.DefaultActions,
	//},
}

func TestOffPolicyControl(t *testing.T) {
	for _, tt := range offpolicytests {
		t.Logf("running maze size (%v, %v): %v (-> (%v)", tt.config.Columns, tt.config.Rows, tt.config.CreateAlgo, tt.clientConfig.ToCell)
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
		numEpisodes := int64(1000) // number of times to run through maze
		maxSteps := int64(1000)    // max steps per run through maze
		svf, policy, err := OffPolicyControl(m, tt.clientID, numEpisodes, tt.theta, tt.df,
			nil, toCell, maxSteps)
		if err != nil {
			t.Fatalf("error evaluating policy: %v\n", err)
		}
		t.Logf("maze:\n%v\n", encoded)
		t.Logf("state-action value function (%v):\n%v", tt.clientID, svf.Reshape(int(tt.config.Rows*tt.config.Columns), len(ml.DefaultActions)))
		t.Logf("optimal policy (%v):\n%v", tt.clientID, policy)
	}
}
