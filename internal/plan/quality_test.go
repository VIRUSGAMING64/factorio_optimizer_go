package plan

import "testing"

func TestWithQuality(t *testing.T) {
	if WithQuality("iron-plate", "normal") != "iron-plate" {
		t.Fatal("normal")
	}
	if WithQuality("iron-plate", "rare") != "iron-plate-rare" {
		t.Fatal("rare")
	}
	base, q := SplitQuality("iron-plate-epic")
	if base != "iron-plate" || q != "epic" {
		t.Fatalf("%s %s", base, q)
	}
}

func TestQualitySpread(t *testing.T) {
	s := QualitySpread(0, 0)
	if s[0] != 1 {
		t.Fatal("no bonus")
	}
	s = QualitySpread(0, 0.3)
	if s[0]+s[1] < 0.99 {
		t.Fatalf("%v", s)
	}
}

func TestQualityProductionCost(t *testing.T) {
	if QualityProductionCost(1, "normal") != 1 {
		t.Fatal("normal")
	}
	if QualityProductionCost(1, "rare") <= 1 {
		t.Fatal("rare should cost more")
	}
}
