package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

func TestFindPathLaneAStar(t *testing.T) {
	pf := NewPathFinder(20, 20, 6)
	pf.MarkMachine(8, 8, 3, 3)
	path, lanes := pf.FindPathLaneAStar(geom.Point{0, 0}, geom.Point{19, 19}, 0, LaneLeft)
	if path == nil {
		t.Fatal("nil path")
	}
	if len(lanes) != len(path) {
		t.Fatalf("lanes %d path %d", len(lanes), len(path))
	}
}

func TestFindPathLaneAStarCapacity(t *testing.T) {
	pf := NewPathFinder(10, 10, 4)
	lt := NewLaneTracker(data.BeltTier{Throughput: 30})
	// Pre-load one lane heavily so search prefers the other
	for y := 0; y < 5; y++ {
		lt.AddLane(pf.idx(5, y), LaneLeft, lt.CapPerLane()*0.9)
	}
	path, lanes := pf.FindPathLaneAStarCapacity(geom.Point{0, 0}, geom.Point{9, 9}, 0, LaneLeft, lt)
	if path == nil || len(lanes) != len(path) {
		t.Fatal("capacity-aware path missing")
	}
}
