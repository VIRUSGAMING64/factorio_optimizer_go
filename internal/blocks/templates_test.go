package blocks

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestTryColumnTemplate(t *testing.T) {
	b := Block{
		Machines: []plan.MachineInstance{
			{ID: 0, Recipe: "iron-gear-wheel", Entity: "assembling-machine-3", W: 3, H: 3},
			{ID: 1, Recipe: "iron-gear-wheel", Entity: "assembling-machine-3", W: 3, H: 3},
			{ID: 2, Recipe: "iron-gear-wheel", Entity: "assembling-machine-3", W: 3, H: 3},
			{ID: 3, Recipe: "iron-gear-wheel", Entity: "assembling-machine-3", W: 3, H: 3},
		},
	}
	tr, ok := TryColumnTemplate(b, TemplateNarrow)
	if !ok {
		t.Fatal("expected template")
	}
	if len(tr.Placement.Machines) != 4 {
		t.Fatalf("machines %d", len(tr.Placement.Machines))
	}
	if len(tr.BeltTiles) == 0 {
		t.Fatal("no belt tiles")
	}
}

func TestTryBestTemplate(t *testing.T) {
	b := Block{
		Machines: []plan.MachineInstance{
			{Recipe: "copper-plate", Entity: "electric-furnace", W: 3, H: 3},
			{Recipe: "copper-plate", Entity: "electric-furnace", W: 3, H: 3},
		},
	}
	tr, ok := TryBestTemplate(b)
	if !ok || tr.Placement.W < 1 {
		t.Fatal("best template")
	}
}
