// Package route – lane capacity model.
//
// LaneTracker tracks throughput usage on individual belt tiles so the router
// can safely reuse an existing belt of the same net when residual capacity
// remains. Factorio belts have two lanes; we conservatively treat total
// throughput as the sum of both lanes (half each) and leave a small safety
// margin.
package route

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"

// LaneTracker records used items/s per tile index.
type LaneTracker struct {
	cap  float64         // usable throughput per tile
	used map[int]float64 // tile index → used rate
}

// NewLaneTracker builds a tracker for the given belt tier.
// A 5 % safety margin is subtracted so we never claim a full belt.
func NewLaneTracker(tier data.BeltTier) *LaneTracker {
	c := tier.Throughput * 0.95
	if c < 1 {
		c = 1
	}
	return &LaneTracker{cap: c, used: make(map[int]float64)}
}

// Remaining returns the unused capacity on a tile (never negative).
func (lt *LaneTracker) Remaining(idx int) float64 {
	u := lt.used[idx]
	if u >= lt.cap {
		return 0
	}
	return lt.cap - u
}

// CanShare reports whether the tile can accept an additional rate.
func (lt *LaneTracker) CanShare(idx int, rate float64) bool {
	if rate <= 0 {
		return true
	}
	return lt.used[idx]+rate <= lt.cap+1e-6
}

// Add records additional usage on a tile.
func (lt *LaneTracker) Add(idx int, rate float64) {
	if rate <= 0 {
		return
	}
	lt.used[idx] += rate
}

// Used returns the currently recorded rate on a tile.
func (lt *LaneTracker) Used(idx int) float64 {
	return lt.used[idx]
}
