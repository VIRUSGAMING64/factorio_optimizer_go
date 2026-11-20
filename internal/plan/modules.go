package plan

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/power"
)

// ApplyModuleAndBeaconEffects adjusts the effective crafting rate of a machine
// given optional modules and beacon coverage. Returns the multiplier for crafts/s.
func ApplyModuleAndBeaconEffects(entity string, modules []string, beacons int, quality string, root *data.RootData) float64 {
	// Base crafting speed is already in the entity data; we return additional multipliers.
	speedMult := 1.0
	prodMult := 1.0

	for _, m := range modules {
		speedMult += power.QualityScaled(power.ModuleEffectSpeed(m), quality)
		prodMult += power.QualityScaled(power.ModuleEffectProd(m), quality)
	}

	if beacons > 0 {
		// Approximate: each beacon has 2 speed-module-3 by default
		beaconSpeed := power.SpeedBonus(beacons, 0.5) * power.ModuleEffectSpeed("speed-module-3")
		speedMult += beaconSpeed
	}

	// Productivity increases output, which effectively reduces required crafts for the same rate.
	// We model it as reducing the number of crafts needed (higher effective rate).
	if prodMult > 1 {
		speedMult *= prodMult
	}
	return speedMult
}

// EffectiveCraftsPerMachine returns crafts/s one machine can sustain after modules/beacons.
func EffectiveCraftsPerMachine(entity string, energy float64, modules []string, beacons int, quality string, root *data.RootData) float64 {
	ent := root.Entities[entity]
	baseSpeed := ent.CraftingSpeed.Float()
	if baseSpeed <= 0 {
		baseSpeed = 1
	}
	if energy <= 0 {
		energy = 0.5
	}
	base := baseSpeed / energy
	mult := ApplyModuleAndBeaconEffects(entity, modules, beacons, quality, root)
	return base * mult
}
