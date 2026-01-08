package engine

import (
	"testing"
)

func TestQuadtree_InsertAndQuery(t *testing.T) {
	boundary := NewAABB(0, 0, 100)
	qt := NewQuadtree(boundary, 4)

	entities := []*Entity{
		{ID: "1", Position: Vec2{X: 10, Y: 10}},
		{ID: "2", Position: Vec2{X: -10, Y: -10}},
		{ID: "3", Position: Vec2{X: 50, Y: 50}},
		{ID: "4", Position: Vec2{X: -50, Y: -50}},
		{ID: "5", Position: Vec2{X: 0, Y: 0}},
	}

	for _, e := range entities {
		qt.Insert(e)
	}

	// Test QueryCircle
	circle := &Circle{CenterX: 0, CenterY: 0, Radius: 20}
	found := qt.QueryCircle(circle, nil)

	if len(found) != 3 {
		t.Errorf("expected 3 entities in circle, found %d", len(found))
	}

	// Double check entity IDs (1, 2, 5 should be in range)
	idMap := make(map[string]bool)
	for _, e := range found {
		idMap[e.ID] = true
	}
	if !idMap["1"] || !idMap["2"] || !idMap["5"] {
		t.Errorf("expected IDs 1, 2, 5 to be found, got %v", idMap)
	}
}

func TestQuadtree_Remove(t *testing.T) {
	boundary := NewAABB(0, 0, 100)
	qt := NewQuadtree(boundary, 1) // Force split

	e1 := &Entity{ID: "1", Position: Vec2{X: 10, Y: 10}}
	e2 := &Entity{ID: "2", Position: Vec2{X: -10, Y: -10}}

	qt.Insert(e1)
	qt.Insert(e2)

	if !qt.Remove("1") {
		t.Errorf("failed to remove entity 1")
	}

	found := qt.QueryAABB(boundary, nil)
	if len(found) != 1 {
		t.Errorf("expected 1 entity after removal, found %d", len(found))
	}
	if found[0].ID != "2" {
		t.Errorf("expected entity 2 to remain, found %s", found[0].ID)
	}
}
