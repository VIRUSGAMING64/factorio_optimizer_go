// Package route lays belts, undergrounds, inserters and poles between placed machines.
package route

import (
	"container/heap"
	"fmt"
	"math"
	"sort"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// EntityKind classifies a placed layout entity.
type EntityKind int

const (
	KindMachine EntityKind = iota
	KindBelt
	KindInserter
	KindUndergroundIn
	KindUndergroundOut
	KindPole
	KindPipe
	KindSplitter
	KindBeacon
	KindChest
	KindPipeUG
)

// Entity is one entity in the routed layout.
type Entity struct {
	Name   string
	Kind   EntityKind
	X, Y   int
	Dir    geom.Direction
	Recipe string
	Item   string
	NetID  int
	UGType string // "input" / "output" for undergrounds
}

// Result is the output of routing.
type Result struct {
	History map[geom.Point]float64 // congestion for place feedback
	Success      bool
	Entities     []Entity
	Belts        int
	Undergrounds int
	Inserters    int
	FailedNets   []int
	Message      string
}

// Router holds working state.
type Router struct {
	plan         *plan.Plan
	placement    place.Placement
	belt         data.BeltTier
	inserter     string
	longInserter string
	w, h         int
	occ          []int // 0 free, -1 machine, -2 pole, >0 belt/ug net+1
	machAt       []int
	maxUGGap     int
	hist         []float64 // congestion history per tile
	currentNet   int
	lanes        *LaneTracker
}

// Route runs the router.
func Route(p *plan.Plan, pl place.Placement, beltTier string) Result {
	tier := data.BeltTierLookup(beltTier)
	r := &Router{
		plan:         p,
		placement:    pl,
		belt:         tier,
		inserter:     "fast-inserter",
		longInserter: "long-handed-inserter",
		w:            pl.W,
		h:            pl.H,
		occ:          make([]int, pl.W*pl.H),
		machAt:       make([]int, pl.W*pl.H),
		maxUGGap:     tier.MaxUndergroundGap,
		hist:         make([]float64, pl.W*pl.H),
		lanes:        NewLaneTracker(tier),
	}
	for i := range r.machAt {
		r.machAt[i] = -1
	}
	for _, m := range pl.Machines {
		mi := p.Machines[m.ID]
		for dy := 0; dy < mi.H; dy++ {
			for dx := 0; dx < mi.W; dx++ {
				x, y := m.X+dx, m.Y+dy
				if r.inBounds(geom.Point{x, y}) {
					idx := r.idx(x, y)
					r.occ[idx] = -1
					r.machAt[idx] = m.ID
				}
			}
		}
	}

	var entities []Entity
	for _, m := range pl.Machines {
		mi := p.Machines[m.ID]
		entities = append(entities, Entity{
			Name: mi.Entity, Kind: KindMachine, X: m.X, Y: m.Y,
			Dir: geom.Direction(m.Rot), Recipe: mi.Recipe,
		})
	}

	belts, ugs, inserters := 0, 0, 0
	// Reserve pole grid before routing so belts do not fill every free tile
	poles := 0
	for y := 1; y < r.h; y += 5 {
		for x := 1; x < r.w; x += 5 {
			for _, dxy := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
				px, py := x+dxy[0], y+dxy[1]
				if !r.inBounds(geom.Point{px, py}) {
					continue
				}
				idx := r.idx(px, py)
				if r.occ[idx] == 0 {
					entities = append(entities, Entity{Name: "medium-electric-pole", Kind: KindPole, X: px, Y: py})
					r.occ[idx] = -2
					poles++
					break
				}
			}
		}
	}

	ringPorts := PlaceRingPorts(p, pl)
	entities = append(entities, RingPortEntities(ringPorts, r.belt.Belt)...)
	for _, rp := range ringPorts {
		if r.inBounds(rp.Tile) && r.occ[r.idx(rp.Tile.X, rp.Tile.Y)] == 0 {
			r.occ[r.idx(rp.Tile.X, rp.Tile.Y)] = rp.NetID + 1
		}
	}

	pf := NewPathFinder(r.w, r.h, r.maxUGGap)
	copy(pf.Occ, r.occ)
	pf.Hist = r.hist

	var failed []int
	placedIns := map[[2]int]bool{}

	// Machine position lookup
	posOf := map[int]place.PlacedMachine{}
	for _, m := range pl.Machines {
		posOf[m.ID] = m
	}

	for _, net := range p.Nets {
		if net.IsFluid {
			continue
		}
		var producers, consumers []int
		for i, mi := range p.Machines {
			for _, pr := range mi.Products {
				if pr == net.Item {
					producers = append(producers, i)
					break
				}
			}
			for _, ing := range mi.Ingredients {
				if ing == net.Item {
					consumers = append(consumers, i)
					break
				}
			}
		}

		free := func(t geom.Point) bool {
			return r.walkable(t)
		}

		var sources, targets []geom.Point
		var srcAccess, tgtAccess []AccessOption

		for _, id := range producers {
			m := posOf[id]
			rect := geom.NewRect(m.X, m.Y, p.Machines[id].W, p.Machines[id].H)
			opts := AccessOptions(id, rect, true, free, true)
			srcAccess = append(srcAccess, opts...)
			for _, o := range opts {
				sources = append(sources, o.BeltTile)
			}
		}
		for _, id := range consumers {
			m := posOf[id]
			rect := geom.NewRect(m.X, m.Y, p.Machines[id].W, p.Machines[id].H)
			opts := AccessOptions(id, rect, false, free, true)
			tgtAccess = append(tgtAccess, opts...)
			for _, o := range opts {
				targets = append(targets, o.BeltTile)
			}
		}
		for _, port := range p.Ports {
			if port.Item != net.Item {
				continue
			}
			tiles := edgePortTiles(pl, port)
			if port.IsInput {
				sources = append(sources, tiles...)
			} else {
				targets = append(targets, tiles...)
			}
		}

		if len(sources) == 0 || len(targets) == 0 {
			continue
		}
		r.currentNet = net.ID
		pf.DecayHistory()

		// Dedup targets by closest unique positions (limit work)
		targets = uniquePoints(targets)
		sources = uniquePoints(sources)

		connected := 0
		for _, tgt := range targets {
			bestPath := []geom.Point(nil)
			bestLen := math.MaxInt
			for _, src := range sources {
				path, _ := pf.FindPathLaneAStar(src, tgt, net.ID, LaneLeft)
				if path == nil {
					path = pf.FindPath(src, tgt, net.ID)
				}
				if path != nil && len(path) < bestLen {
					bestLen = len(path)
					bestPath = path
				}
			}
			if bestPath == nil {
				failed = append(failed, net.ID)
				continue
			}
			connected++
			// Compress long straight segments into undergrounds
			ents, b, u := r.pathToEntities(bestPath, net)
			entities = append(entities, ents...)
			belts += b
			ugs += u
			pf.CommitPath(bestPath, net.ID)
			pf.DetectDeadEnds(bestPath)
			for _, t := range bestPath {
				if r.inBounds(t) {
					r.hist[r.idx(t.X, t.Y)] = pf.Hist[r.idx(t.X, t.Y)]
					r.occ[r.idx(t.X, t.Y)] = pf.Occ[r.idx(t.X, t.Y)]
				}
			}
		}

		// Second pass: ensure each source reaches at least one target via existing or new path
		for _, src := range sources {
			bestPath := []geom.Point(nil)
			bestLen := math.MaxInt
			for _, tgt := range targets {
				path, _ := pf.FindPathLaneAStar(src, tgt, net.ID, LaneLeft)
				if path == nil {
					path = pf.FindPath(src, tgt, net.ID)
				}
				if path != nil && len(path) < bestLen {
					bestLen = len(path)
					bestPath = path
				}
			}
			if bestPath != nil {
				ents, b, u := r.pathToEntities(bestPath, net)
				entities = append(entities, ents...)
				belts += b
				ugs += u
				pf.CommitPath(bestPath, net.ID)
				pf.DetectDeadEnds(bestPath)
				for _, t := range bestPath {
					if r.inBounds(t) {
						r.hist[r.idx(t.X, t.Y)] = pf.Hist[r.idx(t.X, t.Y)]
						r.occ[r.idx(t.X, t.Y)] = pf.Occ[r.idx(t.X, t.Y)]
					}
				}
			}
		}

		// Place inserters from access options that land on our belts
		beltSet := map[geom.Point]bool{}
		for _, e := range entities {
			if e.Kind == KindBelt && e.NetID == net.ID {
				beltSet[geom.Point{e.X, e.Y}] = true
			}
		}
		for _, o := range append(srcAccess, tgtAccess...) {
			if !beltSet[o.BeltTile] {
				continue
			}
			key := [2]int{o.InsTile.X, o.InsTile.Y}
			if placedIns[key] {
				continue
			}
			if !r.inBounds(o.InsTile) {
				continue
			}
			name := r.inserter
			if o.Long {
				name = r.longInserter
			}
			entities = append(entities, Entity{
				Name: name, Kind: KindInserter,
				X: o.InsTile.X, Y: o.InsTile.Y,
				Dir: o.InsDir, Item: net.Item, NetID: net.ID,
			})
			placedIns[key] = true
			inserters++
			// Mark inserter tile occupied if free
			if r.inBounds(o.InsTile) {
				idx := r.idx(o.InsTile.X, o.InsTile.Y)
				if r.occ[idx] == 0 {
					r.occ[idx] = -3
				}
			}
		}
		_ = connected
	}

	// Splitters on junctions for each net
	for _, net := range p.Nets {
		if net.IsFluid {
			continue
		}
		sp := PlaceSplittersOnBranch(entities, r.belt, net)
		// Replace belt at splitter tile with splitter
		if len(sp) > 0 {
			remove := map[[2]int]bool{}
			for _, s := range sp {
				remove[[2]int{s.X, s.Y}] = true
			}
			filtered := entities[:0]
			for _, e := range entities {
				if e.Kind == KindBelt && remove[[2]int{e.X, e.Y}] {
					continue
				}
				filtered = append(filtered, e)
			}
			entities = append(filtered, sp...)
		}
	}

	// Fluid pipes
	pr := RoutePipes(p, pl, r.occ, r.w, r.h)
	entities = append(entities, pr.Entities...)
	if len(pr.Failed) > 0 {
		failed = append(failed, pr.Failed...)
	}

	spExtra, _ := PlaceSpoilageDrains(entities, r.occ, r.w, r.h)
	entities = append(entities, spExtra...)

	_ = pf.ScanDeadEnds()
	success := len(failed) == 0
	msg := fmt.Sprintf("%d belts, %d undergrounds, %d inserters, %d poles, %d pipes", belts, ugs, inserters, poles, len(pr.Entities))
	if !success {
		msg += fmt.Sprintf("; %d net failures", len(uniqueInts(failed)))
	}
	return Result{
		Success: success || belts > 0, Entities: entities,
		Belts: belts, Undergrounds: ugs, Inserters: inserters,
		FailedNets: uniqueInts(failed), Message: msg,
		History: pf.HistoryMap(),
	}
}

