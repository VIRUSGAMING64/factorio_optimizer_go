package validate

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestCheckLayoutRatesSynthetic(t *testing.T) {
	ents := []route.Entity{
		{Name: "fast-transport-belt", Kind: route.KindBelt, X: 0, Y: 0, Dir: geom.East, Item: "iron-plate", NetID: 0},
		{Name: "fast-transport-belt", Kind: route.KindBelt, X: 1, Y: 0, Dir: geom.East, Item: "iron-plate", NetID: 0},
		{Name: "fast-transport-belt", Kind: route.KindBelt, X: 2, Y: 0, Dir: geom.East, Item: "iron-plate", NetID: 0},
	}
	p := &plan.Plan{
		Nets: []plan.Net{{ID: 0, Item: "iron-plate", Rate: 5}},
	}
	rep := CheckLayoutRates(p, ents, "fast")
	if len(rep.Messages) == 0 {
		t.Fatal("expected messages")
	}
	t.Log(rep.String())
}

func TestCheckLayoutRatesIntegration(t *testing.T) {
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
	pl, err := place.Place(p, 1.5, 5000, 1000, 0.995)
	if err != nil {
		t.Fatal(err)
	}
	rt := route.Route(p, pl, "fast")
	rep := CheckLayoutRates(p, rt.Entities, "fast")
	t.Log(rep.String())
}
