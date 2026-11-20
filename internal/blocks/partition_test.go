package blocks

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestPartition(t *testing.T) {
	p := &plan.Plan{}
	for i := 0; i < 50; i++ {
		p.Machines = append(p.Machines, plan.MachineInstance{
			ID: i, Recipe: "r" + string(rune('a'+i%5)), Entity: "assembling-machine-3",
			W: 3, H: 3, Crafts: 1, Products: []string{"out"}, Ingredients: []string{"in"},
		})
	}
	bs := Partition(p, 12)
	if len(bs) < 4 {
		t.Fatalf("expected several blocks, got %d", len(bs))
	}
	total := 0
	for _, b := range bs {
		total += len(b.Machines)
		if len(b.Machines) > 12 {
			t.Fatalf("block too big %d", len(b.Machines))
		}
	}
	if total != 50 {
		t.Fatalf("lost machines %d", total)
	}
}

func TestFloorplanPack(t *testing.T) {
	blocks := []FloorBlock{
		{ID: 0, Shapes: []Shape{{10, 8}}, Pad: 1},
		{ID: 1, Shapes: []Shape{{12, 10}}, Pad: 1},
		{ID: 2, Shapes: []Shape{{6, 6}}, Pad: 1},
	}
	fp := Pack(blocks, []FloorNet{{BlockIDs: []int{0, 1}, Weight: 1}}, 1.0, 5)
	if fp.W == 0 || fp.H == 0 {
		t.Fatal("empty floorplan")
	}
	if len(fp.Boxes) != 3 {
		t.Fatalf("boxes %d", len(fp.Boxes))
	}
}
