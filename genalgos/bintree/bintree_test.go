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
	pathColor := colors.GetColor("yellow")
	wallWidth := 2
	pathWidth := 2

	g := grid.NewGrid(r, c, w, wallWidth, pathWidth, bgColor, borderColor, wallColor, pathColor)
	Apply(g)
}

func BenchmarkBinTreeApply(b *testing.B) {
	r := 10
	c := 10
	w := 1
	bgColor := colors.GetColor("white")
	borderColor := colors.GetColor("red")
	wallColor := colors.GetColor("black")
	pathColor := colors.GetColor("yellow")
	wallWidth := 2
	pathWidth := 2

	for i := 0; i < b.N; i++ {
		g := grid.NewGrid(r, c, w, wallWidth, pathWidth, bgColor, borderColor, wallColor, pathColor)
		Apply(g)
	}

}
