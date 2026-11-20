package search

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestTrialWithFeedbackBlue(t *testing.T) {
	root, err := data.LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = data.LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.RunConfiguration{
		Machine: "assembling-machine-3", BeltTier: "fast", Quality: "normal",
		Targets: []config.ConfigTarget{{Name: "processing-unit", Amount: 0.25}},
	}
	cfg.ApplyDefaults()
	p, err := plan.Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Machines) < 3 {
		t.Fatalf("expected multi-step plan, got %d machines", len(p.Machines))
	}
	tr, err := TrialWithFeedback(p, Options{
		Attempts: 1, SAIters: 8000, AreaFactor: 1.5, BeltTier: "fast", Tuning: config.Defaults(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if tr.Route.Belts < 1 {
		t.Fatal("no belts")
	}
	t.Logf("machines=%d belts=%d fail=%d score=%.1f",
		len(p.Machines), tr.Route.Belts, len(tr.Route.FailedNets), tr.Score)
}
