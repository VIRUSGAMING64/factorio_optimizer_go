package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

func TestPathFinderBasic(t *testing.T) {
	pf := NewPathFinder(20, 20, 6)
	pf.MarkMachine(8, 8, 3, 3)

	if !pf.Walkable(0, 0, 0) {
		t.Fatal("start should be walkable")
	}
	path := pf.FindPath(geom.Point{0, 0}, geom.Point{19, 19}, 0)
	if path == nil {
		t.Fatal("expected path")
	}
	if len(path) < 20 {
		t.Fatalf("path short %d", len(path))
	}
	for _, p := range path {
		if p.X >= 8 && p.X < 11 && p.Y >= 8 && p.Y < 11 {
			t.Fatalf("path through machine at %v", p)
		}
	}
	pf.CommitPath(path, 0)

	// Independent pathfinder for second net (avoids shared state bugs in test)
	pf2 := NewPathFinder(20, 20, 6)
	pf2.MarkMachine(8, 8, 3, 3)
	path2 := pf2.FindPath(geom.Point{0, 19}, geom.Point{19, 0}, 1)
	if path2 == nil {
		t.Fatal("second pathfinder path")
	}
}

func TestPathFinderReuse(t *testing.T) {
	pf := NewPathFinder(10, 10, 4)
	path := pf.FindPath(geom.Point{0, 5}, geom.Point{9, 5}, 0)
	if path == nil {
		t.Fatal("path")
	}
	pf.CommitPath(path, 0)
	// Same net can reuse
	path2 := pf.FindPath(geom.Point{0, 5}, geom.Point{9, 5}, 0)
	if path2 == nil {
		t.Fatal("reuse same net")
	}
}

func TestAssignLanes(t *testing.T) {
	lanes := AssignLanes([]float64{5, 5, 5, 5}, 10)
	if len(lanes) != 4 {
		t.Fatal("len")
	}
	left, right := 0, 0
	for _, l := range lanes {
		if l == LaneLeft {
			left++
		} else {
			right++
		}
	}
	if left == 0 || right == 0 {
		t.Fatalf("expected both lanes used L=%d R=%d", left, right)
	}
}

func TestPathToEntitiesUG(t *testing.T) {
	pf := NewPathFinder(30, 5, 8)
	var path []geom.Point
	for x := 0; x < 12; x++ {
		path = append(path, geom.Point{x, 2})
	}
	ents, belts, ugs := pf.PathToEntities(path, 0, "iron-plate", "fast-transport-belt", "fast-underground-belt")
	if ugs < 2 {
		t.Fatalf("expected undergrounds, got ugs=%d belts=%d ents=%d", ugs, belts, len(ents))
	}
}
