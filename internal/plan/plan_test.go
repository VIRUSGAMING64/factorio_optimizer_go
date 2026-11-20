package plan

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

func loadRoot(t *testing.T) *data.RootData {
	t.Helper()
	root, err := data.LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = data.LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatalf("load data: %v", err)
	}
	return root
}

func TestBuildRedGreen(t *testing.T) {
	root := loadRoot(t)
	cfg := &config.RunConfiguration{
		Machine:  "assembling-machine-3",
		BeltTier: "fast",
		Quality:  "normal",
		Targets: []config.ConfigTarget{
			{Name: "automation-science-pack", Amount: 0.5},
			{Name: "logistic-science-pack", Amount: 0.5},
		},
	}
	cfg.ApplyDefaults()
	p, err := Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(p.Machines) == 0 {
		t.Fatal("no machines")
	}
	if len(p.Nets) == 0 {
		t.Fatal("no nets")
	}
	// targets should be produced
	for _, item := range []string{"automation-science-pack", "logistic-science-pack"} {
		if p.Production[item] < 0.49 {
			t.Fatalf("production of %s = %v", item, p.Production[item])
		}
	}
}

func TestBuildIronPlate(t *testing.T) {
	root := loadRoot(t)
	cfg := &config.RunConfiguration{
		Machine:  "assembling-machine-3",
		BeltTier: "fast",
		Quality:  "normal",
		Targets:  []config.ConfigTarget{{Name: "iron-plate", Amount: 5}},
	}
	cfg.ApplyDefaults()
	p, err := Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if p.Production["iron-plate"] < 4.9 {
		t.Fatalf("iron-plate prod %v", p.Production["iron-plate"])
	}
}

func TestSelectBestRecipe(t *testing.T) {
	root := loadRoot(t)
	r := SelectBestRecipe(root, "iron-gear-wheel", nil, true)
	if r == "" {
		t.Fatal("no recipe for gear")
	}
}
