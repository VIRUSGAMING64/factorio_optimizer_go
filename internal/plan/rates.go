package plan

import (
	"math"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

// BeltThroughput returns items/s for a belt tier with optional stacking.
func BeltThroughput(tier string, stackSize int) float64 {
	t := data.BeltTierLookup(tier)
	th := t.Throughput
	if stackSize > 1 {
		th *= float64(stackSize)
	}
	return th
}

// BeltCountForRate returns how many belts are needed for a given flow rate.
func BeltCountForRate(rate float64, tier string, stackSize int) int {
	cap := BeltThroughput(tier, stackSize)
	if cap <= 0 {
		return 1
	}
	n := int(math.Ceil(rate / cap))
	if n < 1 {
		n = 1
	}
	return n
}

// InserterCount estimates how many inserters are needed to move rate items/s.
func InserterCount(rate float64, inserterName string, fromBelt bool, research config.ResearchConfig, stackSize int) int {
	r := data.InserterRate(inserterName, fromBelt, research.InserterCapacity, stackSize)
	if r <= 0 {
		return 1
	}
	n := int(math.Ceil(rate / r))
	if n < 1 {
		n = 1
	}
	return n
}

// CraftsPerSecond of one machine given recipe energy and entity speed (+ productivity).
func CraftsPerSecond(energy, speed, productivity float64) float64 {
	if energy <= 0 {
		energy = 0.5
	}
	if speed <= 0 {
		speed = 1
	}
	// productivity increases output per craft, not craft rate; craft rate is speed/energy
	return speed / energy
}

// EffectiveOutput accounts for productivity bonus on products per craft.
func EffectiveOutput(basePerCraft, productivity float64) float64 {
	return basePerCraft * (1 + productivity)
}
