package blocks

import (
	"math/rand"
	"testing"
)

func TestPackSequencePairNonOverlap(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	ws := []int{4, 6, 3, 5, 4}
	hs := []int{5, 3, 4, 4, 6}
	plus, minus := randomSequencePair(rng, len(ws))
	x, y, maxX, maxY := packSequencePair(ws, hs, plus, minus)

	if maxX <= 0 || maxY <= 0 {
		t.Fatal("degenerate packing")
	}
	// Verify no overlaps
	for i := 0; i < len(ws); i++ {
		for j := i + 1; j < len(ws); j++ {
			if rectsOverlap(x[i], y[i], ws[i], hs[i], x[j], y[j], ws[j], hs[j]) {
				t.Errorf("overlap between %d and %d", i, j)
			}
		}
	}
	area, _, _ := ScoreSequencePair(ws, hs, plus, minus)
	if area < float64(maxX*maxY)-1 {
		t.Errorf("score mismatch")
	}
}

func rectsOverlap(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
	return x1 < x2+w2 && x2 < x1+w1 && y1 < y2+h2 && y2 < y1+h1
}

func TestPackUsesSequencePair(t *testing.T) {
	blocks := []FloorBlock{
		{ID: 0, Shapes: []Shape{{8, 6}}, Pad: 1},
		{ID: 1, Shapes: []Shape{{5, 5}}, Pad: 1},
		{ID: 2, Shapes: []Shape{{4, 7}}, Pad: 1},
		{ID: 3, Shapes: []Shape{{6, 4}}, Pad: 1},
	}
	nets := []FloorNet{
		{BlockIDs: []int{0, 1}, Weight: 2},
		{BlockIDs: []int{1, 2, 3}, Weight: 1},
	}
	fp := Pack(blocks, nets, 1.2, 12)
	if fp.W <= 0 || fp.H <= 0 || len(fp.Boxes) != 4 {
		t.Fatalf("bad floorplan %+v", fp)
	}
	if fp.Cost == 0 || fp.Cost == 1e308 { // Inf
		t.Fatal("invalid cost")
	}
}

func TestImproveSequencePair(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	ws := []int{4, 6, 3, 5, 4, 7}
	hs := []int{5, 3, 4, 4, 6, 3}
	plus, minus := randomSequencePair(rng, len(ws))
	startArea, _, _ := ScoreSequencePair(ws, hs, plus, minus)
	bp, bm, bestArea := improveSequencePair(ws, hs, plus, minus, 40, rng)
	if bestArea > startArea+1e-6 {
		t.Fatalf("improve worsened area: %v -> %v", startArea, bestArea)
	}
	// Still non-overlapping
	x, y, _, _ := packSequencePair(ws, hs, bp, bm)
	for i := 0; i < len(ws); i++ {
		for j := i + 1; j < len(ws); j++ {
			if rectsOverlap(x[i], y[i], ws[i], hs[i], x[j], y[j], ws[j], hs[j]) {
				t.Errorf("overlap after improve")
			}
		}
	}
}
