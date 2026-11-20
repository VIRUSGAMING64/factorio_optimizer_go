package plan

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

func ptr(f float64) *float64 { return &f }

func TestBalanceResidualsReducesSurplus(t *testing.T) {
	// Recipe that only produces a non-target intermediate that is over-produced
	// relative to its consumption.
	root := &data.RootData{
		Recipes: map[string]data.Recipe{
			"make-stone": {
				Name: "make-stone",
				Products: data.FlexibleProducts{
					{Name: "stone", Amount: 1, Probability: ptr(1)},
				},
				Ingredients: data.FlexibleIngredients{{Name: "raw", Amount: 1}},
				Energy:      1,
			},
		},
	}
	selected := map[string]float64{"make-stone": 10} // produces 10 stone, consumed 0
	targets := map[string]float64{"iron-plate": 5}   // stone is not a target

	adj := BalanceResiduals(root, selected, targets, 3)
	if adj == 0 {
		t.Fatal("expected at least one residual adjustment")
	}
	if selected["make-stone"] >= 10 {
		t.Fatalf("expected crafts reduced, got %v", selected["make-stone"])
	}
	// Must not go negative
	if selected["make-stone"] < 0 {
		t.Fatal("crafts went negative")
	}
}

func TestBalanceResidualsRespectsTargets(t *testing.T) {
	root := &data.RootData{
		Recipes: map[string]data.Recipe{
			"make-both": {
				Name: "make-both",
				Products: data.FlexibleProducts{
					{Name: "target-a", Amount: 1, Probability: ptr(1)},
					{Name: "target-b", Amount: 1, Probability: ptr(1)},
				},
				Energy: 1,
			},
		},
	}
	selected := map[string]float64{"make-both": 5}
	targets := map[string]float64{"target-a": 4, "target-b": 4}
	adj := BalanceResiduals(root, selected, targets, 2)
	if adj != 0 {
		t.Fatalf("should not reduce a dual-target recipe, adj=%d", adj)
	}
	if selected["make-both"] != 5 {
		t.Fatal("crafts changed unexpectedly")
	}
}
