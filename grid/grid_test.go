package grid

import (
	"mazes/colors"
	"testing"
)

func TestNewGrid(t *testing.T) {
	r := 10
	c := 10
	w := 1
	bgColor := colors.GetColor("white")
	borderColor := colors.GetColor("red")
	wallColor := colors.GetColor("black")
	g := NewGrid(r, c, w, bgColor, borderColor, wallColor)

	if g.Size() != r*c {
		t.Errorf("Expected size [%v], but have [%v]", r*c, g.Size())
	}
}

func BenchmarkNewGrid(b *testing.B) {
	r := 10
	c := 10
	w := 1
	bgColor := colors.GetColor("white")
	borderColor := colors.GetColor("red")
	wallColor := colors.GetColor("black")

	for i := 0; i < b.N; i++ {
		NewGrid(r, c, w, bgColor, borderColor, wallColor)
	}

}
