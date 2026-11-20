package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestRouteMultiPass(t *testing.T) {
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
			{Name: "automation-science-pack", Amount: 0.5},
			{Name: "logistic-science-pack", Amount: 0.5},
		},
	}
	cfg.ApplyDefaults()
	p, err := plan.Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	pl, err := place.Place(p, 1.5, 8000, 1000, 0.995)
	if err != nil {
		t.Fatal(err)
	}
	rt := RouteMultiPass(p, pl, "fast", 2)
	if len(rt.Entities) == 0 {
		t.Fatal("no entities")
	}
	if rt.Belts == 0 {
		t.Fatal("no belts")
	}
}
