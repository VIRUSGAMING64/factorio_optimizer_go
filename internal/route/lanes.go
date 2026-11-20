// Package route – lane capacity model.
//
// LaneTracker tracks throughput usage on individual belt tiles so the router
// can safely reuse an existing belt of the same net when residual capacity
// remains. Factorio belts have two independent lanes (left/right relative to
// heading). Each lane has half the belt throughput; we leave a 5 % safety
// margin on the total.
package route

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"

// LaneTracker records used items/s per tile, split by left and right lane.
type LaneTracker struct {
	capPerLane float64           // usable throughput per lane
	left       map[int]float64   // tile index → used on left
	right      map[int]float64   // tile index → used on right
}

// NewLaneTracker builds a tracker for the given belt tier.
// Total usable capacity is 95 % of the tier throughput, split equally
// between the two lanes.
func NewLaneTracker(tier data.BeltTier) *LaneTracker {
	total := tier.Throughput * 0.95
	if total < 2 {
		total = 2
	}
	return &LaneTracker{
		capPerLane: total / 2,
		left:       make(map[int]float64),
		right:      make(map[int]float64),
	}
}

// RemainingLane returns unused capacity on a specific lane (never negative).
func (lt *LaneTracker) RemainingLane(idx int, lane Lane) float64 {
	u := lt.used(idx, lane)
	if u >= lt.capPerLane {
		return 0
	}
	return lt.capPerLane - u
}

// Remaining returns the sum of unused capacity on both lanes.
func (lt *LaneTracker) Remaining(idx int) float64 {
	return lt.RemainingLane(idx, LaneLeft) + lt.RemainingLane(idx, LaneRight)
}

// CanShareLane reports whether a specific lane can accept the rate.
func (lt *LaneTracker) CanShareLane(idx int, lane Lane, rate float64) bool {
	if rate <= 0 {
		return true
	}
	return lt.used(idx, lane)+rate <= lt.capPerLane+1e-6
}

// CanShare reports whether either lane (or the sum) can accept the rate.
// Prefer a single lane when possible.
func (lt *LaneTracker) CanShare(idx int, rate float64) bool {
	if rate <= 0 {
		return true
	}
	if lt.CanShareLane(idx, LaneLeft, rate) || lt.CanShareLane(idx, LaneRight, rate) {
		return true
	}
	// Allow splitting across both if necessary
	return lt.Remaining(idx) >= rate-1e-6
}

// AddLane records usage on a specific lane.
func (lt *LaneTracker) AddLane(idx int, lane Lane, rate float64) {
	if rate <= 0 {
		return
	}
	if lane == LaneLeft {
		lt.left[idx] += rate
	} else {
		lt.right[idx] += rate
	}
}

// Add records usage, preferring the lane with more remaining capacity.
func (lt *LaneTracker) Add(idx int, rate float64) {
	if rate <= 0 {
		return
	}
	// Prefer the emptier lane
	if lt.RemainingLane(idx, LaneLeft) >= lt.RemainingLane(idx, LaneRight) {
		if lt.CanShareLane(idx, LaneLeft, rate) {
			lt.AddLane(idx, LaneLeft, rate)
			return
		}
	}
	if lt.CanShareLane(idx, LaneRight, rate) {
		lt.AddLane(idx, LaneRight, rate)
		return
	}
	// Split if needed
	half := rate / 2
	lt.AddLane(idx, LaneLeft, half)
	lt.AddLane(idx, LaneRight, half)
}

// Used returns total used rate on a tile (both lanes).
func (lt *LaneTracker) Used(idx int) float64 {
	return lt.left[idx] + lt.right[idx]
}

// UsedLane returns used rate on a specific lane.
func (lt *LaneTracker) UsedLane(idx int, lane Lane) float64 {
	return lt.used(idx, lane)
}

func (lt *LaneTracker) used(idx int, lane Lane) float64 {
	if lane == LaneLeft {
		return lt.left[idx]
	}
	return lt.right[idx]
}

// CapPerLane returns the capacity of a single lane (for tests and costs).
func (lt *LaneTracker) CapPerLane() float64 {
	return lt.capPerLane
}
