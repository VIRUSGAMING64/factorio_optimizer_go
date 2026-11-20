// Package power places electric poles and optionally beacons/heat.
package power

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

// PolePlacement is a placed pole.
type PolePlacement struct {
	Name string
	X, Y int
}

// Cover reports whether a pole covers a machine centre.
func Cover(pole PolePlacement, mx, my float64) bool {
	spec, ok := data.BasePoles[pole.Name]
	if !ok {
		return false
	}
	px := float64(pole.X) + float64(spec.Size)/2
	py := float64(pole.Y) + float64(spec.Size)/2
	return mathAbs(px-mx) <= spec.SupplyHalf && mathAbs(py-my) <= spec.SupplyHalf
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// PlaceGrid places medium poles on a lattice that covers the box.
func PlaceGrid(w, h int, occupied func(x, y int) bool) []PolePlacement {
	spec := data.BasePoles["medium-electric-pole"]
	step := int(spec.SupplyHalf*2) - 1
	if step < 4 {
		step = 4
	}
	var out []PolePlacement
	for y := 1; y < h; y += step {
		for x := 1; x < w; x += step {
			placed := false
			for _, dxy := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {-1, 0}, {0, -1}, {2, 1}, {1, 2}} {
				px, py := x+dxy[0], y+dxy[1]
				if px < 0 || py < 0 || px >= w || py >= h {
					continue
				}
				if occupied(px, py) {
					continue
				}
				out = append(out, PolePlacement{Name: "medium-electric-pole", X: px, Y: py})
				placed = true
				break
			}
			_ = placed
		}
	}
	return out
}

// Ensure geom import used
var _ = geom.North
