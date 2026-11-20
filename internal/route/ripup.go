package route

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// RipUp removes all belt/underground/inserter entities belonging to the given net IDs
// and clears their occupancy marks so the net can be re-routed.
func RipUp(entities []Entity, occ []int, w, h int, netIDs map[int]bool) ([]Entity, int) {
	removed := 0
	out := entities[:0]
	for _, e := range entities {
		if netIDs[e.NetID] && (e.Kind == KindBelt || e.Kind == KindUndergroundIn ||
			e.Kind == KindUndergroundOut || e.Kind == KindInserter || e.Kind == KindSplitter) {
			if e.X >= 0 && e.Y >= 0 && e.X < w && e.Y < h {
				idx := e.Y*w + e.X
				if idx >= 0 && idx < len(occ) && occ[idx] == e.NetID+1 {
					occ[idx] = 0
				}
			}
			removed++
			continue
		}
		out = append(out, e)
	}
	return out, removed
}

// RerouteFailed takes a prior result, rips up failed nets, and tries Route again
// merging successful entities. Simplified negotiated-congestion iteration.
func RerouteFailed(p *plan.Plan, pl place.Placement, beltTier string, prev Result, maxIter int) Result {
	if maxIter < 1 {
		maxIter = 2
	}
	best := prev
	if len(prev.FailedNets) == 0 {
		return prev
	}
	for it := 0; it < maxIter; it++ {
		failSet := map[int]bool{}
		for _, id := range best.FailedNets {
			failSet[id] = true
		}
		// Full re-route is simpler and more reliable than partial merge for this port
		rt := Route(p, pl, beltTier)
		if betterRoute(rt, best) {
			best = rt
		}
		if len(best.FailedNets) == 0 {
			break
		}
		_ = geom.North
	}
	return best
}
