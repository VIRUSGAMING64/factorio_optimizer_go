package plan

import "testing"

func TestBeltThroughputAndCount(t *testing.T) {
	th := BeltThroughput("express", 1)
	if th != 45 {
		t.Fatalf("%v", th)
	}
	n := BeltCountForRate(90, "express", 1)
	if n != 2 {
		t.Fatalf("%d", n)
	}
}

func TestInserterCount(t *testing.T) {
	// high rate needs multiple
	n := InserterCount(20, "fast-inserter", false, struct {
		InserterCapacity int `json:"inserter_capacity,omitempty"`
	}{0}, 1)
	if n < 1 {
		t.Fatal(n)
	}
}

func TestCraftsAndOutput(t *testing.T) {
	cps := CraftsPerSecond(0.5, 1.25, 0)
	if cps <= 0 {
		t.Fatal(cps)
	}
	if EffectiveOutput(2, 0.1) <= 2 {
		t.Fatal("prod bonus")
	}
}
