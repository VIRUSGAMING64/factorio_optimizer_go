package plan

import (
	"math"
	"sort"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

// RefineMachineCounts re-solves machine counts for each recipe given fixed recipe choices,
// minimising total machines subject to meeting target rates (simple closed-form, not a full LP).
func RefineMachineCounts(root *data.RootData, cfg *config.RunConfiguration, p *Plan) {
	// Group machines by recipe
	type group struct {
		recipe string
		entity string
		totalCrafts float64
		indices []int
	}
	byRecipe := map[string]*group{}
	for i, mi := range p.Machines {
		g, ok := byRecipe[mi.Recipe]
		if !ok {
			g = &group{recipe: mi.Recipe, entity: mi.Entity}
			byRecipe[mi.Recipe] = g
		}
		g.totalCrafts += mi.Crafts
		g.indices = append(g.indices, i)
	}

	// Rebuild machines with exact ceil counts
	var newMachines []MachineInstance
	id := 0
	recipes := make([]string, 0, len(byRecipe))
	for r := range byRecipe {
		recipes = append(recipes, r)
	}
	sort.Strings(recipes)

	for _, rn := range recipes {
		g := byRecipe[rn]
		r := root.Recipes[rn]
		ent := root.Entities[g.entity]
		speed := ent.CraftingSpeed.Float()
		if speed <= 0 {
			speed = 1
		}
		energy := r.Energy
		if energy <= 0 {
			energy = 0.5
		}
		beacons := 0
		if cfg.Tricks != nil && cfg.Tricks["beacons"] {
			beacons = 4
		}
		modules := []string(nil)
		if beacons > 0 {
			modules = []string{"speed-module-3", "speed-module-3"}
		}
		cps := EffectiveCraftsPerMachine(g.entity, energy, modules, beacons, cfg.Quality, root)
		if cps <= 0 {
			cps = 1
		}
		n := int(math.Ceil(g.totalCrafts / cps))
		if n < 1 {
			n = 1
		}
		each := g.totalCrafts / float64(n)
		fp := data.Footprint(g.entity, root.Entities)
		prods := make([]string, 0, len(r.Products))
		for _, pr := range r.Products {
			prods = append(prods, pr.Name)
		}
		ings := make([]string, 0, len(r.Ingredients))
		for _, ing := range r.Ingredients {
			ings = append(ings, ing.Name)
		}
		for i := 0; i < n; i++ {
			newMachines = append(newMachines, MachineInstance{
				ID: id, Recipe: rn, Entity: g.entity, Count: 1,
				W: fp.W, H: fp.H, Quality: cfg.Quality, Crafts: each,
				Products: prods, Ingredients: ings,
			})
			id++
		}
	}
	p.Machines = newMachines
}

// SelectBestRecipe picks the best recipe for an item given scoring preferences.
func SelectBestRecipe(root *data.RootData, item string, forbidden map[string]struct{}, preferBase bool) string {
	type scored struct {
		name  string
		score int
	}
	var list []scored
	for name, r := range root.Recipes {
		if _, bad := forbidden[name]; bad {
			continue
		}
		if _, bad := data.ExcludedCategories[r.Category]; bad {
			continue
		}
		if _, bad := data.ExcludedSubgroups[r.Subgroup]; bad {
			continue
		}
		produces := false
		for _, p := range r.Products {
			if p.Name == item {
				produces = true
				break
			}
		}
		if !produces {
			continue
		}
		score := 100
		if r.Enabled {
			score += 5
		}
		score -= len(r.Ingredients)
		if _, ok := data.PreferBaseCategories[r.Category]; ok {
			score += 20
		}
		if preferBase {
			// penalise space-age-ish names
			n := name
			for _, bad := range []string{"asteroid", "bacteria", "bioflux", "yumako", "jellynut", "melting", "casting-", "lava"} {
				if contains(n, bad) {
					score -= 50
				}
			}
		}
		if r.MainProduct != nil && r.MainProduct.Name == item {
			score += 10
		}
		list = append(list, scored{name, score})
	}
	if len(list) == 0 {
		return ""
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score
		}
		return list[i].name < list[j].name
	})
	return list[0].name
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
