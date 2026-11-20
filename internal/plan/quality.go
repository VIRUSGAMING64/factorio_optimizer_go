package plan

import (
	"fmt"
	"strings"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

// Quality tiers in ascending order.
var QualityTiers = []string{"normal", "uncommon", "rare", "epic", "legendary"}

// WithQuality appends a quality suffix to an item/recipe name when not normal.
func WithQuality(name, quality string) string {
	if quality == "" || quality == "normal" {
		return name
	}
	return name + "-" + quality
}

// SplitQuality extracts base name and quality from a possibly-qualified name.
func SplitQuality(name string) (base, quality string) {
	for i := len(QualityTiers) - 1; i >= 1; i-- {
		suf := "-" + QualityTiers[i]
		if strings.HasSuffix(name, suf) {
			return name[:len(name)-len(suf)], QualityTiers[i]
		}
	}
	return name, "normal"
}

// QualityLevelIndex returns 0..4 for normal..legendary.
func QualityLevelIndex(q string) int {
	for i, t := range QualityTiers {
		if t == q {
			return i
		}
	}
	return 0
}

// QualitySpread models how a crafting craft at quality level `craftQ` distributes
// product quality when quality modules are present (simplified Factorio 2.0 model).
// Returns map qualityIndex → share of product.
func QualitySpread(craftLevel int, qualityModuleBonus float64) map[int]float64 {
	// Without quality modules everything stays at craft level.
	out := map[int]float64{craftLevel: 1.0}
	if qualityModuleBonus <= 0 || craftLevel >= 4 {
		return out
	}
	// Simple model: bonus chance to upgrade one tier.
	up := qualityModuleBonus
	if up > 0.9 {
		up = 0.9
	}
	out[craftLevel] = 1 - up
	out[craftLevel+1] = up
	return out
}

// ApplyQualityToRecipe returns a virtual recipe whose item names carry quality.
func ApplyQualityToRecipe(r data.Recipe, quality string) data.Recipe {
	if quality == "" || quality == "normal" {
		return r
	}
	cp := r
	cp.Name = WithQuality(r.Name, quality)
	ings := make(data.FlexibleIngredients, len(r.Ingredients))
	for i, ing := range r.Ingredients {
		ings[i] = ing
		if !isFluidName(ing.Name) && !ing.IsFluid() {
			ings[i].Name = WithQuality(ing.Name, quality)
		}
	}
	cp.Ingredients = ings
	prods := make(data.FlexibleProducts, len(r.Products))
	for i, pr := range r.Products {
		prods[i] = pr
		if !isFluidName(pr.Name) && pr.Type != "fluid" {
			prods[i].Name = WithQuality(pr.Name, quality)
		}
	}
	cp.Products = prods
	return cp
}

func isFluidName(name string) bool {
	switch name {
	case "water", "steam", "crude-oil", "petroleum-gas", "light-oil", "heavy-oil",
		"lubricant", "sulfuric-acid", "lava", "fluorine", "lithium-brine",
		"ammoniacal-solution", "thruster-fuel", "thruster-oxidizer":
		return true
	}
	return false
}

// QualityProductionCost estimates extra recyclers/machines needed when targeting
// a higher quality of an item (very rough).
func QualityProductionCost(baseRate float64, targetQuality string) float64 {
	lvl := QualityLevelIndex(targetQuality)
	if lvl == 0 {
		return baseRate
	}
	// Each quality tier roughly needs ~5x more base crafts with quality modules.
	mult := 1.0
	for i := 0; i < lvl; i++ {
		mult *= 4.0
	}
	return baseRate * mult
}

// DescribeQuality returns a short summary string.
func DescribeQuality(q string) string {
	return fmt.Sprintf("quality=%s (tier %d)", q, QualityLevelIndex(q))
}
