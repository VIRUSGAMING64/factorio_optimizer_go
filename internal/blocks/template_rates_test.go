package blocks

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestInserterNeed(t *testing.T) {
	if InserterNeed(20, 7.5) < 2 {
		t.Fatal("high rate")
	}
	if InserterNeed(1, 7.5) != 1 {
		t.Fatal("low rate")
	}
}

func TestTemplateWithInserters(t *testing.T) {
	b := Block{Machines: []plan.MachineInstance{
		{W: 3, H: 3, Crafts: 5, Entity: "assembling-machine-3"},
		{W: 3, H: 3, Crafts: 5, Entity: "assembling-machine-3"},
	}}
	pl, needs, ok := TemplateWithInserters(b, 15)
	if !ok || len(needs) != 2 || pl.W < 1 {
		t.Fatalf("ok=%v needs=%v pl=%v", ok, needs, pl)
	}
}
