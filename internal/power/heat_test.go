package power

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

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

func TestConnectHeatTree(t *testing.T) {
	// Two distant coverage pipes should become connected by intermediates.
	hn := HeatNetwork{Pipes: []HeatPipePlacement{
		{X: 0, Y: 0},
		{X: 6, Y: 3},
		{X: 3, Y: 9},
	}}
	connected := ConnectHeatTree(hn, nil)
	if len(connected.Pipes) <= len(hn.Pipes) {
		t.Fatalf("expected intermediates, got %d pipes from %d", len(connected.Pipes), len(hn.Pipes))
	}
	if !isConnected(connected) {
		t.Fatal("ConnectHeatTree result is not 4-connected")
	}
	// Original coverage points must still be present.
	for _, p := range hn.Pipes {
		found := false
		for _, q := range connected.Pipes {
			if q.X == p.X && q.Y == p.Y {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("original pipe (%d,%d) missing after connect", p.X, p.Y)
		}
	}
}

func TestConnectHeatTreeSingle(t *testing.T) {
	hn := HeatNetwork{Pipes: []HeatPipePlacement{{X: 2, Y: 2}}}
	out := ConnectHeatTree(hn, nil)
	if len(out.Pipes) != 1 {
		t.Fatalf("single pipe should stay single, got %d", len(out.Pipes))
	}
}

func isConnected(hn HeatNetwork) bool {
	if len(hn.Pipes) <= 1 {
		return true
	}
	idx := map[geom.Point]int{}
	for i, p := range hn.Pipes {
		idx[geom.Point{p.X, p.Y}] = i
	}
	vis := make([]bool, len(hn.Pipes))
	var dfs func(int)
	dfs = func(i int) {
		vis[i] = true
		p := hn.Pipes[i]
		for _, d := range []geom.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			k := geom.Point{p.X + d.X, p.Y + d.Y}
			if j, ok := idx[k]; ok && !vis[j] {
				dfs(j)
			}
		}
	}
	dfs(0)
	for _, v := range vis {
		if !v {
			return false
		}
	}
	return true
}
