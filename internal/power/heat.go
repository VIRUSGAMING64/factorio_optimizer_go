package power

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"

// HeatPipePlacement is a heat-pipe tile for Aquilo / nuclear setups.
type HeatPipePlacement struct {
	X, Y int
}

// HeatNetwork connects heat pipes in 4-neighbour adjacency.
type HeatNetwork struct {
	Pipes []HeatPipePlacement
}

// MachineCenter is the centre of a placed machine used for heat coverage checks.
type MachineCenter struct {
	X, Y float64
}

// BuildHeatGrid lays a simple lattice of heat pipes covering a box with given step.
func BuildHeatGrid(w, h, step int) HeatNetwork {
	if step < 1 {
		step = 2
	}
	var pipes []HeatPipePlacement
	for y := 0; y < h; y += step {
		for x := 0; x < w; x++ {
			pipes = append(pipes, HeatPipePlacement{X: x, Y: y})
		}
	}
	for x := 0; x < w; x += step {
		for y := 0; y < h; y++ {
			// verticals (may duplicate lattice points – OK)
			pipes = append(pipes, HeatPipePlacement{X: x, Y: y})
		}
	}
	// dedup
	seen := map[geom.Point]bool{}
	var uniq []HeatPipePlacement
	for _, p := range pipes {
		k := geom.Point{p.X, p.Y}
		if seen[k] {
			continue
		}
		seen[k] = true
		uniq = append(uniq, p)
	}
	return HeatNetwork{Pipes: uniq}
}

// BuildHeatCoverage places a minimal set of heat pipes on a step lattice so that
// every machine centre is within manhattan reach of at least one pipe.
// Candidates are free lattice points. Falls back to a full grid when no machines
// are supplied. The greedy set-cover prefers pipes that cover the most remaining
// machines; residual capacity / occupancy is respected via the occupied predicate.
func BuildHeatCoverage(w, h, step, reach int, machines []MachineCenter, occupied func(int, int) bool) HeatNetwork {
	if step < 1 {
		step = 2
	}
	if reach < 1 {
		reach = 2
	}
	if len(machines) == 0 {
		return BuildHeatGrid(w, h, step)
	}

	// Candidate lattice points that are free.
	var candidates []HeatPipePlacement
	for y := 0; y < h; y += step {
		for x := 0; x < w; x += step {
			if occupied != nil && occupied(x, y) {
				continue
			}
			candidates = append(candidates, HeatPipePlacement{X: x, Y: y})
		}
	}
	if len(candidates) == 0 {
		return BuildHeatGrid(w, h, step)
	}

	covered := make([]bool, len(machines))
	remaining := len(machines)
	var selected []HeatPipePlacement
	used := map[geom.Point]bool{}

	for remaining > 0 {
		bestIdx := -1
		bestCount := 0
		for i, c := range candidates {
			k := geom.Point{c.X, c.Y}
			if used[k] {
				continue
			}
			cnt := 0
			for mi, m := range machines {
				if covered[mi] {
					continue
				}
				dx := absF(float64(c.X)+0.5 - m.X)
				dy := absF(float64(c.Y)+0.5 - m.Y)
				if dx+dy <= float64(reach) {
					cnt++
				}
			}
			if cnt > bestCount {
				bestCount = cnt
				bestIdx = i
			}
		}
		if bestIdx < 0 || bestCount == 0 {
			// No further coverage possible from remaining candidates.
			break
		}
		c := candidates[bestIdx]
		selected = append(selected, c)
		used[geom.Point{c.X, c.Y}] = true
		for mi, m := range machines {
			if covered[mi] {
				continue
			}
			dx := absF(float64(c.X)+0.5 - m.X)
			dy := absF(float64(c.Y)+0.5 - m.Y)
			if dx+dy <= float64(reach) {
				covered[mi] = true
				remaining--
			}
		}
	}

	// If we could not cover everything, fall back to the denser grid so
	// the caller still gets a connected lattice.
	if remaining > 0 {
		return BuildHeatGrid(w, h, step)
	}
	return HeatNetwork{Pipes: selected}
}

