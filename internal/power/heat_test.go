package power

import "testing"

func TestBuildHeatGrid(t *testing.T) {
	hn := BuildHeatGrid(10, 10, 3)
	if len(hn.Pipes) == 0 {
		t.Fatal("no pipes")
	}
	if !hn.Covers(1.5, 1.5, 2) {
		t.Fatal("should cover near lattice")
	}
}
