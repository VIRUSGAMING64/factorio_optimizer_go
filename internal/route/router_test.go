package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestRouteRedGreen(t *testing.T) {
	root, err := data.LoadRootData("../../data/recipes.json")
	if err != nil {
		root, err = data.LoadRootData("data/recipes.json")
	}
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.RunConfiguration{
		Machine: "assembling-machine-3", BeltTier: "fast", Quality: "normal",
		Targets: []config.ConfigTarget{
			{Name: "automation-science-pack", Amount: 0.5},
			{Name: "logistic-science-pack", Amount: 0.5},
		},
	}
	cfg.ApplyDefaults()
	p, err := plan.Build(root, cfg, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	pl, err := place.Place(p, 1.5, 5000, 1000, 0.995)
	if err != nil {
		t.Fatal(err)
	}
	rt := Route(p, pl, "fast")
	if len(rt.Entities) == 0 {
		t.Fatal("no entities")
	}
	if rt.Belts == 0 {
		t.Fatal("expected some belts")
	}
	machines := 0
	for _, e := range rt.Entities {
		if e.Kind == KindMachine {
			machines++
		}
	}
	if machines != len(p.Machines) {
		t.Fatalf("machines in route %d want %d", machines, len(p.Machines))
	}
}

func TestAStarPath(t *testing.T) {
	r := &Router{w: 10, h: 10, occ: make([]int, 100), machAt: make([]int, 100), hist: make([]float64, 100)}
	for i := range r.machAt {
		r.machAt[i] = -1
	}
	r.occ[4*10+4] = -1
	path := r.astar(geom.Point{0, 0}, geom.Point{9, 9})
	if path == nil {
		t.Fatal("expected path around obstacle")
	}
	if len(path) < 10 {
		t.Fatalf("path too short %d", len(path))
	}
}

func TestPathToEntitiesWithLanesPreference(t *testing.T) {
	// Minimal router just to exercise lane recording
	tier := data.BeltTier{Key: "fast", Throughput: 30, MaxUndergroundGap: 9, Belt: "fast-transport-belt", Underground: "fast-underground-belt"}
	r := &Router{
		w: 10, h: 10,
		occ:   make([]int, 100),
		hist:  make([]float64, 100),
		lanes: NewLaneTracker(tier),
		belt:  tier,
	}
	path := []geom.Point{{0, 0}, {1, 0}, {2, 0}}
	lanes := []Lane{LaneRight, LaneRight, LaneRight}
	net := plan.Net{ID: 1, Rate: 5, Item: "iron-plate"}
	ents, belts, ugs := r.pathToEntitiesWithLanes(path, lanes, net)
	if belts != 3 || ugs != 0 || len(ents) != 3 {
		t.Fatalf("expected 3 belts, got belts=%d ugs=%d ents=%d", belts, ugs, len(ents))
	}
	idx := r.idx(0, 0)
	if r.lanes.UsedLane(idx, LaneRight) < 4.9 {
		t.Fatalf("expected load on preferred right lane, got left=%v right=%v",
			r.lanes.UsedLane(idx, LaneLeft), r.lanes.UsedLane(idx, LaneRight))
	}
}
