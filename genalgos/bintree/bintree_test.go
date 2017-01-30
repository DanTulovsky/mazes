package bintree

import (
	"mazes/grid"
	"testing"
)

var bintreeapplytests = []struct {
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

func TestBinTreeApply(t *testing.T) {

	for _, tt := range bintreeapplytests {
		g, err := grid.NewGrid(tt.config)

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}
		}

		Apply(g)
	}
}

func BenchmarkBinTreeApply(b *testing.B) {
	config := &grid.Config{
		Rows:    3,
		Columns: 3,
	}

	for i := 0; i < b.N; i++ {
		g, err := grid.NewGrid(config)
		if err != nil {
			b.Errorf("invalid config: %v", err)
		}
		Apply(g)
	}

}
