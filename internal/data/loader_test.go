package data

import "testing"

func TestLoadRecipes(t *testing.T) {
	root, err := LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(root.Recipes) < 100 {
		t.Fatalf("recipes=%d", len(root.Recipes))
	}
	if _, ok := root.Recipes["iron-plate"]; !ok {
		// may be named differently; check copper-plate or electronic-circuit
		if _, ok2 := root.Recipes["electronic-circuit"]; !ok2 {
			t.Fatal("missing common recipe")
		}
	}
	if len(root.Items) < 50 {
		t.Fatalf("items=%d", len(root.Items))
	}
}

func TestFootprint(t *testing.T) {
	s := Footprint("assembling-machine-3", nil)
	if s.W != 3 || s.H != 3 {
		t.Fatalf("asm3 size %v", s)
	}
	s = Footprint("transport-belt", nil)
	if s.W != 1 || s.H != 1 {
		t.Fatalf("belt size %v", s)
	}
}

func TestInserterRate(t *testing.T) {
	r := InserterRate("fast-inserter", false, 0, 1)
	if r <= 0 {
		t.Fatal("rate should be positive")
	}
	r2 := InserterRate("fast-inserter", false, 7, 1)
	if r2 < r {
		t.Fatal("higher research should not lower rate")
	}
}

func TestBeltTier(t *testing.T) {
	bt := BeltTierLookup("express")
	if bt.Throughput != 45 {
		t.Fatalf("express throughput %v", bt.Throughput)
	}
	if TierOfEntity("fast-transport-belt") == nil {
		t.Fatal("tier of fast belt")
	}
}
