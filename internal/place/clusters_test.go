package place

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestDIClusters(t *testing.T) {
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
	clusters := DIClusters(p, false)
	// May or may not find clusters depending on plan shape; should not panic
	t.Logf("clusters=%d", len(clusters))
	seen := map[int]bool{}
	for _, c := range clusters {
		if len(c.Members) != len(c.DX) || len(c.Members) != len(c.DY) {
			t.Fatalf("mismatched offsets in %s", c.Kind)
		}
		for _, m := range c.Members {
			if seen[m] {
				t.Fatalf("machine %d in multiple clusters", m)
			}
			seen[m] = true
		}
	}
}
