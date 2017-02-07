package maze

import "testing"

var config = &Config{
	Rows:    10,
	Columns: 15,
}

var stacktests = []struct {
	cells []*Cell
}{
	{cells: []*Cell{
		NewCell(0, 0, config),
		NewCell(1, 8, config),
		NewCell(5, 7, config)},
	},
}

func TestStack(t *testing.T) {

	for _, tt := range stacktests {
		stack := NewStack()

		if stack.Size() != 0 {
			t.Errorf("empty stack should have length 0, but has %v", len(stack.List()))
		}
		for _, cell := range tt.cells {
			stack.Push(cell)
		}

		if stack.Size() != len(tt.cells) {
			t.Errorf("expected stack size [%v] does not match received [%v]", len(tt.cells), stack.Size())
		}

		c := stack.Pop()
		if stack.Size() != len(tt.cells)-1 {
			t.Errorf("expected stack size [%v] does not match received [%v]", len(tt.cells)-1, stack.Size())
		}
		l := c.Location()
		if l.X != 5 || l.Y != 7 {
			t.Errorf("expected cell at (5, 7), got (%v)", l)
		}

		c = stack.Top()
		if stack.Size() != len(tt.cells)-1 {
			t.Errorf("expected stack size [%v] does not match received [%v]", len(tt.cells)-1, stack.Size())
		}
		l = c.Location()
		if l.X != 1 || l.Y != 8 {
			t.Errorf("expected cell at (1, 8), got (%v)", l)
		}

		if stack.Size() != len(stack.List()) {
			t.Errorf("size [%v] and length of list [%v] stack mismatch", stack.Size(), len(stack.List()))
		}

	}
}
