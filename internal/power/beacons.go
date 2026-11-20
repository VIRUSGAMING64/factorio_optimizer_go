package power

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

// BeaconSpec describes planned beacons for a job.
type BeaconSpec struct {
	PerMachine int
	Module     string // e.g. speed-module-3
	ModuleCount int
}

// BeaconPlacement is a placed beacon.
type BeaconPlacement struct {
	X, Y int
	Modules []string
}

// SpeedBonus approximates beacon speed effect with n beacons in range (Factorio 2.0 softens with sqrt).
func SpeedBonus(n int, distributionEffectivity float64) float64 {
	if n <= 0 {
		return 0
	}
	// Each beacon contributes effectivity/sqrt(n) in 2.0
	return distributionEffectivity * float64(n) / sqrt(float64(n))
}

func sqrt(x float64) float64 {
	// simple Newton
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// ModuleEffectSpeed returns speed bonus of a module name (approximate).
func ModuleEffectSpeed(name string) float64 {
	switch name {
	case "speed-module":
		return 0.2
	case "speed-module-2":
		return 0.3
	case "speed-module-3":
		return 0.5
	default:
		return 0
	}
}

// ModuleEffectProd returns productivity bonus.
func ModuleEffectProd(name string) float64 {
	switch name {
	case "productivity-module":
		return 0.04
	case "productivity-module-2":
		return 0.06
	case "productivity-module-3":
		return 0.1
	default:
		return 0
	}
}

// QualityScaled multiplies a module bonus by quality level.
func QualityScaled(base float64, quality string) float64 {
	return base * data.ModuleBonusMultiplier(data.QualityLevel(quality))
}

// MachineRect is a minimal machine footprint for beacon coverage checks.
type MachineRect struct {
	X, Y, W, H int
}

// PlaceBeaconGrid places 3x3 beacons in free space near machines to maximise coverage.
// occupied returns true if a tile is already taken.
// Returns simple placements; caller converts to route entities.
func PlaceBeaconGrid(w, h int, machines []MachineRect, occupied func(x, y int) bool, modules []string) []BeaconPlacement {
	if len(machines) == 0 || w < 5 || h < 5 {
		return nil
	}
	if modules == nil {
		modules = []string{"speed-module-3", "speed-module-3"}
	}

	var out []BeaconPlacement
	step := 4
	placed := map[[2]int]bool{}
	for y := 1; y < h-2; y += step {
		for x := 1; x < w-2; x += step {
			free := true
			for dy := 0; dy < 3 && free; dy++ {
				for dx := 0; dx < 3 && free; dx++ {
					if occupied(x+dx, y+dy) {
						free = false
					}
				}
			}
			if !free {
				continue
			}
			covers := 0
			for _, m := range machines {
				if BeaconAffects(x, y, m.X, m.Y, m.W, m.H) {
					covers++
				}
			}
			if covers == 0 {
				continue
			}
			key := [2]int{x / 2, y / 2}
			if placed[key] {
				continue
			}
			placed[key] = true
			out = append(out, BeaconPlacement{X: x, Y: y, Modules: modules})
		}
	}
	return out
}

// BeaconReach tiles (beacon is 3x3, supply area BeaconSupply beyond).
func BeaconAffects(bx, by, mx, my, mw, mh int) bool {
	// Beacon centre
	bcx, bcy := float64(bx)+1.5, float64(by)+1.5
	// Machine centre
	mcx, mcy := float64(mx)+float64(mw)/2, float64(my)+float64(mh)/2
	reach := float64(data.BeaconSupply) + 1.5
	return (bcx-mcx)*(bcx-mcx)+(bcy-mcy)*(bcy-mcy) <= reach*reach
}

// Ensure geom import used
var _ = geom.North
