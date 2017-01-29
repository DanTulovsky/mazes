package bintree

import (
	"mazes/grid"
	"testing"
)

func TestBinTreeApply(t *testing.T) {
	config := &grid.Config{}

	g := grid.NewGrid(config)
	Apply(g)
}

func BenchmarkBinTreeApply(b *testing.B) {
	config := &grid.Config{}

	for i := 0; i < b.N; i++ {
		g := grid.NewGrid(config)
		Apply(g)
	}

}
