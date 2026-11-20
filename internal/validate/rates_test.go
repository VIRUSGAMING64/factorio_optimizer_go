package validate

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestCheckPlanRatesRedGreen(t *testing.T) {
	root, err := data.LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = data.LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatalf("%v", err)
	}
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
	p, err := plan.Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	rep := CheckPlanRates(p, cfg, root)
	if !rep.OK {
		t.Fatalf("rates not OK: %s", rep.String())
	}
}
