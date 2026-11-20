// Package search orchestrates multi-attempt placement+route trials.
package search

import (
	"fmt"
	"time"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// Trial is one attempt at place+route.
type Trial struct {
	Placement place.Placement
	Route     route.Result
	Score     float64
	OK        bool
	Duration  time.Duration
}

// Options controls the search loop.
type Options struct {
	Attempts   int
	SAIters    int
	AreaFactor float64
	BeltTier  string
	Tuning     config.Tuning
}

// RunTrials tries several place+route combinations and returns the best by score.
func RunTrials(p *plan.Plan, opts Options) (Trial, error) {
	if opts.Attempts < 1 {
		opts.Attempts = 3
	}
	if opts.SAIters < 1 {
		opts.SAIters = opts.Tuning.SAIterations
	}
	if opts.SAIters < 1 {
		opts.SAIters = 20000
	}
	if opts.AreaFactor < 1.1 {
		opts.AreaFactor = opts.Tuning.AreaFactor
	}
	if opts.AreaFactor < 1.1 {
		opts.AreaFactor = 1.4
	}
	if opts.BeltTier == "" {
		opts.BeltTier = "fast"
	}

	best := Trial{Score: 1e18}
	var lastErr error
	for i := 0; i < opts.Attempts; i++ {
		t0 := time.Now()
		pl, err := place.Place(p, opts.AreaFactor, opts.SAIters, opts.Tuning.SAInitialTemp, opts.Tuning.SACoolingRate)
		if err != nil {
			lastErr = err
			continue
		}
		rt := route.RouteMultiPass(p, pl, opts.BeltTier, 3)
		score := scoreTrial(pl, rt)
		tr := Trial{
			Placement: pl,
			Route:     rt,
			Score:     score,
			OK:        rt.Success || rt.Belts > 0,
			Duration:  time.Since(t0),
		}
		if tr.Score < best.Score {
			best = tr
		}
	}
	if best.Placement.Machines == nil {
		if lastErr != nil {
			return best, lastErr
		}
		return best, fmt.Errorf("all %d trials failed", opts.Attempts)
	}
	return best, nil
}

func scoreTrial(pl place.Placement, rt route.Result) float64 {
	// Lower is better: area + failures - coverage bonus
	area := float64(pl.W * pl.H)
	fail := float64(len(rt.FailedNets)) * 200
	// penalize missing belts heavily
	if rt.Belts == 0 {
		fail += 500
	}
	return area + float64(rt.Belts)*0.05 + fail + pl.Cost*0.01 - float64(rt.Undergrounds)*0.02
}
