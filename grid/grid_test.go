package grid

import "testing"

func TestNewGrid(t *testing.T) {
	rows := 10
	columns := 10
	g := NewGrid(rows, columns)

	if g.Size() != rows*columns {
		t.Errorf("Expected size [%v], but have [%v]", rows*columns, g.Size())
	}
}

func BenchmarkNewGrid(b *testing.B) {
	rows := 10
	columns := 10
	for i := 0; i < b.N; i++ {
		NewGrid(rows, columns)
	}

}
