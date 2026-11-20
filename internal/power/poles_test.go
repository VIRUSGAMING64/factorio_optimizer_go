package power

import "testing"

func TestPlaceGrid(t *testing.T) {
	occ := map[[2]int]bool{}
	poles := PlaceGrid(20, 20, func(x, y int) bool { return occ[[2]int{x, y}] })
	if len(poles) == 0 {
		t.Fatal("no poles")
	}
	// poles should not overlap
	seen := map[[2]int]bool{}
	for _, p := range poles {
		k := [2]int{p.X, p.Y}
		if seen[k] {
			t.Fatalf("duplicate pole at %v", k)
		}
		seen[k] = true
	}
}

func TestCover(t *testing.T) {
	p := PolePlacement{Name: "medium-electric-pole", X: 5, Y: 5}
	if !Cover(p, 5.5, 5.5) {
		t.Fatal("should cover centre")
	}
	if Cover(p, 50, 50) {
		t.Fatal("should not cover far")
	}
}

func TestBeaconAffects(t *testing.T) {
	if !BeaconAffects(0, 0, 3, 0, 3, 3) {
		t.Fatal("nearby machine should be affected")
	}
	if BeaconAffects(0, 0, 30, 30, 3, 3) {
		t.Fatal("far machine should not")
	}
}

func TestSpeedBonus(t *testing.T) {
	if SpeedBonus(0, 0.5) != 0 {
		t.Fatal("0 beacons")
	}
	b := SpeedBonus(4, 0.5)
	if b <= 0 {
		t.Fatal("positive bonus")
	}
}
