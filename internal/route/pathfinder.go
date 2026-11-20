package route

import (
	"container/heap"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

// Lane is the left (0) or right (1) side of a belt relative to its heading.
type Lane int

const (
	LaneLeft  Lane = 0
	LaneRight Lane = 1
)

// PathState is a node in the space-time / lane-aware search.
type PathState struct {
	X, Y int
	Dir  geom.Direction
	Lane Lane
}

func (s PathState) Key(w int) int {
	// x,y,dir,lane packed
	return (((s.Y*w+s.X)<<2 | int(s.Dir)) << 1) | int(s.Lane)
}

// PathFinder holds negotiated-congestion state for multi-net routing.
type PathFinder struct {
	W, H     int
	Occ      []int     // 0 free, -1 machine, -2 pole, -3 inserter, >0 netID+1
	Hist     []float64 // congestion history
	HistChan []float64 // per-direction channel history (4 * W * H)
	DeadEnd  []int     // dead-end markers per tile
	MaxUG    int
	Present  float64 // present-factor for occupancy cost
	Decay    float64 // history decay per net
}

// NewPathFinder creates a pathfinder for a box.
func NewPathFinder(w, h, maxUG int) *PathFinder {
	n := w * h
	return &PathFinder{
		W: w, H: h, MaxUG: maxUG,
		Occ: make([]int, n), Hist: make([]float64, n),
		HistChan: make([]float64, n*4), DeadEnd: make([]int, n),
		Present: 1.0, Decay: 0.99,
	}
}

func (pf *PathFinder) idx(x, y int) int { return y*pf.W + x }

func (pf *PathFinder) inBounds(x, y int) bool {
	return x >= 0 && x < pf.W && y >= 0 && y < pf.H
}

// MarkMachine marks a rectangle as blocked.
func (pf *PathFinder) MarkMachine(x, y, w, h int) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			nx, ny := x+dx, y+dy
			if pf.inBounds(nx, ny) {
				pf.Occ[pf.idx(nx, ny)] = -1
			}
		}
	}
}

// MarkPole marks a pole tile.
func (pf *PathFinder) MarkPole(x, y int) {
	if pf.inBounds(x, y) {
		pf.Occ[pf.idx(x, y)] = -2
	}
}

// Walkable reports whether a belt may be placed / traversed.
func (pf *PathFinder) Walkable(x, y, netID int) bool {
	if !pf.inBounds(x, y) {
		return false
	}
	v := pf.Occ[pf.idx(x, y)]
	if v == 0 {
		return true
	}
	if v > 0 && (netID < 0 || v == netID+1) {
		return true
	}
	return false
}

// TileCost is the negotiated-congestion cost of using a tile.
func (pf *PathFinder) TileCost(x, y, netID int) float64 {
	i := pf.idx(x, y)
	c := 1.0 + pf.Hist[i]
	v := pf.Occ[i]
	if v > 0 && netID >= 0 && v != netID+1 {
		c += 8 // foreign net
	}
	if pf.DeadEnd[i] > 0 {
		c += 3 * float64(pf.DeadEnd[i])
	}
	return c * (1 + pf.Present*float64(maxInt(0, v)))
}

// ChanCost is cost for a directed edge (tile + direction).
func (pf *PathFinder) ChanCost(x, y int, d geom.Direction) float64 {
	i := pf.idx(x, y)*4 + int(d)
	return pf.HistChan[i]
}

