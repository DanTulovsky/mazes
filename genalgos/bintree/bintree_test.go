package bintree

import (
	"mazes/colors"
	"mazes/grid"
	"testing"
)

func TestBinTreeApply(t *testing.T) {
	r := 10
	c := 10
	w := 1
	bgColor := colors.GetColor("white")
	borderColor := colors.GetColor("red")
	wallColor := colors.GetColor("black")
	g := grid.NewGrid(r, c, w, bgColor, borderColor, wallColor)
	Apply(g)
}

func BenchmarkBinTreeApply(b *testing.B) {
	r := 10
	c := 10
	w := 1
	bgColor := colors.GetColor("white")
	borderColor := colors.GetColor("red")
	wallColor := colors.GetColor("black")
	for i := 0; i < b.N; i++ {
		g := grid.NewGrid(r, c, w, bgColor, borderColor, wallColor)
		Apply(g)
	}

}
