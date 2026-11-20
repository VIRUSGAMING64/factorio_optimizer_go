package plan

import (
	"math"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

// BalanceResiduals iteratively reduces overproduction of intermediate items
// by scaling down the crafts of their primary producers, provided the reduction
// does not create a deficit for any other item. Pure function; safe to call
// multiple times. Returns the number of adjustments performed.
func BalanceResiduals(root *data.RootData, selected map[string]float64, targets map[string]float64, maxPasses int) int {
	if maxPasses < 1 {
		maxPasses = 3
	}
	adjustments := 0
	for pass := 0; pass < maxPasses; pass++ {
		prod := map[string]float64{}
		cons := map[string]float64{}
		// Map item → list of (recipe, amountPerCraft) that produce it
		producersOf := map[string][]struct {
			recipe string
			amt    float64
		}{}

		for recipe, crafts := range selected {
			if crafts <= 1e-12 {
				continue
			}
			r, ok := root.Recipes[recipe]
			if !ok {
				continue
			}
			for _, pr := range r.Products {
				amt := pr.ExpectedAmount()
				if amt > 0 {
					prod[pr.Name] += crafts * amt
					producersOf[pr.Name] = append(producersOf[pr.Name], struct {
						recipe string
						amt    float64
					}{recipe, amt})
				}
			}
			for _, ing := range r.Ingredients {
				cons[ing.Name] += crafts * ing.Amount
			}
		}

		changed := false
		for item, p := range prod {
			if _, isTarget := targets[item]; isTarget {
				continue
			}
			c := cons[item]
			surplus := p - c
			if surplus <= 1e-6 {
				continue
			}
			// Try to reduce the largest producer of this item
			cands := producersOf[item]
			if len(cands) == 0 {
				continue
			}
			// Pick the producer with the highest current contribution
			bestIdx := 0
			bestContrib := 0.0
			for i, cnd := range cands {
				contrib := selected[cnd.recipe] * cnd.amt
				if contrib > bestContrib {
					bestContrib = contrib
					bestIdx = i
				}
			}
			cnd := cands[bestIdx]
			if cnd.amt <= 0 || selected[cnd.recipe] <= 1e-12 {
				continue
			}
			// How many crafts we can safely remove
			maxReduce := surplus / cnd.amt
			// Do not reduce more than 30 % of this recipe in one step (stability)
			maxReduce = math.Min(maxReduce, selected[cnd.recipe]*0.3)
			if maxReduce < 1e-4 {
				continue
			}
			// Simulate the reduction and check no new deficit is created for other products
			// of the same recipe (conservative: only reduce if the recipe has no other non-target products that would go negative)
			safe := true
			r := root.Recipes[cnd.recipe]
			for _, pr := range r.Products {
				if pr.Name == item {
					continue
				}
				if _, isTarget := targets[pr.Name]; isTarget {
					// Never reduce a recipe that also produces a target
					safe = false
					break
				}
				// Check projected production of other products
				projProd := prod[pr.Name] - maxReduce*pr.ExpectedAmount()
				if projProd < cons[pr.Name]-1e-6 {
					safe = false
					break
				}
			}
			if !safe {
				continue
			}
			selected[cnd.recipe] -= maxReduce
			if selected[cnd.recipe] < 1e-9 {
				selected[cnd.recipe] = 0
			}
			adjustments++
			changed = true
		}
		if !changed {
			break
		}
	}
	return adjustments
}
