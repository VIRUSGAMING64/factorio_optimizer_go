package validate

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestCheckLayoutDetailUGPair(t *testing.T) {
	ents := []route.Entity{
		{Name: "fast-underground-belt", Kind: route.KindUndergroundIn, X: 0, Y: 0, Dir: geom.East, UGType: "input", NetID: 1},
		{Name: "fast-underground-belt", Kind: route.KindUndergroundOut, X: 5, Y: 0, Dir: geom.East, UGType: "output", NetID: 1},
	}
	rep := CheckLayoutDetail(ents, 20, 10)
	if !rep.OK {
		t.Fatalf("paired UG should be OK: %s", rep.String())
	}
}

func TestCheckLayoutDetailUGUnpaired(t *testing.T) {
	ents := []route.Entity{
		{Name: "fast-underground-belt", Kind: route.KindUndergroundIn, X: 0, Y: 0, Dir: geom.East, UGType: "input", NetID: 1},
	}
	rep := CheckLayoutDetail(ents, 20, 10)
	if rep.OK {
		t.Fatal("unpaired UG should fail")
	}
}

func TestCheckLayoutDetailContamination(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindBelt, X: 1, Y: 1, Item: "iron-plate", NetID: 1},
		{Kind: route.KindBelt, X: 1, Y: 1, Item: "copper-plate", NetID: 2},
	}
	rep := CheckLayoutDetail(ents, 10, 10)
	if rep.OK {
		t.Fatal("contamination expected")
	}
}

func TestCheckLayoutDetailInserterNote(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindInserter, Name: "fast-inserter", X: 5, Y: 5, Dir: geom.East},
	}
	rep := CheckLayoutDetail(ents, 20, 20)
	// soft notes only
	t.Log(rep.String())
}
