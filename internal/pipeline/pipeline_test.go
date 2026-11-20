package pipeline_test

import (
	"strings"
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/emit"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/validate"
)

func TestEndToEndRedGreen(t *testing.T) {
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
	rates := validate.CheckPlanRates(p, cfg, root)
	if !rates.OK {
		t.Fatalf("plan rates: %s", rates.String())
	}
	pl, err := place.Place(p, 1.5, 10000, 1000, 0.995)
	if err != nil {
		t.Fatal(err)
	}
	rt := route.Route(p, pl, "fast")
	rep := validate.Check(p, pl, rt)
	if !rep.OK {
		t.Fatalf("validate: %s", rep.String())
	}
	layout := validate.CheckLayout(rt.Entities)
	if !layout.OK {
		t.Fatalf("layout: %s", layout.String())
	}
	bp, err := emit.FromRoute(rt.Entities, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(bp, "0") {
		t.Fatal("blueprint prefix")
	}
	js, err := emit.DecodeToJSON(bp)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(js, "entities") {
		t.Fatal("no entities in json")
	}
		// count nets with belts
	covered := map[int]bool{}
	for _, e := range rt.Entities {
		if e.Kind == route.KindBelt || e.Kind == route.KindUndergroundIn {
			covered[e.NetID] = true
		}
	}
	itemNets := 0
	for _, n := range p.Nets {
		if !n.IsFluid {
			itemNets++
		}
	}
	if len(covered) == 0 {
		t.Fatal("no nets covered by belts")
	}
	t.Logf("machines=%d belts=%d ug=%d inserters=%d coverage=%d/%d bp=%d bytes",
		len(p.Machines), rt.Belts, rt.Undergrounds, rt.Inserters, len(covered), itemNets, len(bp))
}
