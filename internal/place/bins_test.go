package place

import "testing"

func TestCongestionBinsPenalty(t *testing.T) {
	b := NewCongestionBins(20, 20, 4)
	px := []int{2, 3, 2, 15}
	py := []int{2, 2, 3, 15}
	mw := []int{1, 1, 1, 1}
	mh := []int{1, 1, 1, 1}
	b.Rebuild(px, py, mw, mh)
	p := b.Penalty(2)
	if p < 1 {
		t.Fatalf("expected overcrowding penalty, got %v counts=%v", p, b.Counts)
	}
}

func TestDICredit(t *testing.T) {
	px := []int{0, 4}
	py := []int{0, 0}
	mw := []int{3, 3}
	mh := []int{3, 3}
	c := DICredit(px, py, mw, mh, [][2]int{{0, 1}}, 1)
	if c < 1 {
		t.Fatalf("adjacent machines should get DI credit, got %v", c)
	}
}
