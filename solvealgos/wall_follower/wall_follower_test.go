package wall_follower

import (
	"mazes/grid"
	"testing"
	"fmt"
	"mazes/genalgos/aldous_broder"
	"log"
	"mazes/genalgos"
)


func setup() (genalgos.Algorithmer, *WallFollower) {

	return &aldous_broder.AldousBroder{}, &WallFollower{}
}

var applytests = []struct {
	config  *grid.Config
	wantErr bool
}{
	{
		config: &grid.Config{
			Rows:    4,
			Columns: 4,
		},
		wantErr: false,
	}, {
		config: &grid.Config{
			Rows:    5,
			Columns: 5,
		},
		wantErr: false,
	},
}


func TestSolveAldousBroder(t *testing.T) {
	for _, tt := range applytests {
		g, err := grid.NewGrid(tt.config)
		gen, solv := setup()

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}
		}

		if g, err = gen.Apply(g); err != nil {
			t.Errorf("apply failed: %v", err)
		}

		if err := gen.CheckGrid(g); err != nil {
			fmt.Printf("%v\n", g)
			t.Fatalf("grid is not valid: %v", err)
		}

		g.ResetVisited()
		fromCell := g.RandomCell()
		toCell := g.RandomCell()
		if g, err = solv.Solve(g, fromCell, toCell); err != nil {
			log.Printf("\n%v\n", g)
			t.Fatalf("failed to solve: %v", err)
		}
	}
}
