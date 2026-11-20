package search

import (
	"time"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// TrialWithFeedback runs place→route, then a second SA biased by route congestion history
// when the first attempt left failed nets.
func TrialWithFeedback(p *plan.Plan, opts Options) (Trial, error) {
	first, err := RunTrials(p, opts)
	if err != nil {
		return first, err
	}
	if len(first.Route.FailedNets) == 0 {
		return first, nil
	}
	w, h := first.Placement.W, first.Placement.H
	if w < 1 || h < 1 {
		return first, nil
	}
	hist := make([]float64, w*h)
	for pt, v := range first.Route.History {
		if pt.X >= 0 && pt.Y >= 0 && pt.X < w && pt.Y < h {
			hist[pt.Y*w+pt.X] = v
		}
	}
	fb := place.FeedbackFromHist(hist, w, h, len(first.Route.FailedNets))

	iters := opts.SAIters
	if iters < 1 {
		iters = opts.Tuning.SAIterations
	}
	t0 := time.Now()
	pl2, err := place.Place(p, opts.AreaFactor*1.1, iters*2, opts.Tuning.SAInitialTemp, opts.Tuning.SACoolingRate)
	if err != nil {
		return first, nil
	}
	s1 := first.Placement.ScoreWithFeedback(fb)
	s2 := pl2.ScoreWithFeedback(fb)
	pl := pl2
	if s1 <= s2 {
		pl = first.Placement
	}
	rt := route.RouteMultiPass(p, pl, opts.BeltTier, 3)
	out := Trial{
		Placement: pl,
		Route:     rt,
		Score:     scoreTrial(pl, rt),
		OK:        rt.Success || rt.Belts > 0,
		Duration:  time.Since(t0),
	}
	if out.Score < first.Score {
		return out, nil
	}
	return first, nil
}
