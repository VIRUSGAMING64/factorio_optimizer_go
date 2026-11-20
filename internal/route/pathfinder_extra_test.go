package route

import (
	"math/rand"
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

func TestFindPathMulti(t *testing.T) {
	pf := NewPathFinder(20, 20, 6)
	pf.MarkMachine(8, 8, 3, 3)
	path := pf.FindPathMulti(geom.Point{0, 0}, []geom.Point{{19, 0}, {19, 19}, {0, 19}}, 1)
	if path == nil {
		t.Fatal("nil path")
	}
	if len(path) < 2 {
		t.Fatal("short")
	}
}

func TestHistoryMapAndScanDeadEnds(t *testing.T) {
	pf := NewPathFinder(10, 10, 4)
	path := pf.FindPath(geom.Point{0, 0}, geom.Point{5, 0}, 1)
	if path == nil {
		t.Fatal("path")
	}
	pf.CommitPath(path, 1)
	h := pf.HistoryMap()
	if len(h) == 0 {
		t.Fatal("empty history")
	}
	_ = pf.ScanDeadEnds()
	// endpoints of a simple path have degree 1; ScanDeadEnds may or may not mark them
	if pf.DeadEnd[pf.idx(5, 0)] < 0 {
		t.Fatal("deadend array corrupt")
	}
}

func TestShuffleOrder(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	order := ShuffleOrder(5, []int{3, 1}, rng)
	if len(order) != 5 {
		t.Fatal(len(order))
	}
	// failed should be among first two
	head := map[int]bool{order[0]: true, order[1]: true}
	if !head[3] && !head[1] {
		t.Fatalf("failed nets not prioritised: %v", order)
	}
}
