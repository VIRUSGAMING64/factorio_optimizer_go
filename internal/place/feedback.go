package place

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"

// RouteFeedback aggregates belt congestion and failed nets so a subsequent SA
// run can bias placements away from hard-to-route regions.
type RouteFeedback struct {
	TileCost  map[geom.Point]float64
	FailCount int
}

// FeedbackFromHist builds feedback from a congestion history grid and fail count.
func FeedbackFromHist(hist []float64, w, h, failCount int) RouteFeedback {
	fb := RouteFeedback{
		TileCost:  map[geom.Point]float64{},
		FailCount: failCount,
	}
	if len(hist) == w*h {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				v := hist[y*w+x]
				if v > 2 {
					fb.TileCost[geom.Point{x, y}] = v * 0.5
				}
			}
		}
	}
	return fb
}

// ScoreWithFeedback adds tile costs into the placement score.
func (pl *Placement) ScoreWithFeedback(fb RouteFeedback) float64 {
	s := pl.Cost + float64(fb.FailCount)*50
	for _, m := range pl.Machines {
		for dy := 0; dy < 3; dy++ {
			for dx := 0; dx < 3; dx++ {
				s += fb.TileCost[geom.Point{m.X + dx, m.Y + dy}]
			}
		}
	}
	return s
}
