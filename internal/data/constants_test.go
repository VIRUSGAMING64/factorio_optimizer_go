package data

import "testing"

func TestBeltTierLookupDefault(t *testing.T) {
	bt := BeltTierLookup("nope")
	if bt.Key != "fast" {
		t.Fatalf("%s", bt.Key)
	}
}

func TestHandSizeAt(t *testing.T) {
	nb, b := HandSizeAt(0)
	if nb < 1 || b < 1 {
		t.Fatal(nb, b)
	}
	_, b7 := HandSizeAt(7)
	if b7 < b {
		t.Fatal("research should not lower")
	}
	nb, b = HandSizeAt(100) // clamp
	if nb < 1 {
		t.Fatal("clamp")
	}
}

func TestIsRawAndTakesRecipe(t *testing.T) {
	if !IsRaw("iron-ore") || IsRaw("iron-plate") {
		t.Fatal("raw")
	}
	if !TakesRecipe("assembling-machine-3") {
		t.Fatal("asm takes recipe")
	}
	if TakesRecipe("stone-furnace") {
		t.Fatal("furnace")
	}
}

func TestFluidsOf(t *testing.T) {
	c, _ := FluidsOf("chemical-plant")
	if len(c) == 0 {
		t.Fatal("chem plant fluids")
	}
}

func TestQualityLevelAndModule(t *testing.T) {
	if QualityLevel("legendary") != 4 {
		t.Fatal("legendary")
	}
	if ModuleBonusMultiplier(2) <= 1 {
		t.Fatal("bonus")
	}
}

func TestPoleAndInserterSpecs(t *testing.T) {
	if BasePoles["substation"].WireReach < 10 {
		t.Fatal("substation")
	}
	if BaseInserters["bulk-inserter"].Bulk != true {
		t.Fatal("bulk")
	}
}
