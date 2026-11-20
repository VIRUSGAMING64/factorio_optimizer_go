package blocks

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestSplitByRecipe(t *testing.T) {
	p := &plan.Plan{}
	for i := 0; i < 10; i++ {
		r := "recipe-a"
		if i >= 6 {
			r = "recipe-b"
		}
		p.Machines = append(p.Machines, plan.MachineInstance{
			ID: i, Recipe: r, Entity: "assembling-machine-3", W: 3, H: 3,
			Products: []string{"out"}, Ingredients: []string{"in"},
		})
	}
	bs := SplitByRecipe(p)
	if len(bs) != 2 {
		t.Fatalf("blocks %d", len(bs))
	}
}

func TestMergeTinyBlocks(t *testing.T) {
	bs := []Block{
		{ID: 0, Machines: make([]plan.MachineInstance, 2)},
		{ID: 1, Machines: make([]plan.MachineInstance, 10)},
		{ID: 2, Machines: make([]plan.MachineInstance, 1)},
	}
	out := MergeTinyBlocks(bs, 5)
	if len(out) >= len(bs) {
		t.Fatalf("expected merge, got %d", len(out))
	}
}
