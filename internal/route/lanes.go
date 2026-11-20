// Package route – lane capacity model.
//
// LaneTracker tracks throughput usage on individual belt tiles so the router
// can safely reuse an existing belt of the same net when residual capacity
// remains. Factorio belts have two independent lanes (left/right relative to
// travel direction); we model each at half of 95% of the tier throughput.
package route

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"

// LaneTracker records used items/s per tile, split by left and right lane.
type LaneTracker struct {
	capPerLane float64           // usable throughput per lane
	left       map[int]float64   // tile idx → used on left lane
	right      map[int]float64   // tile idx → used on right lane
}

// NewLaneTracker builds a tracker for the given belt tier.
// Total usable capacity is 95 % of the tier throughput, split equally.
func NewLaneTracker(tier data.BeltTier) *LaneTracker {
	total := tier.Throughput * 0.95
	if total < 1 {
		total = 15 * 0.95 // fallback fast belt
	}
	return &LaneTracker{
		capPerLane: total / 2,
		left:       map[int]float64{},
		right:      map[int]float64{},
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

func (lt *LaneTracker) CanShareLane(idx int, lane Lane, rate float64) bool {
	if rate <= 0 {
		return true
	}
	return lt.used(idx, lane)+rate <= lt.capPerLane+1e-6
}

func (lt *LaneTracker) CanShare(idx int, rate float64) bool {
	if rate <= 0 {
		return true
	}
	return lt.Remaining(idx)+1e-6 >= rate
}

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
	rl := lt.RemainingLane(idx, LaneLeft)
	rr := lt.RemainingLane(idx, LaneRight)
	if rl >= rr {
		lt.AddLane(idx, LaneLeft, rate)
	} else {
		lt.AddLane(idx, LaneRight, rate)
	}
}

func (lt *LaneTracker) Used(idx int) float64 {
	return lt.used(idx, LaneLeft) + lt.used(idx, LaneRight)
}

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
