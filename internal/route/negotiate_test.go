package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

func TestNegotiate(t *testing.T) {
	pf := NewPathFinder(20, 20, 6)
	pf.MarkMachine(9, 9, 2, 2)
	pairs := []EndpointPair{
		{Src: geom.Point{0, 0}, Dst: geom.Point{19, 0}, NetID: 0, Item: "a"},
		{Src: geom.Point{0, 19}, Dst: geom.Point{19, 19}, NetID: 1, Item: "b"},
	}
	res := pf.Negotiate(pairs, 3)
	if res.Iters < 1 {
		t.Fatal("iters")
	}
	ok := 0
	for _, p := range res.Paths {
		if p != nil {
			ok++
		}
	}
	if ok < 2 {
		t.Fatalf("paths ok=%d failed=%v", ok, res.Failed)
	}
}
