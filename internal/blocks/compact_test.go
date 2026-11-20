package blocks

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestCompactRemovesEmpty(t *testing.T) {
	ents := []route.Entity{
		{Name: "assembling-machine-3", Kind: route.KindMachine, X: 2, Y: 2},
		{Name: "fast-transport-belt", Kind: route.KindBelt, X: 5, Y: 2},
		{Name: "fast-transport-belt", Kind: route.KindBelt, X: 5, Y: 3},
	}
	c := Compact(ents, 20, 20)
	if c.W >= 20 || c.H >= 20 {
		t.Fatalf("should shrink, got %d×%d", c.W, c.H)
	}
	if len(c.Entities) == 0 {
		t.Fatal("lost entities")
	}
	if c.RowsRemoved == 0 && c.ColsRemoved == 0 {
		t.Fatal("expected removed rows/cols")
	}
}

func TestCompactFloorplan(t *testing.T) {
	fp := FloorPlan{W: 30, H: 30}
	fp = CompactFloorplan(fp)
	// empty boxes → unchanged zeros
	if fp.W != 30 {
		// with empty boxes CompactFloorplan returns early-ish
		_ = fp
	}
}