// pathToEntities places belts and optionally undergrounds along a path.
func (r *Router) pathToEntities(path []geom.Point, net plan.Net) (ents []Entity, belts, ugs int) {
	if len(path) == 0 {
		return nil, 0, 0
	}
	i := 0
	for i < len(path) {
		// Try to start an underground run if straight segment is long enough
		if i+3 < len(path) {
			dir := dirBetween(path[i], path[i+1])
			run := 1
			for j := i + 1; j < len(path); j++ {
				if j+1 < len(path) && dirBetween(path[j], path[j+1]) == dir {
					run++
				} else {
					break
				}
			}
			// Underground: entrance, gap up to maxUGGap, exit. Total span = gap+1 tiles between.
			if run >= 4 {
				gap := run - 1
				if gap > r.maxUGGap {
					gap = r.maxUGGap
				}
				if gap >= 2 {
					entrance := path[i]
					exit := path[i+gap]
					if r.inBounds(entrance) && r.inBounds(exit) {
						ei := r.idx(entrance.X, entrance.Y)
						xi := r.idx(exit.X, exit.Y)
						if r.occ[ei] == 0 && r.occ[xi] == 0 {
							r.occ[ei] = net.ID + 1
							r.occ[xi] = net.ID + 1
							ents = append(ents, Entity{
								Name: r.belt.Underground, Kind: KindUndergroundIn,
								X: entrance.X, Y: entrance.Y, Dir: dir, Item: net.Item, NetID: net.ID, UGType: "input",
							})
							ents = append(ents, Entity{
								Name: r.belt.Underground, Kind: KindUndergroundOut,
								X: exit.X, Y: exit.Y, Dir: dir, Item: net.Item, NetID: net.ID, UGType: "output",
							})
							ugs += 2
							i += gap + 1
							continue
						}
					}
				}
			}
		}
		// Regular belt
		t := path[i]
		if r.inBounds(t) {
			idx := r.idx(t.X, t.Y)
			rateShare := math.Abs(net.Rate)
			if rateShare > r.lanes.cap {
				rateShare = r.lanes.cap * 0.5 // clamp extreme rates
			}
			if r.occ[idx] == 0 {
				dir := geom.East
				if i+1 < len(path) {
					dir = dirBetween(path[i], path[i+1])
				} else if i > 0 {
					dir = dirBetween(path[i-1], path[i])
				}
				r.occ[idx] = net.ID + 1
				r.lanes.Add(idx, rateShare)
				ents = append(ents, Entity{
					Name: r.belt.Belt, Kind: KindBelt,
					X: t.X, Y: t.Y, Dir: dir, Item: net.Item, NetID: net.ID,
				})
				belts++
			} else if r.occ[idx] == net.ID+1 && r.lanes.CanShare(idx, rateShare*0.5) {
				// Same-net reuse: record additional load
				r.lanes.Add(idx, rateShare*0.5)
			}
		}
		i++
	}
	return ents, belts, ugs
}

