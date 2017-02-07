package tree

import "fmt"

// spanningTree implements a spanning tree of string objects
// to be used for a tree of cells, where the value of each cell is the coordinate: "x,y"
type Tree struct {
	root *node
}

// NewTree returns a new tree with node n as the root
func NewTree(n *node) (*Tree, error) {
	return &Tree{
		root: n,
	}, nil
}

func (t *Tree) String() string {

	var output string
	var walk func(n *node) *node

	walk = func(n *node) *node {
		output = fmt.Sprintf("%v\nnode: %v; children: %v", output, n, n.Children())
		for _, c := range n.Children() {
			if walk(c) != nil {
				return c
			}
		}
		return nil
	}

	walk(t.Root())

	return output
}

// AddNode adds node n to the tree at parent p
func (t *Tree) AddNode(n *node, p *node) error {
	if p == nil {
		return fmt.Errorf("new tree node [%v] cannot have a nil parent [%v]", n, p)
	}
	if n == nil {
		return fmt.Errorf("new tree node cannot be nil: %v", n)
	}

	if t.Node(n.value) != nil {
		return fmt.Errorf("node with value [%v] already exists", n.value)
	}

	n.parent = p
	p.children[n] = true

	return nil
}

// NodeCount returns the number of nodes in the tree
func (t *Tree) NodeCount() int {

	var nodes int
	var walk func(n *node) *node

	walk = func(n *node) *node {
		nodes++
		for _, c := range n.Children() {
			if walk(c) != nil {
				return c
			}
		}
		return nil
	}

	walk(t.Root())

	return nodes
}

// findNode returns the node with value v, nil otherwise
func findNode(n *node, v string) *node {
	if n.value == v {
		return n
	}

	for _, n := range n.Children() {
		if r := findNode(n, v); r != nil {
			return r
		}
	}

	return nil
}

// Node returns the node with the provided value, nil if not found
func (t *Tree) Node(v string) *node {
	return findNode(t.Root(), v)

}

// Root returns the root node of the tree
func (t *Tree) Root() *node {
	return t.root
}

type node struct {
	value    string
	parent   *node
	children map[*node]bool
}

// NewNode creates a new node
func NewNode(v string) *node {
	return &node{
		value:    v,
		children: make(map[*node]bool),
	}
}

func (n *node) String() string {
	return fmt.Sprintf("%v", n.value)
}

func (n *node) isLeaf() bool {
	return len(n.children) == 0
}

func keys(m map[*node]bool) []*node {
	var k []*node
	for key := range m {
		k = append(k, key)
	}
	return k
}

// Children returns a list of children of the current node
func (n *node) Children() []*node {
	return keys(n.children)
}

// Parent returns the parent of the current node
func (n *node) Parent() *node {
	return n.parent
}
