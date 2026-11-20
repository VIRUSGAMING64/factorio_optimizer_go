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
