package validate

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestCheckBeltShapesOK(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindBelt, X: 0, Y: 0, Dir: geom.East, NetID: 1},
		{Kind: route.KindBelt, X: 1, Y: 0, Dir: geom.East, NetID: 1},
	}
	r := CheckBeltShapes(ents)
	if !r.OK {
		t.Fatal(r.String())
	}
}

func TestCheckBeltShapesMismatch(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindBelt, X: 0, Y: 0, Dir: geom.East, NetID: 1},
		{Kind: route.KindBelt, X: 1, Y: 0, Dir: geom.East, NetID: 2},
	}
	r := CheckBeltShapes(ents)
	if r.OK {
		t.Fatal("expected shape error")
	}
}

func TestCheckLaneContamination(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindBelt, X: 1, Y: 1, Item: "a"},
		{Kind: route.KindBelt, X: 1, Y: 1, Item: "b"},
	}
	errs := CheckLaneContamination(ents)
	if len(errs) == 0 {
		t.Fatal("expected contamination")
	}
}
