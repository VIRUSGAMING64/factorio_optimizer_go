package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestMergeBeltPaths(t *testing.T) {
	shared, branches := MergeBeltPaths([][]geom.Point{
		{{0, 0}, {1, 0}, {2, 0}, {3, 0}},
		{{0, 0}, {1, 0}, {2, 1}, {2, 2}},
	})
	if len(shared) < 2 {
		t.Fatalf("shared %v", shared)
	}
	if len(branches) < 1 {
		t.Fatalf("branches %v", branches)
	}
}

func TestPlaceSplittersOnBranch(t *testing.T) {
	ents := []Entity{
		{Kind: KindBelt, X: 5, Y: 5, Dir: geom.East, NetID: 1, Item: "x"},
		{Kind: KindBelt, X: 6, Y: 5, Dir: geom.East, NetID: 1, Item: "x"},
		{Kind: KindBelt, X: 5, Y: 4, Dir: geom.North, NetID: 1, Item: "x"},
		{Kind: KindBelt, X: 5, Y: 6, Dir: geom.South, NetID: 1, Item: "x"},
	}
	sp := PlaceSplittersOnBranch(ents, data.BeltTierLookup("fast"), plan.Net{ID: 1, Item: "x"})
	t.Logf("splitters=%d", len(sp))
}