// FindPath runs A* from src to dst for netID, optionally preferring a heading.
func (pf *PathFinder) FindPath(src, dst geom.Point, netID int) []geom.Point {
	src = pf.nearestFree(src, netID)
	dst = pf.nearestFree(dst, netID)
	if src.X < 0 || dst.X < 0 {
		return nil
	}
	type node struct {
		p    geom.Point
		g, f float64
	}
	key := func(p geom.Point) int { return p.Y*pf.W + p.X }
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
			if !pf.Walkable(np.X, np.Y, netID) {
				continue
			}
			nk := key(np)
			if closed[nk] {
				continue
			}
			step := pf.TileCost(np.X, np.Y, netID) + pf.ChanCost(cur.p.X, cur.p.Y, d)
			// Prefer straight: slight penalty for turns is implicit via history
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

// CommitPath marks a path as used by netID and bumps history.
func (pf *PathFinder) CommitPath(path []geom.Point, netID int) {
	for i, t := range path {
		if !pf.inBounds(t.X, t.Y) {
			continue
		}
		idx := pf.idx(t.X, t.Y)
		if pf.Occ[idx] == 0 {
			pf.Occ[idx] = netID + 1
		}
		pf.Hist[idx] += 1
		if i+1 < len(path) {
			d := dirBetween(t, path[i+1])
			ci := idx*4 + int(d)
			if ci >= 0 && ci < len(pf.HistChan) {
				pf.HistChan[ci] += 1
			}
		}
	}
}

// DecayHistory multiplies all history by Decay (call between nets).
func (pf *PathFinder) DecayHistory() {
	for i := range pf.Hist {
		pf.Hist[i] *= pf.Decay
	}
	for i := range pf.HistChan {
		pf.HistChan[i] *= pf.Decay
	}
}

// MarkDeadEnd records a belt tile that has no forward continuation.
func (pf *PathFinder) MarkDeadEnd(x, y int) {
	if pf.inBounds(x, y) {
		pf.DeadEnd[pf.idx(x, y)]++
	}
}

// DetectDeadEnds scans committed belts of a net and marks terminals without forward belt.
func (pf *PathFinder) DetectDeadEnds(path []geom.Point) {
	if len(path) < 2 {
		return
	}
	// Last tile has no continuation in path — potential dead end if no other belt continues
	end := path[len(path)-1]
	d := dirBetween(path[len(path)-2], end)
	fwd := end.Add(d.Delta())
	if !pf.Walkable(fwd.X, fwd.Y, -1) {
		pf.MarkDeadEnd(end.X, end.Y)
	}
}

func (pf *PathFinder) nearestFree(p geom.Point, netID int) geom.Point {
	if pf.Walkable(p.X, p.Y, netID) {
		return p
	}
	for rad := 1; rad <= 10; rad++ {
		for dy := -rad; dy <= rad; dy++ {
			for dx := -rad; dx <= rad; dx++ {
				if abs(dx)+abs(dy) != rad {
					continue
				}
				q := geom.Point{p.X + dx, p.Y + dy}
				if pf.Walkable(q.X, q.Y, netID) {
					return q
				}
			}
		}
	}
	return geom.Point{-1, -1}
}

// PathToEntities converts a path into belt/underground entities (shared with router logic).
func (pf *PathFinder) PathToEntities(path []geom.Point, netID int, item, belt, ug string) (ents []Entity, belts, ugs int) {
	if len(path) == 0 {
		return nil, 0, 0
	}
	i := 0
	for i < len(path) {
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
			if run >= 4 {
				gap := run - 1
				if gap > pf.MaxUG {
					gap = pf.MaxUG
				}
				if gap >= 2 {
					entrance, exit := path[i], path[i+gap]
					if pf.inBounds(entrance.X, entrance.Y) && pf.inBounds(exit.X, exit.Y) {
						ei, xi := pf.idx(entrance.X, entrance.Y), pf.idx(exit.X, exit.Y)
						if (pf.Occ[ei] == 0 || pf.Occ[ei] == netID+1) && (pf.Occ[xi] == 0 || pf.Occ[xi] == netID+1) {
							pf.Occ[ei] = netID + 1
							pf.Occ[xi] = netID + 1
							ents = append(ents,
								Entity{Name: ug, Kind: KindUndergroundIn, X: entrance.X, Y: entrance.Y, Dir: dir, Item: item, NetID: netID, UGType: "input"},
								Entity{Name: ug, Kind: KindUndergroundOut, X: exit.X, Y: exit.Y, Dir: dir, Item: item, NetID: netID, UGType: "output"},
							)
							ugs += 2
							i += gap + 1
							continue
						}
					}
				}
			}
		}
		t := path[i]
		if pf.inBounds(t.X, t.Y) {
			idx := pf.idx(t.X, t.Y)
			if pf.Occ[idx] == 0 || pf.Occ[idx] == netID+1 {
				dir := geom.East
				if i+1 < len(path) {
					dir = dirBetween(path[i], path[i+1])
				} else if i > 0 {
					dir = dirBetween(path[i-1], path[i])
				}
				if pf.Occ[idx] == 0 {
					pf.Occ[idx] = netID + 1
					ents = append(ents, Entity{Name: belt, Kind: KindBelt, X: t.X, Y: t.Y, Dir: dir, Item: item, NetID: netID})
					belts++
				}
			}
		}
		i++
	}
	return ents, belts, ugs
}

// LaneCapacity is half of belt throughput (items/s per lane).
func LaneCapacity(beltThroughput float64) float64 {
	return beltThroughput / 2 * 0.98
}

