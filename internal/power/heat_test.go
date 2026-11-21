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

func TestBuildHeatCoverage(t *testing.T) {
	machines := []MachineCenter{
		{X: 2.5, Y: 2.5},
		{X: 7.5, Y: 7.5},
		{X: 2.5, Y: 7.5},
	}
	occupied := func(x, y int) bool { return false }
	hn := BuildHeatCoverage(12, 12, 3, 4, machines, occupied)
	if len(hn.Pipes) == 0 {
		t.Fatal("expected some pipes")
	}
	// Should be fewer pipes than a full grid for the same step.
	full := BuildHeatGrid(12, 12, 3)
	if len(hn.Pipes) >= len(full.Pipes) {
		t.Fatalf("coverage placer should be sparser than full grid: got %d vs %d", len(hn.Pipes), len(full.Pipes))
	}
	for _, m := range machines {
		if !hn.Covers(m.X, m.Y, 4) {
			t.Fatalf("machine at %.1f,%.1f not covered", m.X, m.Y)
		}
	}
}

func TestBuildHeatCoverageFallback(t *testing.T) {
	// No free candidates → falls back to grid.
	occupied := func(x, y int) bool { return true }
	hn := BuildHeatCoverage(8, 8, 2, 3, []MachineCenter{{X: 3, Y: 3}}, occupied)
	if len(hn.Pipes) == 0 {
		t.Fatal("fallback grid should still produce pipes")
	}
}
