package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

func TestLaneTrackerBasic(t *testing.T) {
	// Fast belt ~30 items/s; with 5 % margin usable ≈ 28.5, ~14.25 per lane
	tier := data.BeltTier{Key: "fast", Throughput: 30, MaxUndergroundGap: 9}
	lt := NewLaneTracker(tier)
	if lt.CapPerLane() <= 0 {
		t.Fatal("capacity must be positive")
	}

	idx := 42
	if !lt.CanShare(idx, 10) {
		t.Fatal("empty tile should accept 10")
	}
	lt.Add(idx, 10)
	if lt.Used(idx) != 10 {
		t.Fatalf("used = %v want 10", lt.Used(idx))
	}
	if lt.Remaining(idx) >= lt.CapPerLane()*2 {
		t.Fatal("remaining should have decreased")
	}

	// Prefer left when equal
	if lt.UsedLane(idx, LaneLeft) == 0 && lt.UsedLane(idx, LaneRight) == 0 {
		t.Fatal("expected usage on one lane")
	}

	// Fill one lane almost completely
	lt.AddLane(idx, LaneLeft, lt.CapPerLane())
	if lt.CanShareLane(idx, LaneLeft, 1) {
		t.Fatal("left should be full")
	}
	if !lt.CanShareLane(idx, LaneRight, 5) {
		t.Fatal("right should still accept")
	}
	if !lt.CanShare(idx, 5) {
		t.Fatal("total should still accept on the free lane")
	}
}

func TestLaneTrackerSplit(t *testing.T) {
	tier := data.BeltTier{Throughput: 40}
	lt := NewLaneTracker(tier)
	idx := 1
	// Force a split by filling both partially then adding a large amount
	lt.AddLane(idx, LaneLeft, 15)
	lt.AddLane(idx, LaneRight, 15)
	if !lt.CanShare(idx, 8) {
		t.Fatal("should accept by splitting")
	}
	lt.Add(idx, 8)
	if lt.Used(idx) < 38-1e-3 {
		t.Fatalf("expected near full, got %v", lt.Used(idx))
	}
}
