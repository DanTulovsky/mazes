package bintree

import (
	"mazes/grid"
	"testing"
)

func TestBinTreeApply(t *testing.T) {
	rows := 10
	columns := 10
	g := grid.NewGrid(rows, columns)
	Apply(g)
}

func BenchmarkBinTreeApply(b *testing.B) {
	rows := 10
	columns := 10
	for i := 0; i < b.N; i++ {
		g := grid.NewGrid(rows, columns)
		Apply(g)
	}

}
