package validate

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestBeltGraphPropagate(t *testing.T) {
	ents := []route.Entity{
		{Name: "fast-transport-belt", Kind: route.KindBelt, X: 0, Y: 0, Dir: geom.East, Item: "iron-plate"},
		{Name: "fast-transport-belt", Kind: route.KindBelt, X: 1, Y: 0, Dir: geom.East, Item: "iron-plate"},
		{Name: "fast-transport-belt", Kind: route.KindBelt, X: 2, Y: 0, Dir: geom.East, Item: "iron-plate"},
	}
	g := BuildBeltGraph(ents)
	if len(g.Nodes) != 3 {
		t.Fatalf("nodes=%d", len(g.Nodes))
	}
	g.PropagateItems()
	for i := range g.Nodes {
		if !g.Items[i]["iron-plate"] {
			t.Fatalf("node %d missing iron-plate", i)
		}
	}
	rep := CheckLayout(ents)
	if !rep.OK {
		t.Fatalf("layout: %s", rep.String())
	}
}
