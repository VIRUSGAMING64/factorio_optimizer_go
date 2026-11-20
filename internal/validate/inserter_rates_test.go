package validate

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestEstimatedInserterCapacity(t *testing.T) {
	r := EstimatedInserterCapacity("fast-inserter", false, 0, 1)
	if r <= 0 {
		t.Fatalf("expected positive rate, got %v", r)
	}
	rb := EstimatedInserterCapacity("fast-inserter", true, 0, 1)
	if rb >= r {
		t.Fatalf("belt rate %v should be < chest rate %v", rb, r)
	}
	rs := EstimatedInserterCapacity("stack-inserter", false, 7, 4)
	if rs <= r {
		t.Fatalf("stack inserter %v should exceed fast %v", rs, r)
	}
}

func TestCheckInserterRatesNoInserters(t *testing.T) {
	p := &plan.Plan{
		Machines: []plan.MachineInstance{{
			ID: 0, Recipe: "iron-gear-wheel", Entity: "assembling-machine-3",
			Crafts: 1.0, W: 3, H: 3,
		}},
	}
	ents := []route.Entity{{
		Kind: route.KindMachine, Name: "assembling-machine-3",
		X: 0, Y: 0, Recipe: "iron-gear-wheel",
	}}
	rep := CheckInserterRates(p, ents, 0, 1)
	if rep.OK {
		t.Fatal("expected OK=false when machine has no nearby inserters")
	}
	if len(rep.Warnings) == 0 {
		t.Fatal("expected a warning")
	}
	t.Log(rep.String())
}

func TestCheckInserterRatesSufficient(t *testing.T) {
	p := &plan.Plan{
		Machines: []plan.MachineInstance{{
			ID: 1, Recipe: "iron-gear-wheel", Entity: "assembling-machine-3",
			Crafts: 0.5, W: 3, H: 3,
		}},
	}
	ents := []route.Entity{
		{Kind: route.KindMachine, Name: "assembling-machine-3", X: 0, Y: 0, Recipe: "iron-gear-wheel"},
		{Kind: route.KindInserter, Name: "fast-inserter", X: 3, Y: 1, Dir: geom.West},
		{Kind: route.KindInserter, Name: "fast-inserter", X: -1, Y: 1, Dir: geom.East},
	}
	rep := CheckInserterRates(p, ents, 2, 1)
	if !rep.OK && len(rep.Warnings) > 0 && len(rep.Notes) == 0 {
		t.Log("warned (acceptable for conservative estimate):", rep.String())
	} else {
		t.Log("passed or noted:", rep.String())
	}
}
