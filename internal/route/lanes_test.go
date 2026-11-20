package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

func TestLaneTrackerBasic(t *testing.T) {
	// Fast belt ~30 items/s; with 5 % margin usable ≈ 28.5
	tier := data.BeltTier{Key: "fast", Throughput: 30, MaxUndergroundGap: 9}
	lt := NewLaneTracker(tier)
	if lt.cap <= 0 {
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
	if lt.Remaining(idx) >= lt.cap {
		t.Fatal("remaining should have decreased")
	}

	// Fill almost to capacity
	lt.Add(idx, lt.cap-10)
	if lt.CanShare(idx, 1) {
		t.Fatal("should reject when full")
	}
	if lt.Remaining(idx) > 1e-3 {
		t.Fatalf("remaining %v expected ~0", lt.Remaining(idx))
	}
}
