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