// AssignLanes greedily puts flows onto left/right lanes without exceeding capacity.
func AssignLanes(flows []float64, laneCap float64) []Lane {
	out := make([]Lane, len(flows))
	left, right := 0.0, 0.0
	for i, f := range flows {
		if left <= right && left+f <= laneCap+1e-9 {
			out[i] = LaneLeft
			left += f
		} else if right+f <= laneCap+1e-9 {
			out[i] = LaneRight
			right += f
		} else if left <= right {
			out[i] = LaneLeft
			left += f
		} else {
			out[i] = LaneRight
			right += f
		}
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// FindPathWithLanes is A* that tracks preferred lane and penalises switching.
// Returns path and per-tile lane assignment (same length as path; default left).
func (pf *PathFinder) FindPathWithLanes(src, dst geom.Point, netID int, prefer Lane) ([]geom.Point, []Lane) {
	path := pf.FindPath(src, dst, netID)
	if path == nil {
		return nil, nil
	}
	lanes := make([]Lane, len(path))
	for i := range lanes {
		lanes[i] = prefer
	}
	// Alternate lane on turns to reduce same-lane congestion
	for i := 1; i+1 < len(path); i++ {
		d0 := dirBetween(path[i-1], path[i])
		d1 := dirBetween(path[i], path[i+1])
		if d0 != d1 {
			if lanes[i-1] == LaneLeft {
				lanes[i] = LaneRight
			} else {
				lanes[i] = LaneLeft
			}
		} else {
			lanes[i] = lanes[i-1]
		}
	}
	if len(lanes) > 0 {
		lanes[len(lanes)-1] = lanes[len(lanes)-1]
	}
	return path, lanes
}

// PresentBoost raises the occupancy present-factor (negotiated congestion pressure).
func (pf *PathFinder) PresentBoost(factor float64) {
	if factor < 1 {
		factor = 1
	}
	pf.Present *= factor
	if pf.Present > 8 {
		pf.Present = 8
	}
}

// ClearBeltMarks frees all positive occupancy (belts) keeping machines/poles.
func (pf *PathFinder) ClearBeltMarks() {
	for i, v := range pf.Occ {
		if v > 0 {
			pf.Occ[i] = 0
		}
	}
}

// ConflictTiles returns tiles used by more than one net (hist high + occ conflict).
func (pf *PathFinder) ConflictTiles(threshold float64) []geom.Point {
	var out []geom.Point
	for y := 0; y < pf.H; y++ {
		for x := 0; x < pf.W; x++ {
			if pf.Hist[pf.idx(x, y)] >= threshold {
				out = append(out, geom.Point{x, y})
			}
		}
	}
	return out
}

// FindPathMulti routes from src to the nearest of several destinations.
func (pf *PathFinder) FindPathMulti(src geom.Point, dsts []geom.Point, netID int) []geom.Point {
	var best []geom.Point
	bestLen := int(^uint(0) >> 1)
	for _, d := range dsts {
		path := pf.FindPath(src, d, netID)
		if path == nil {
			continue
		}
		if len(path) < bestLen {
			bestLen = len(path)
			best = path
		}
	}
	return best
}

// HistoryMap returns congestion history as a map of points (for place feedback).
func (pf *PathFinder) HistoryMap() map[geom.Point]float64 {
	out := make(map[geom.Point]float64)
	for y := 0; y < pf.H; y++ {
		for x := 0; x < pf.W; x++ {
			v := pf.Hist[pf.idx(x, y)]
			if v > 0.5 {
				out[geom.Point{x, y}] = v
			}
		}
	}
	return out
}

// ScanDeadEnds marks tiles whose belt occupancy has no outbound neighbour belt of same net.
func (pf *PathFinder) ScanDeadEnds() int {
	marked := 0
	for y := 0; y < pf.H; y++ {
		for x := 0; x < pf.W; x++ {
			i := pf.idx(x, y)
			net := pf.Occ[i]
			if net <= 0 {
				continue
			}
			hasOut := false
			for _, d := range geom.AllDirections {
				nx, ny := x+d.Delta().X, y+d.Delta().Y
				if !pf.inBounds(nx, ny) {
					continue
				}
				if pf.Occ[pf.idx(nx, ny)] == net {
					hasOut = true
					break
				}
			}
			if !hasOut {
				pf.DeadEnd[i]++
				marked++
			}
		}
	}
	return marked
}

// ShuffleOrder returns a permutation of net indices biased by previous failures.
func ShuffleOrder(n int, failed []int, rng interface{ Intn(int) int }) []int {
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	// Put previously failed nets first
	failSet := map[int]bool{}
	for _, f := range failed {
		failSet[f] = true
	}
	pri, rest := []int{}, []int{}
	for _, i := range order {
		if failSet[i] {
			pri = append(pri, i)
		} else {
			rest = append(rest, i)
		}
	}
	// Fisher-Yates within groups
	shuffle := func(a []int) {
		for i := len(a) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			a[i], a[j] = a[j], a[i]
		}
	}
	shuffle(pri)
	shuffle(rest)
	return append(pri, rest...)
}
