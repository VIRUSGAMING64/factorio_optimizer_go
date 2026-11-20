package place

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestPlaceNoOverlap(t *testing.T) {
	root, err := data.LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = data.LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.RunConfiguration{
		Machine: "assembling-machine-3", BeltTier: "fast", Quality: "normal",
		Targets: []config.ConfigTarget{{Name: "iron-gear-wheel", Amount: 2}},
	}
	cfg.ApplyDefaults()
	p, err := plan.Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	pl, err := Place(p, 1.5, 8000, 1000, 0.995)
	if err != nil {
		t.Fatal(err)
	}
	if len(pl.Machines) != len(p.Machines) {
		t.Fatalf("placed %d want %d", len(pl.Machines), len(p.Machines))
	}
	// no overlaps
	occ := map[[2]int]int{}
	for _, m := range pl.Machines {
		mi := p.Machines[m.ID]
		for dy := 0; dy < mi.H; dy++ {
			for dx := 0; dx < mi.W; dx++ {
				k := [2]int{m.X + dx, m.Y + dy}
				if prev, ok := occ[k]; ok {
					t.Fatalf("overlap at %v machines %d %d", k, prev, m.ID)
				}
				occ[k] = m.ID
				if m.X+dx < 0 || m.Y+dy < 0 || m.X+dx >= pl.W || m.Y+dy >= pl.H {
					t.Fatalf("out of bounds machine %d", m.ID)
				}
			}
		}
	}
	ascii := pl.ASCII(p)
	if len(ascii) < 10 {
		t.Fatal("empty ascii")
	}
}
