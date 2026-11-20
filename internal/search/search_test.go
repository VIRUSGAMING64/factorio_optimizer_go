package search

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestRunTrials(t *testing.T) {
	root, err := data.LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = data.LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.RunConfiguration{
		Machine: "assembling-machine-3", BeltTier: "fast", Quality: "normal",
		Targets: []config.ConfigTarget{{Name: "iron-gear-wheel", Amount: 1}},
	}
	cfg.ApplyDefaults()
	p, err := plan.Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	trial, err := RunTrials(p, Options{Attempts: 2, SAIters: 3000, AreaFactor: 1.5, BeltTier: "fast", Tuning: config.Defaults()})
	if err != nil {
		t.Fatal(err)
	}
	if len(trial.Route.Entities) == 0 {
		t.Fatal("no entities")
	}
}
