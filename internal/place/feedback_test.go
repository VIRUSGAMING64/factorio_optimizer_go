package place

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

func TestFeedbackFromHist(t *testing.T) {
	hist := make([]float64, 100)
	hist[55] = 5
	fb := FeedbackFromHist(hist, 10, 10, 2)
	if fb.FailCount != 2 {
		t.Fatal("fail count")
	}
	if fb.TileCost[geom.Point{5, 5}] < 1 {
		t.Fatalf("tile cost %v", fb.TileCost[geom.Point{5, 5}])
	}
	pl := Placement{W: 10, H: 10, Cost: 100, Machines: []PlacedMachine{{X: 4, Y: 4}}}
	s := pl.ScoreWithFeedback(fb)
	if s <= 100 {
		t.Fatalf("score %v", s)
	}
}
