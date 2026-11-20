package plan

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

func TestEffectiveCraftsWithModules(t *testing.T) {
	root := &data.RootData{
		Entities: map[string]data.Entity{
			"assembling-machine-3": {CraftingSpeed: data.FlexibleFloat(1.25)},
		},
	}
	base := EffectiveCraftsPerMachine("assembling-machine-3", 0.5, nil, 0, "normal", root)
	withMods := EffectiveCraftsPerMachine("assembling-machine-3", 0.5, []string{"speed-module-3", "speed-module-3"}, 0, "normal", root)
	if withMods <= base {
		t.Errorf("modules should increase crafts: base=%v with=%v", base, withMods)
	}
	withBeacons := EffectiveCraftsPerMachine("assembling-machine-3", 0.5, nil, 4, "normal", root)
	if withBeacons <= base {
		t.Errorf("beacons should increase crafts: base=%v with=%v", base, withBeacons)
	}
}

func TestEffectiveCraftsWithBeaconsHigher(t *testing.T) {
	root := &data.RootData{
		Entities: map[string]data.Entity{
			"assembling-machine-3": {CraftingSpeed: data.FlexibleFloat(1.25)},
		},
	}
	base := EffectiveCraftsPerMachine("assembling-machine-3", 0.5, nil, 0, "normal", root)
	with := EffectiveCraftsPerMachine("assembling-machine-3", 0.5, []string{"speed-module-3"}, 4, "normal", root)
	if with <= base {
		t.Errorf("beacons+modules should raise crafts: base=%v with=%v", base, with)
	}
}
