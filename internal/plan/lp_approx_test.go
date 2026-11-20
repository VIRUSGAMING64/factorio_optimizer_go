package plan

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

func TestRefineMachineCounts(t *testing.T) {
	root, err := data.LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = data.LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.RunConfiguration{
		Machine: "assembling-machine-3", Quality: "normal",
		Targets: []config.ConfigTarget{{Name: "iron-gear-wheel", Amount: 3}},
	}
	cfg.ApplyDefaults()
	p, err := Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	before := len(p.Machines)
	RefineMachineCounts(root, cfg, p)
	if len(p.Machines) == 0 {
		t.Fatal("wiped")
	}
	t.Logf("machines before refine hook in Build already applied; count=%d (was logged %d)", len(p.Machines), before)
}

func TestSelectBestRecipeIronPlate(t *testing.T) {
	root, err := data.LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = data.LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatal(err)
	}
	r := SelectBestRecipe(root, "iron-plate", nil, true)
	if r == "" {
		t.Fatal("no recipe")
	}
}
