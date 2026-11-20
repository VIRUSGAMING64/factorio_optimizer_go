package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestRipUp(t *testing.T) {
	ents := []Entity{
		{Kind: KindMachine, X: 0, Y: 0, NetID: 0},
		{Kind: KindBelt, X: 1, Y: 0, NetID: 1, Item: "a"},
		{Kind: KindBelt, X: 2, Y: 0, NetID: 2, Item: "b"},
	}
	occ := []int{0, 2, 3, 0}
	out, n := RipUp(ents, occ, 4, 1, map[int]bool{1: true})
	if n != 1 {
		t.Fatalf("removed %d", n)
	}
	if len(out) != 2 {
		t.Fatalf("left %d", len(out))
	}
	_ = geom.North
}

func TestRerouteFailed(t *testing.T) {
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
	pl, err := place.Place(p, 1.5, 4000, 1000, 0.995)
	if err != nil {
		t.Fatal(err)
	}
	prev := Route(p, pl, "fast")
	rt := RerouteFailed(p, pl, "fast", prev, 2)
	if len(rt.Entities) == 0 {
		t.Fatal("empty")
	}
}
