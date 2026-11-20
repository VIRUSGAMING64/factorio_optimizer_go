package validate

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestCheckEmptyPlan(t *testing.T) {
	p := &plan.Plan{}
	pl := place.Placement{W: 5, H: 5}
	rt := route.Result{}
	rep := Check(p, pl, rt)
	if rep.OK {
		t.Fatal("empty plan should not be OK")
	}
}

func TestMaxFlowRaiseAndFlowOn(t *testing.T) {
	mf := NewMaxFlow(3)
	id := mf.Edge(0, 1, 5)
	mf.Edge(1, 2, 5)
	mf.Raise(id, 5)
	f := mf.Run(0, 2)
	if f < 4.9 {
		t.Fatalf("%v", f)
	}
	_ = mf.FlowOn(id)
}

func TestBeltGraphUGLink(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindUndergroundIn, X: 0, Y: 0, Dir: 1, Item: "iron-plate"},  // East
		{Kind: route.KindUndergroundOut, X: 4, Y: 0, Dir: 1, Item: "iron-plate"},
	}
	g := BuildBeltGraph(ents)
	g.PropagateItems()
	if len(g.Nodes) != 2 {
		t.Fatal(len(g.Nodes))
	}
}

func TestIntegrationValidateRedGreen(t *testing.T) {
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
	pl, err := place.Place(p, 1.5, 6000, 1000, 0.995)
	if err != nil {
		t.Fatal(err)
	}
	rt := route.RouteMultiPass(p, pl, "fast", 2)
	rep := Check(p, pl, rt)
	rates := CheckPlanRates(p, cfg, root)
	layout := CheckLayout(rt.Entities)
	lr := CheckLayoutRates(p, rt.Entities, "fast")
	if !rates.OK {
		t.Fatal(rates.String())
	}
	t.Log(rep.String(), layout.String(), lr.String())
}
