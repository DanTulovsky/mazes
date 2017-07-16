package dp

import (
	"mazes/algos"
	"mazes/maze"
	"testing"

	"github.com/tevino/abool"
)

func TestPolicyImprovement(t *testing.T) {
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

		policy, vf, err := PolicyImprovement(m, tt.clientID, tt.df, tt.theta, tt.actions)
		if err != nil {
			t.Fatalf("%v", err)
		}

		encoded, err := m.Encode()
		t.Logf("maze:\n%v\npolicy:\n%v\nvalue function (%v):\n%v", encoded, policy, tt.clientID, vf.Reshape(int(tt.config.Rows), int(tt.config.Columns)))

	}
}
