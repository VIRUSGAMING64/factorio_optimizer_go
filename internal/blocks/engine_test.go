package blocks

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestBuildFactorySmall(t *testing.T) {
	root, err := data.LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = data.LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.RunConfiguration{
		Machine: "assembling-machine-3", BeltTier: "fast", Quality: "normal",
		Targets: []config.ConfigTarget{
			{Name: "automation-science-pack", Amount: 0.3},
			{Name: "logistic-science-pack", Amount: 0.3},
		},
	}
	cfg.ApplyDefaults()
	p, err := plan.Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	// force multiple blocks with small max
	fr, err := BuildFactory(p, cfg, config.Defaults(), 5, 3000)
	if err != nil {
		t.Fatal(err)
	}
	if len(fr.Blocks) < 1 {
		t.Fatal("no blocks")
	}
	if len(fr.Entities) == 0 {
		t.Fatal("no entities")
	}
	t.Log(fr.Message)
}

func TestBlockPlan(t *testing.T) {
	b := Block{
		ID: 0,
		Machines: []plan.MachineInstance{
			{ID: 0, Recipe: "iron-gear-wheel", Entity: "assembling-machine-3", W: 3, H: 3, Crafts: 1,
				Products: []string{"iron-gear-wheel"}, Ingredients: []string{"iron-plate"}},
		},
		Inputs:  []string{"iron-plate"},
		Outputs: []string{"iron-gear-wheel"},
	}
	bp := BlockPlan(b)
	if len(bp.Machines) != 1 || len(bp.Ports) < 2 {
		t.Fatal("block plan")
	}
}
