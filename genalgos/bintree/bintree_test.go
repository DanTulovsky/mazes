package bintree

import (
	"github.com/DanTulovsky/mazes/maze"
	pb "github.com/DanTulovsky/mazes/proto"
	"github.com/DanTulovsky/mazes/utils"
	"github.com/tevino/abool"
	"testing"
)

var applytests = []struct {
	config  *pb.MazeConfig
	wantErr bool
}{
	{
		config: &pb.MazeConfig{
			Rows:    utils.Random64(5, 40),
			Columns: utils.Random64(5, 40),
		},
		wantErr: false,
	}, {
		config: &pb.MazeConfig{
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
		g, err := maze.NewMaze(tt.config, nil)
		a := setup()

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}
		}

		if err := a.Apply(g, 0, abool.NewBool(true)); err != nil {
			t.Errorf("apply failed: %v", err)
		}

		if err := a.CheckGrid(g); err != nil {
			t.Errorf("grid is not valid: %v", err)
		}
	}
}

func BenchmarkApply(b *testing.B) {
	config := &pb.MazeConfig{
		Rows:    3,
		Columns: 3,
	}

	for i := 0; i < b.N; i++ {
		g, err := maze.NewMaze(config, nil)
		if err != nil {
			b.Errorf("invalid config: %v", err)
		}
		a := setup()
		a.Apply(g, 0, abool.NewBool(true))
	}

}
