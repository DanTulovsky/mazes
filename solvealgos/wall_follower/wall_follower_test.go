package wall_follower

import (
	"fmt"
	"log"
	"mazes/genalgos/aldous_broder"
	"mazes/genalgos/recursive_backtracker"
	"mazes/maze"
	"mazes/utils"
	"testing"
)

var applytests = []struct {
	config      *maze.Config
	orphanCells []*maze.Location
	wantErr     bool
}{
	{
		config: &maze.Config{
			Rows:    utils.Random(5, 10),
			Columns: utils.Random(5, 20),
		},
		orphanCells: []*maze.Location{
			{0, 0},
			{4, 4},
		},
		wantErr: false,
	}, {
		config: &maze.Config{
			Rows:    utils.Random(1, 40),
			Columns: utils.Random(1, 20),
		},
		wantErr: false,
	},
}

func TestSolveAldousBroder(t *testing.T) {
	for _, tt := range applytests {
		g, err := maze.NewGrid(tt.config)
		gen, solv := &aldous_broder.AldousBroder{}, &WallFollower{}

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}
		}

		if g, err = gen.Apply(g, 0); err != nil {
			t.Errorf("apply failed: %v", err)
		}

		if err := gen.CheckGrid(g); err != nil {
			fmt.Printf("%v\n", g)
			t.Fatalf("grid is not valid: %v", err)
		}

		g.ResetVisited()
		fromCell := g.RandomCell()
		toCell := g.RandomCell()
		if g, err = solv.Solve(g, fromCell, toCell, 0); err != nil {
			log.Printf("\n%v\n", g)
			t.Fatalf("failed to solve: %v", err)
		}
	}
}

func TestSolveRecursiveBacktracker(t *testing.T) {
	for _, tt := range applytests {
		g, err := maze.NewGrid(tt.config)
		gen, solv := &recursive_backtracker.RecursiveBacktracker{}, &WallFollower{}

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}
		}

		// orphan cells
		for _, l := range tt.orphanCells {
			cell, err := g.Cell(l.X, l.Y)
			if err != nil {
				t.Fatalf(err.Error())
			}
			cell.Orphan()
		}

		if g, err = gen.Apply(g, 0); err != nil {
			t.Errorf("apply failed: %v", err)
		}

		if err := gen.CheckGrid(g); err != nil {
			fmt.Printf("%v\n", g)
			t.Fatalf("grid is not valid: %v", err)
		}

		g.ResetVisited()
		fromCell := g.RandomCell()
		toCell := g.RandomCell()
		if g, err = solv.Solve(g, fromCell, toCell, 0); err != nil {
			log.Printf("\n%v\n", g)
			t.Fatalf("failed to solve: %v", err)
		}

		for _, o := range g.OrphanCells() {
			// make sure orphan cells are not in the solution
			if maze.CellInCellList(o, g.SolvePath.ListCells()) {
				t.Errorf("orpha cell %v is in solvePath [%v]", o, g.SolvePath)
			}
			if maze.CellInCellList(o, g.TravelPath.ListCells()) {
				t.Errorf("orpha cell %v is in travelPath [%v]", o, g.TravelPath)
			}
		}
	}
}
