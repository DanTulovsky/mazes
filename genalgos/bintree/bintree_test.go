package bintree

import (
	"mazes/maze"
	"mazes/utils"
	"testing"
)

var applytests = []struct {
	config  *maze.Config
	wantErr bool
}{
	{
		config: &maze.Config{
			Rows:    utils.Random(5, 40),
			Columns: utils.Random(5, 40),
		},
		wantErr: false,
	}, {
		config: &maze.Config{
			Rows:    10,
			Columns: 15,
		},
		wantErr: false,
	},
}

func setup() *Bintree {
	return &Bintree{}
}

func TestApply(t *testing.T) {

	for _, tt := range applytests {
		g, err := maze.NewMaze(tt.config)
		a := setup()

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}
		}

		if g, err = a.Apply(g, 0); err != nil {
			t.Errorf("apply failed: %v", err)
		}

		if err := a.CheckGrid(g); err != nil {
			t.Errorf("grid is not valid: %v", err)
		}
	}
}

func BenchmarkApply(b *testing.B) {
	config := &maze.Config{
		Rows:    3,
		Columns: 3,
	}

	for i := 0; i < b.N; i++ {
		g, err := maze.NewMaze(config)
		if err != nil {
			b.Errorf("invalid config: %v", err)
		}
		a := setup()
		a.Apply(g, 0)
	}

}
