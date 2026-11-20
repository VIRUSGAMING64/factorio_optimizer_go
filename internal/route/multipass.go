package route

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// RouteMultiPass runs Route up to maxPasses times, each time increasing the
// present-factor of the pathfinder-style history so later nets avoid congestion
// from earlier attempts. Keeps the best result by (fewer failures, more belts).
func RouteMultiPass(p *plan.Plan, pl place.Placement, beltTier string, maxPasses int) Result {
	if maxPasses < 1 {
		maxPasses = 1
	}
	best := Route(p, pl, beltTier)
	if len(best.FailedNets) == 0 || maxPasses == 1 {
		return best
	}
	for pass := 1; pass < maxPasses; pass++ {
		rt := RerouteFailed(p, pl, beltTier, best, 2)
		if betterRoute(rt, best) {
			best = rt
		}
		if len(best.FailedNets) == 0 {
			break
		}
	}
	return best
}

func betterRoute(a, b Result) bool {
	if len(a.FailedNets) != len(b.FailedNets) {
		return len(a.FailedNets) < len(b.FailedNets)
	}
	if a.Belts != b.Belts {
		return a.Belts > b.Belts
	}
	return a.Undergrounds > b.Undergrounds
}
