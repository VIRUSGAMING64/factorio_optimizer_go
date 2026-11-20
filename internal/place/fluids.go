package place

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"

// FluidMachine reports whether the machine has fluid ingredients or products.
func FluidMachine(mi plan.MachineInstance) bool {
	for _, n := range mi.Ingredients {
		if isLikelyFluid(n) {
			return true
		}
	}
	for _, n := range mi.Products {
		if isLikelyFluid(n) {
			return true
		}
	}
	return false
}

func isLikelyFluid(name string) bool {
	switch name {
	case "water", "steam", "crude-oil", "petroleum-gas", "light-oil", "heavy-oil",
		"lubricant", "sulfuric-acid", "fluoroketone-cold", "fluoroketone-hot",
		"thruster-fuel", "thruster-oxidizer", "lava", "molten-iron", "molten-copper":
		return true
	}
	return false
}

// FluidSpreadPenalty mildly separates fluid machines from dense solid clusters
// so pipe routing has room (placeholder for full fluid stubs).
func (m *Model) FluidSpreadPenalty() float64 {
	if m.plan == nil {
		return 0
	}
	var pen float64
	for i, mi := range m.plan.Machines {
		if !FluidMachine(mi) {
			continue
		}
		// Prefer edge proximity (lower x or y) for fluid machines
		x, y := m.px[i], m.py[i]
		distEdge := x
		if y < distEdge {
			distEdge = y
		}
		if m.w-x < distEdge {
			distEdge = m.w - x
		}
		if m.h-y < distEdge {
			distEdge = m.h - y
		}
		// Penalise being deep in the centre
		if distEdge > 6 {
			pen += float64(distEdge-6) * 0.15
		}
	}
	return pen
}
