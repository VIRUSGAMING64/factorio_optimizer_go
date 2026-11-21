package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

func TestPreferredLane(t *testing.T) {
	// East-bound belt: left side is North, right side is South.
	if got := PreferredLane(geom.East, geom.North); got != LaneLeft {
		t.Errorf("east+north: got %v, want LaneLeft", got)
	}
	if got := PreferredLane(geom.East, geom.South); got != LaneRight {
		t.Errorf("east+south: got %v, want LaneRight", got)
	}
	// North-bound belt: left side is West.
	if got := PreferredLane(geom.North, geom.West); got != LaneLeft {
		t.Errorf("north+west: got %v, want LaneLeft", got)
	}
	// Parallel approach falls back to left.
	if got := PreferredLane(geom.East, geom.East); got != LaneLeft {
		t.Errorf("parallel: got %v, want LaneLeft", got)
	}
}

func TestApproachSide(t *testing.T) {
	ao := AccessOption{
		BeltTile: geom.Point{5, 5},
		InsTile:  geom.Point{5, 4}, // north of belt
	}
	if got := ApproachSide(ao); got != geom.North {
		t.Errorf("got %v, want North", got)
	}
	ao.InsTile = geom.Point{6, 5}
	if got := ApproachSide(ao); got != geom.East {
		t.Errorf("got %v, want East", got)
	}
}

func TestApplyCommitPolarity(t *testing.T) {
	// Eastbound first step; approach from north should prefer LaneLeft.
	path := []geom.Point{{0, 0}, {1, 0}}
	lanes := []Lane{LaneRight, LaneLeft}
	access := []AccessOption{{
		BeltTile: geom.Point{0, 0},
		InsTile:  geom.Point{0, -1}, // north of belt
	}}
	got := applyCommitPolarity(path, lanes, access)
	if got[0] != LaneLeft {
		t.Errorf("expected LaneLeft for east+north, got %v", got[0])
	}
	// Southbound first step; approach from west should prefer LaneRight.
	path2 := []geom.Point{{0, 0}, {0, 1}}
	lanes2 := []Lane{LaneLeft}
	access2 := []AccessOption{{
		BeltTile: geom.Point{0, 0},
		InsTile:  geom.Point{-1, 0}, // west
	}}
	got2 := applyCommitPolarity(path2, lanes2, access2)
	if got2[0] != LaneRight {
		t.Errorf("expected LaneRight for south+west, got %v", got2[0])
	}
}
