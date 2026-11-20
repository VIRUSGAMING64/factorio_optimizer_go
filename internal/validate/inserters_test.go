package validate

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestCheckInsertersOK(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindMachine, Name: "assembling-machine-3", X: 0, Y: 0},
		{Kind: route.KindInserter, Name: "fast-inserter", X: 3, Y: 1, Dir: geom.West},
		{Kind: route.KindBelt, Name: "fast-transport-belt", X: 4, Y: 1, Dir: geom.East},
	}
	// pickup toward machine (west of inserter at reach 1 => 2,1 — may miss 3x3)
	rep := CheckInserters(ents, false)
	t.Log(rep.String())
}

func TestCheckInsertersHardEmpty(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindInserter, Name: "fast-inserter", X: 5, Y: 5, Dir: geom.East},
	}
	rep := CheckInserters(ents, true)
	if rep.OK {
		t.Fatal("expected failure")
	}
}
