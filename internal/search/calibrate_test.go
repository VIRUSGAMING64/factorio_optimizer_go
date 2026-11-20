package search

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestCalibrate(t *testing.T) {
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
	res, err := Calibrate(p, config.Defaults(), "fast", 2)
	if err != nil {
		t.Fatal(err)
	}
	if res.Samples < 1 {
		t.Fatal("samples")
	}
	t.Logf("avg score=%.1f area=%.0f belts=%.0f fail=%.1f", res.AvgScore, res.AvgArea, res.AvgBelts, res.AvgFailures)
}
