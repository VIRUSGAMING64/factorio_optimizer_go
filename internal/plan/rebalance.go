package plan

import "math"

// Rebalance trims intermediate over-production so residuals are closer to target
// rates (conservative version of residual balancing). Mutates p in place.
func Rebalance(p *Plan, tolerance float64) {
	if p == nil || tolerance <= 0 {
		tolerance = 0.05
	}
	// For each intermediate (produced and consumed), if production exceeds
	// consumption by more than tolerance fraction, scale down machine Crafts
	// of pure producers of that item.
	for item, prod := range p.Production {
		cons := p.Consumption[item]
		if cons < 1e-9 {
			continue // final product or free byproduct
		}
		if prod <= cons*(1+tolerance) {
			continue
		}
		target := cons * (1 + tolerance*0.5)
		scale := target / prod
		if scale >= 0.99 {
			continue
		}
		for i := range p.Machines {
			m := &p.Machines[i]
			produces := false
			consumes := false
			for _, pr := range m.Products {
				if pr == item {
					produces = true
				}
			}
			for _, ing := range m.Ingredients {
				if ing == item {
					consumes = true
				}
			}
			if produces && !consumes {
				m.Crafts *= scale
				if m.Crafts < 1e-6 {
					m.Crafts = 1e-6
				}
			}
		}
		p.Production[item] = prod * scale
	}
	// Drop near-zero craft machines (optional soft prune)
	_ = math.Abs
}