// Covers reports whether a machine centre is within manhattan reach of any heat pipe.
func (hn HeatNetwork) Covers(mx, my float64, reach int) bool {
	if reach < 1 {
		reach = 1
	}
	for _, p := range hn.Pipes {
		dx := absF(float64(p.X)+0.5 - mx)
		dy := absF(float64(p.Y)+0.5 - my)
		if dx+dy <= float64(reach) {
			return true
		}
	}
	return false
}

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// ConnectHeatTree links the pipes of a coverage network into a connected tree.
// It builds a minimum spanning tree on manhattan distance (Prim) and inserts
// intermediate heat-pipe tiles along each edge so the result is 4-connected.
// When an occupancy predicate is supplied it first tries an A* path that
// strongly prefers free tiles; falls back to the classic L-shaped pair if A*
// cannot improve occupancy.
func ConnectHeatTree(hn HeatNetwork, occupied func(int, int) bool) HeatNetwork {
	n := len(hn.Pipes)
	if n <= 1 {
		return hn
	}

	// Deduplicate input while preserving order.
	seen := map[geom.Point]bool{}
	var pipes []HeatPipePlacement
	for _, p := range hn.Pipes {
		k := geom.Point{p.X, p.Y}
		if seen[k] {
			continue
		}
		seen[k] = true
		pipes = append(pipes, p)
	}
	n = len(pipes)
	if n <= 1 {
		return HeatNetwork{Pipes: pipes}
	}

	// Prim's algorithm: parent[i] = index of the node that connects i.
	inTree := make([]bool, n)
	parent := make([]int, n)
	dist := make([]int, n)
	for i := range dist {
		dist[i] = int(1e9)
		parent[i] = -1
	}
	dist[0] = 0

	for count := 0; count < n; count++ {
		u := -1
		best := int(1e9)
		for i := 0; i < n; i++ {
			if !inTree[i] && dist[i] < best {
				best = dist[i]
				u = i
			}
		}
		if u < 0 {
			break
		}
		inTree[u] = true
		for v := 0; v < n; v++ {
			if inTree[v] {
				continue
			}
			d := abs(pipes[u].X-pipes[v].X) + abs(pipes[u].Y-pipes[v].Y)
			if d < dist[v] {
				dist[v] = d
				parent[v] = u
			}
		}
	}

	// Collect all tiles: original + intermediates along MST edges.
	outSeen := map[geom.Point]bool{}
	var out []HeatPipePlacement
	add := func(x, y int) {
		k := geom.Point{x, y}
		if outSeen[k] {
			return
		}
		outSeen[k] = true
		out = append(out, HeatPipePlacement{X: x, Y: y})
	}
	for _, p := range pipes {
		add(p.X, p.Y)
	}

	for i := 1; i < n; i++ {
		if parent[i] < 0 {
			continue
		}
		a, b := pipes[parent[i]], pipes[i]
		path := connectPath(a, b, occupied)
		for _, tile := range path {
			add(tile.X, tile.Y)
		}
	}
	return HeatNetwork{Pipes: out}
}

// connectPath returns a 4-connected path from a to b.
// Prefers an A* route that minimises occupied tiles when a predicate is given;
// otherwise (or on failure) falls back to the classic L-shaped pair.
func connectPath(a, b HeatPipePlacement, occupied func(int, int) bool) []HeatPipePlacement {
	if occupied != nil {
		if p := heatPathAStar(a, b, occupied); len(p) > 0 {
			return p
		}
	}
	return manhattanPath(a, b, occupied)
}