func (r *Router) idx(x, y int) int { return y*r.w + x }

func (r *Router) inBounds(p geom.Point) bool {
	return p.X >= 0 && p.X < r.w && p.Y >= 0 && p.Y < r.h
}

func (r *Router) walkable(p geom.Point) bool {
	if !r.inBounds(p) {
		return false
	}
	v := r.occ[r.idx(p.X, p.Y)]
	// free, any existing belt (path can follow corridors), not machines/poles/inserters
	return v == 0 || v > 0
}

func (r *Router) astar(src, dst geom.Point) []geom.Point {
	src = r.nearestFree(src)
	dst = r.nearestFree(dst)
	if src.X < 0 || dst.X < 0 {
		return nil
	}
	key := func(p geom.Point) int { return p.Y*r.w + p.X }
	open := &minHeap{}
	heap.Init(open)
	heap.Push(open, aNode{p: src, g: 0, f: float64(src.Manhattan(dst))})
	came := map[int]geom.Point{}
	gscore := map[int]float64{key(src): 0}
	closed := map[int]bool{}

	for open.Len() > 0 {
		cur := heap.Pop(open).(aNode)
		ck := key(cur.p)
		if closed[ck] {
			continue
		}
		closed[ck] = true
		if cur.p == dst {
			path := []geom.Point{cur.p}
			for {
				prev, ok := came[key(path[len(path)-1])]
				if !ok {
					break
				}
				path = append(path, prev)
			}
			for a, b := 0, len(path)-1; a < b; a, b = a+1, b-1 {
				path[a], path[b] = path[b], path[a]
			}
			return path
		}
		for _, d := range geom.AllDirections {
			np := cur.p.Add(d.Delta())
			if !r.walkable(np) {
				continue
			}
			nk := key(np)
			if closed[nk] {
				continue
			}
			step := 1.0 + 0.25*r.hist[r.idx(np.X, np.Y)]
			// Prefer sharing belts of the same net when residual lane capacity remains
			idx := r.idx(np.X, np.Y)
			v := r.occ[idx]
			if v > 0 && r.currentNet >= 0 && v == r.currentNet+1 {
				if r.lanes.CanShare(idx, 5) {
					step *= 0.40 // strong preference for reuse when capacity allows
				} else {
					step += 3.0 // capacity exhausted – treat as congested
				}
			}
			if v > 0 && r.currentNet >= 0 && v != r.currentNet+1 {
				step += 5 // prefer not to route through foreign nets
			}
			ng := cur.g + step
			if old, ok := gscore[nk]; ok && ng >= old {
				continue
			}
			gscore[nk] = ng
			came[nk] = cur.p
			heap.Push(open, aNode{p: np, g: ng, f: ng + float64(np.Manhattan(dst))})
		}
	}
	return nil
}

