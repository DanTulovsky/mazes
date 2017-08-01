package mc

import (
	"mazes/maze"
	"mazes/ml"
	"testing"

	"github.com/tevino/abool"
)

func TestOffPolicyControlImportanceSampling(t *testing.T) {
	for _, tt := range policytests {
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
		svf, policy, err := OffPolicyControlImportanceSampling(m, tt.clientID, numEpisodes, tt.theta, tt.df,
			nil, toCell, maxSteps)
		if err != nil {
			t.Fatalf("error evaluating policy: %v", err)
		}
		t.Logf("maze:\n%v\n", encoded)
		t.Logf("state-action value function (%v):\n%v", tt.clientID, svf.Reshape(int(tt.config.Rows*tt.config.Columns), len(ml.DefaultActions)))
		t.Logf("optimal policy (%v):\n%v", tt.clientID, policy)
	}
}
