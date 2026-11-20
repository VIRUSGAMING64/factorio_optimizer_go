package route

import (
	"testing"

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
