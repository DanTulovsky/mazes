package grid

// Stack is a stack of *Cell objects
type Stack struct {
	cells []*Cell
}

func NewStack() *Stack {
	return &Stack{
		cells: make([]*Cell, 0),
	}
}

func (s *Stack) Push(c *Cell) {
	s.cells = append(s.cells, c)

}

func (s *Stack) Pop() (cell *Cell) {
	if len(s.cells) == 0 {
		return nil
	}
	cell, s.cells = s.cells[len(s.cells)-1], s.cells[:len(s.cells)-1]
	return cell
}

func (s *Stack) Size() int {
	return len(s.cells)
}

// Top returns the topmost cell (without popping it off)
func (s *Stack) Top() *Cell {
	if len(s.cells) == 0 {
		return nil
	}
	return s.cells[len(s.cells)-1]
}

// List returns the list of cells in the stack
func (s *Stack) List() []*Cell {
	return s.cells
}
