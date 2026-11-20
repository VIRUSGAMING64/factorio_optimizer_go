package power

import "testing"

func TestPlaceBeaconGridPlacesSome(t *testing.T) {
	machines := []MachineRect{
		{X: 5, Y: 5, W: 3, H: 3},
		{X: 12, Y: 8, W: 3, H: 3},
	}
	occupied := func(x, y int) bool { return false }
	bps := PlaceBeaconGrid(20, 20, machines, occupied, nil)
	if len(bps) == 0 {
		t.Fatal("expected at least one beacon")
	}
	for _, b := range bps {
		if b.X < 0 || b.Y < 0 || len(b.Modules) == 0 {
			t.Errorf("unexpected placement %+v", b)
		}
	}
}

func TestSpeedBonusPositive(t *testing.T) {
	if SpeedBonus(0, 0.5) != 0 {
		t.Error("zero beacons should give 0")
	}
	if SpeedBonus(4, 0.5) <= 0 {
		t.Error("positive beacons should give positive bonus")
	}
}
