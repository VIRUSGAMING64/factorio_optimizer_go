package place

import "testing"

func TestNetRemovalCredit(t *testing.T) {
	// Two machines adjacent
	px := []int{0, 3}
	py := []int{0, 0}
	mw := []int{3, 3}
	mh := []int{3, 3}
	terms := []NetTerm{{Producers: []int{0}, Consumers: []int{1}, Weight: 2}}
	c := NetRemovalCredit(px, py, mw, mh, terms, 2.0)
	if c <= 0 {
		t.Fatalf("expected credit for adjacent pair, got %v", c)
	}
	// Far apart
	px[1] = 20
	c2 := NetRemovalCredit(px, py, mw, mh, terms, 2.0)
	if c2 >= c {
		t.Fatalf("far should score less: %v vs %v", c2, c)
	}
}

func TestPinPenalty(t *testing.T) {
	p := PinPenalty([]int{0}, []int{0}, []int{3}, []int{3}, []int{2})
	if p < 0 {
		t.Fatal(p)
	}
}
