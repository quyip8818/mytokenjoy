package tree_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/tree"
)

type node struct {
	ID       string
	Children []node
}

func TestFlatten(t *testing.T) {
	t.Parallel()
	input := []node{
		{ID: "a", Children: []node{
			{ID: "a1", Children: nil},
			{ID: "a2", Children: []node{
				{ID: "a2i"},
			}},
		}},
		{ID: "b", Children: nil},
	}

	result := tree.Flatten(input,
		func(n node) []node { return n.Children },
		func(n *node) { n.Children = nil },
	)

	wantIDs := []string{"a", "a1", "a2", "a2i", "b"}
	if len(result) != len(wantIDs) {
		t.Fatalf("Flatten returned %d nodes, want %d", len(result), len(wantIDs))
	}
	for i, n := range result {
		if n.ID != wantIDs[i] {
			t.Errorf("result[%d].ID = %q, want %q", i, n.ID, wantIDs[i])
		}
		if len(n.Children) != 0 {
			t.Errorf("result[%d].Children should be cleared", i)
		}
	}
}

func TestFlattenEmpty(t *testing.T) {
	t.Parallel()
	result := tree.Flatten([]node{},
		func(n node) []node { return n.Children },
		func(n *node) { n.Children = nil },
	)
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d", len(result))
	}
}

func TestFlattenNilClearChildren(t *testing.T) {
	t.Parallel()
	input := []node{
		{ID: "root", Children: []node{{ID: "child"}}},
	}

	result := tree.Flatten(input,
		func(n node) []node { return n.Children },
		nil,
	)

	if len(result) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result))
	}
	// With nil clearChildren, the root node keeps its children field
	if len(result[0].Children) != 1 {
		t.Errorf("expected root to keep children with nil clearChildren")
	}
}
