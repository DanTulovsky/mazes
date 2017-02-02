package wall_follower

import (
	"mazes/grid"
	"mazes/genalgos"
	"testing"
	"fmt"
	"mazes/genalgos/aldous_broder"
	"log"
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
			Rows:    10,
			Columns: 10,
		},
		wantErr: false,
	}, {
		config: &grid.Config{
			Rows:    10,
			Columns: 15,
		},
		wantErr: false,
	},
}


func TestSolve(t *testing.T) {
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

		fromCell := g.RandomCell()
		toCell := g.RandomCell()
		if g, err = solv.Solve(g, fromCell, toCell); err != nil {
			log.Printf("\n%v\n", g)
			t.Fatalf("failed to solve: %v", err)
		}
	}
}
