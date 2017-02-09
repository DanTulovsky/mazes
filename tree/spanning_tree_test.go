package tree

import (
	"sort"
	"testing"
)

var treetests = []struct {
	wantErr bool
	nodes   map[string][]string // parent nodes are created in order, and must sort that way
}{
	{
		nodes: map[string][]string{
			// "parent": {"child1", "child2"}
			"0": {"1", "2"},
			"1": {"3", "4"},
			"2": {"5", "6"},
			"6": {"7", "8"},
			"8": {"10"},
		},
		wantErr: false,
	}, {
		nodes: map[string][]string{
			// "parent": {"child1", "child2"}
			"0": {"1", "2"},
			"1": {"3", "4"},
			"2": {"5", "6"},
			"6": {"7", "8"},
			"8": {"10", "0"}, // invalid link
		},
		wantErr: true,
	}, {
		nodes: map[string][]string{
			// "parent": {"child1", "child2"}
			"0": {"1", "2"},
			"1": {"3", "4"},
			"2": {"5", "6"},
			"6": {"7", "8"},
			"8": {"9", "96"},
			"9": {"91", "92", "93", "94", "95"},
		},
		wantErr: false,
	},
}

func makeTree(nodes map[string][]string) (*Tree, error) {
	root := NewNode("0")
	tree, err := NewTree(root)
	if err != nil {
		return nil, err
	}

	keys := []string{}
	for k := range nodes {
		keys = append(keys, k)
		sort.Strings(keys)
	}

	for _, k := range keys {
		p := k
		children := nodes[k]
		parent := tree.Node(p)

		for _, c := range children {
			n := tree.Node(c)

			if n == nil {
				n = NewNode(c)
			}
			if err := tree.AddNode(n, parent); err != nil {
				return nil, err
			}
		}
	}

	return tree, nil
}

func TestTree(t *testing.T) {

	for _, tt := range treetests {
		trie, err := makeTree(tt.nodes)
		if err != nil && !tt.wantErr {
			t.Errorf("failed to create tree: %v\n%v\n", err, trie)
		}

	}
}
