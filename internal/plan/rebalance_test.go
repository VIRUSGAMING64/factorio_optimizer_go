package plan

import "testing"

func TestRebalance(t *testing.T) {
	p := &Plan{
		Machines: []MachineInstance{
			{Recipe: "iron-plate", Crafts: 10, Products: []string{"iron-plate"}, Ingredients: []string{"iron-ore"}},
			{Recipe: "iron-gear-wheel", Crafts: 2, Products: []string{"iron-gear-wheel"}, Ingredients: []string{"iron-plate"}},
		},
		Production:  map[string]float64{"iron-plate": 10, "iron-gear-wheel": 2},
		Consumption: map[string]float64{"iron-plate": 2, "iron-ore": 10},
	}
	Rebalance(p, 0.05)
	if p.Production["iron-plate"] > 3 {
		t.Fatalf("expected trim, got %v", p.Production["iron-plate"])
	}
}