func (r *Router) nearestFree(p geom.Point) geom.Point {
	if r.walkable(p) {
		return p
	}
	best := geom.Point{-1, -1}
	for rad := 1; rad <= 8; rad++ {
		for dy := -rad; dy <= rad; dy++ {
			for dx := -rad; dx <= rad; dx++ {
				if abs(dx)+abs(dy) != rad {
					continue
				}
				q := geom.Point{p.X + dx, p.Y + dy}
				if r.walkable(q) {
					return q
				}
			}
		}
	}
	return best
}

func edgePortTiles(pl place.Placement, port plan.Port) []geom.Point {
	var out []geom.Point
	switch port.Side {
	case "west":
		for y := 0; y < pl.H; y++ {
			out = append(out, geom.Point{0, y})
		}
	case "east":
		for y := 0; y < pl.H; y++ {
			out = append(out, geom.Point{pl.W - 1, y})
		}
	case "north":
		for x := 0; x < pl.W; x++ {
			out = append(out, geom.Point{x, 0})
		}
	case "south":
		for x := 0; x < pl.W; x++ {
			out = append(out, geom.Point{x, pl.H - 1})
		}
	}
	return out
}

func dirBetween(a, b geom.Point) geom.Direction {
	dx, dy := b.X-a.X, b.Y-a.Y
	if abs(dx) >= abs(dy) {
		if dx >= 0 {
			return geom.East
		}
		return geom.West
	}
	if dy >= 0 {
		return geom.South
	}
	return geom.North
}

func uniquePoints(pts []geom.Point) []geom.Point {
	seen := map[geom.Point]bool{}
	var out []geom.Point
	for _, p := range pts {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

func uniqueInts(a []int) []int {
	m := map[int]bool{}
	var out []int
	for _, v := range a {
		if !m[v] {
			m[v] = true
			out = append(out, v)
		}
	}
	sort.Ints(out)
	return out
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