// heatPathAStar finds a 4-connected path from a to b that strongly prefers
// free tiles. Occupied tiles cost 100, free tiles cost 1. Returns nil if no
// path is found within a reasonable node budget (caller falls back).
func heatPathAStar(a, b HeatPipePlacement, occupied func(int, int) bool) []HeatPipePlacement {
	if a.X == b.X && a.Y == b.Y {
		return []HeatPipePlacement{a}
	}
	type node struct {
		x, y, g, f int
		parent     *node
	}
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	open := []*node{{x: a.X, y: a.Y, g: 0, f: abs(a.X-b.X) + abs(a.Y-b.Y)}}
	closed := map[geom.Point]bool{}
	const maxNodes = 8000
	expanded := 0

	for len(open) > 0 && expanded < maxNodes {
		// extract min-f (simple linear scan; heat graphs are small)
		bestI := 0
		for i := 1; i < len(open); i++ {
			if open[i].f < open[bestI].f {
				bestI = i
			}
		}
		cur := open[bestI]
		open = append(open[:bestI], open[bestI+1:]...)
		k := geom.Point{cur.x, cur.y}
		if closed[k] {
			continue
		}
		closed[k] = true
		expanded++
		if cur.x == b.X && cur.y == b.Y {
			// reconstruct
			var rev []HeatPipePlacement
			for n := cur; n != nil; n = n.parent {
				rev = append(rev, HeatPipePlacement{X: n.x, Y: n.y})
			}
			// reverse
			for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
				rev[i], rev[j] = rev[j], rev[i]
			}
			return rev
		}
		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			nk := geom.Point{nx, ny}
			if closed[nk] {
				continue
			}
			step := 1
			if occupied(nx, ny) {
				step = 100
			}
			ng := cur.g + step
			nf := ng + abs(nx-b.X) + abs(ny-b.Y)
			open = append(open, &node{x: nx, y: ny, g: ng, f: nf, parent: cur})
		}
	}
	return nil
}

// manhattanPath returns the intermediate tiles (inclusive) that connect a to b
// with a 4-connected path. It tries both L-shapes and prefers the one with
// fewer occupied tiles.
func manhattanPath(a, b HeatPipePlacement, occupied func(int, int) bool) []HeatPipePlacement {
	if a.X == b.X && a.Y == b.Y {
		return []HeatPipePlacement{a}
	}
	hThenV := fillL(a.X, a.Y, b.X, b.Y, true)
	vThenH := fillL(a.X, a.Y, b.X, b.Y, false)
	if occupied == nil {
		return hThenV
	}
	score := func(path []HeatPipePlacement) int {
		s := 0
		for _, p := range path {
			if occupied(p.X, p.Y) {
				s++
			}
		}
		return s
	}
	if score(vThenH) < score(hThenV) {
		return vThenH
	}
	return hThenV
}

// fillL builds an L-shaped path. horizontalFirst walks X then Y.
func fillL(x1, y1, x2, y2 int, horizontalFirst bool) []HeatPipePlacement {
	var path []HeatPipePlacement
	if horizontalFirst {
		dx := 1
		if x2 < x1 {
			dx = -1
		}
		for x := x1; x != x2; x += dx {
			path = append(path, HeatPipePlacement{X: x, Y: y1})
		}
		path = append(path, HeatPipePlacement{X: x2, Y: y1})
		dy := 1
		if y2 < y1 {
			dy = -1
		}
		for y := y1; y != y2; y += dy {
			path = append(path, HeatPipePlacement{X: x2, Y: y})
		}
		path = append(path, HeatPipePlacement{X: x2, Y: y2})
	} else {
		dy := 1
		if y2 < y1 {
			dy = -1
		}
		for y := y1; y != y2; y += dy {
			path = append(path, HeatPipePlacement{X: x1, Y: y})
		}
		path = append(path, HeatPipePlacement{X: x1, Y: y2})
		dx := 1
		if x2 < x1 {
			dx = -1
		}
		for x := x1; x != x2; x += dx {
			path = append(path, HeatPipePlacement{X: x, Y: y2})
		}
		path = append(path, HeatPipePlacement{X: x2, Y: y2})
	}
	// Dedup consecutive duplicates that appear when one delta is zero.
	var uniq []HeatPipePlacement
	prev := geom.Point{-1e9, -1e9}
	for _, p := range path {
		k := geom.Point{p.X, p.Y}
		if k == prev {
			continue
		}
		uniq = append(uniq, p)
		prev = k
	}
	return uniq
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
