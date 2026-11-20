package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

func TestFindPathWithLanes(t *testing.T) {
	pf := NewPathFinder(15, 15, 6)
	path, lanes := pf.FindPathWithLanes(geom.Point{0, 0}, geom.Point{10, 5}, 0, LaneLeft)
	if path == nil || len(lanes) != len(path) {
		t.Fatalf("path=%v lanes=%d", path != nil, len(lanes))
	}
}

func TestPresentBoostAndConflicts(t *testing.T) {
	pf := NewPathFinder(8, 8, 4)
	path := pf.FindPath(geom.Point{0, 0}, geom.Point{7, 0}, 0)
	pf.CommitPath(path, 0)
	pf.PresentBoost(2)
	if pf.Present < 1.5 {
		t.Fatal("boost")
	}
	c := pf.ConflictTiles(0.5)
	if len(c) == 0 {
		t.Fatal("expected hist tiles")
	}
}
