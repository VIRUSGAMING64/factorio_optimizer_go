package search

import (
	"fmt"
	"math"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// CalibrationResult holds measured averages over a few trials.
type CalibrationResult struct {
	AvgScore    float64
	AvgArea     float64
	AvgBelts    float64
	AvgFailures float64
	Samples     int
	Suggested   config.Tuning
}

// Calibrate runs several trials and suggests mild tuning adjustments.
func Calibrate(p *plan.Plan, base config.Tuning, beltTier string, samples int) (CalibrationResult, error) {
	if samples < 1 {
		samples = 3
	}
	var sumScore, sumArea, sumBelts, sumFail float64
	ok := 0
	for i := 0; i < samples; i++ {
		tr, err := RunTrials(p, Options{
			Attempts: 1, SAIters: max(5000, base.SAIterations/4),
			AreaFactor: base.AreaFactor, BeltTier: beltTier, Tuning: base,
		})
		if err != nil {
			continue
		}
		ok++
		sumScore += tr.Score
		sumArea += float64(tr.Placement.W * tr.Placement.H)
		sumBelts += float64(tr.Route.Belts)
		sumFail += float64(len(tr.Route.FailedNets))
	}
	if ok == 0 {
		return CalibrationResult{}, fmt.Errorf("calibration: all %d samples failed", samples)
	}
	n := float64(ok)
	res := CalibrationResult{
		AvgScore: sumScore / n, AvgArea: sumArea / n,
		AvgBelts: sumBelts / n, AvgFailures: sumFail / n, Samples: ok,
		Suggested: base,
	}
	// Heuristic tweaks
	if res.AvgFailures > 2 {
		res.Suggested.AreaFactor = math.Min(2.0, base.AreaFactor*1.1)
		res.Suggested.SAIterations = int(float64(base.SAIterations) * 1.2)
	}
	if res.AvgArea > 400 {
		res.Suggested.CongestionWt = base.CongestionWt * 1.1
	}
	return res, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
