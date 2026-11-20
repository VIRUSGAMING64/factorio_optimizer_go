// Package plan turns production targets into a concrete set of machines and nets.
package plan

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

// MachineInstance is one physical machine that will be placed later.
// Count is always 1 after Expand(); during planning it may be fractional.
type MachineInstance struct {
	ID          int
	Recipe      string
	Entity      string
	Count       float64 // fractional during planning; 1 after Expand
	W, H        int     // footprint in tiles
	Modules     []string
	Quality     string
	Crafts      float64 // crafts/s this instance should run at
	Products    []string
	Ingredients []string
}

// Net is a material flow between producers and consumers (or ports).
type Net struct {
	ID      int
	Item    string
	Rate    float64
	IsFluid bool
	Quality string
}

// Port is an input or output on the block boundary.
type Port struct {
	ID      int
	Item    string
	Rate    float64
	IsInput bool
	IsFluid bool
	Side    string // "west", "east", "north", "south"
}

// Plan is the result of the planning stage.
type Plan struct {
	Machines    []MachineInstance
	Nets        []Net
	Ports       []Port
	Production  map[string]float64
	Consumption map[string]float64
}

// Build constructs a Plan from game data and a job configuration.
// Recipe selection prefers base-game recipes and machines over Space Age alternatives.
func Build(root *data.RootData, cfg *config.RunConfiguration, _ config.Tuning) (*Plan, error) {
	if len(cfg.Targets) == 0 {
		return nil, fmt.Errorf("no targets in configuration")
	}

	// item → ranked list of recipe names (best first)
	producers := buildProducerIndex(root, cfg)

	need := map[string]float64{}
	for _, t := range cfg.Targets {
		need[t.ResolvedItem()] += t.ResolvedRate()
	}

	forbidden := map[string]struct{}{}
	for _, f := range cfg.ForbiddenRecipes {
		forbidden[f] = struct{}{}
	}

	selected := map[string]float64{} // recipe → crafts/s
	visited := map[string]bool{}
	queue := make([]string, 0, len(need))
	for item := range need {
		queue = append(queue, item)
	}

	const maxIter = 500
	for iter := 0; iter < maxIter && len(queue) > 0; iter++ {
		item := queue[0]
		queue = queue[1:]
		if visited[item] {
			continue
		}
		visited[item] = true

		rate := need[item]
		if rate <= 1e-12 {
			continue
		}
		if data.IsRaw(item) {
			continue // external supply
		}

		chosen := ""
		for _, rn := range producers[item] {
			if _, bad := forbidden[rn]; bad {
				continue
			}
			chosen = rn
			break
		}
		if chosen == "" {
			continue // treat as external
		}

		r := root.Recipes[chosen]
		var outPerCraft float64
		for _, p := range r.Products {
			if p.Name == item {
				outPerCraft += p.ExpectedAmount()
			}
		}
		if outPerCraft <= 0 {
			continue
		}

		crafts := rate / outPerCraft
		selected[chosen] += crafts

		for _, ing := range r.Ingredients {
			need[ing.Name] += crafts * ing.Amount
			if !visited[ing.Name] {
				queue = append(queue, ing.Name)
			}
		}
	}

	// Residual balancing: reduce intermediate overproduction when safe.
	targets := map[string]float64{}
	for _, t := range cfg.Targets {
		targets[t.ResolvedItem()] = t.ResolvedRate()
	}
	BalanceResiduals(root, selected, targets, 4)

	// Optional pure-Go LP refinement for more accurate multi-product rates.
	if refined := RefineRatesWithLP(root, cfg, selected, targets); refined != nil {
		selected = refined
	}

	p := &Plan{
		Production:  map[string]float64{},
		Consumption: map[string]float64{},
	}

	// Aggregate by (recipe, entity) then expand to individual machines.
	type key struct{ recipe, entity string }
	agg := map[key]float64{} // crafts/s
	entityOf := map[string]string{}

	for recipeName, craftsPerSec := range selected {
		r := root.Recipes[recipeName]
		entity := pickMachine(root, cfg, r)
		if entity == "" {
			return nil, fmt.Errorf("no machine can craft recipe %q (categories %v)", recipeName, r.AllCategories())
		}
		entityOf[recipeName] = entity
		agg[key{recipeName, entity}] += craftsPerSec
	}

	// Deterministic order.
	keys := make([]key, 0, len(agg))
	for k := range agg {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].recipe != keys[j].recipe {
			return keys[i].recipe < keys[j].recipe
		}
		return keys[i].entity < keys[j].entity
	})

	machineID := 0
	for _, k := range keys {
		craftsPerSec := agg[k]
		r := root.Recipes[k.recipe]
		ent := root.Entities[k.entity]
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
			// default speed modules when beaconing
			modules = []string{"speed-module-3", "speed-module-3"}
		}
		craftsPerMachine := EffectiveCraftsPerMachine(k.entity, energy, modules, beacons, cfg.Quality, root)
		if craftsPerMachine <= 0 {
			craftsPerMachine = 1
		}
		nMachines := int(math.Ceil(craftsPerSec / craftsPerMachine))
		if nMachines < 1 {
			nMachines = 1
		}
		craftsEach := craftsPerSec / float64(nMachines)
		fp := data.Footprint(k.entity, root.Entities)

		prods := make([]string, 0, len(r.Products))
		for _, pr := range r.Products {
			prods = append(prods, pr.Name)
		}
		ings := make([]string, 0, len(r.Ingredients))
		for _, ing := range r.Ingredients {
			ings = append(ings, ing.Name)
		}
		for i := 0; i < nMachines; i++ {
			p.Machines = append(p.Machines, MachineInstance{
				ID:          machineID,
				Recipe:      k.recipe,
				Entity:      k.entity,
				Count:       1,
				W:           fp.W,
				H:           fp.H,
				Quality:     cfg.Quality,
				Crafts:      craftsEach,
				Products:    prods,
				Ingredients: ings,
			})
			machineID++
		}

		for _, prod := range r.Products {
			p.Production[prod.Name] += craftsPerSec * prod.ExpectedAmount()
		}
		for _, ing := range r.Ingredients {
			p.Consumption[ing.Name] += craftsPerSec * ing.Amount
		}
	}

	// Ports from net demand / surplus.
	portID := 0
	netID := 0
	allItems := map[string]struct{}{}
	for k := range p.Production {
		allItems[k] = struct{}{}
	}
	for k := range p.Consumption {
		allItems[k] = struct{}{}
	}
	for _, t := range cfg.Targets {
		allItems[t.ResolvedItem()] = struct{}{}
	}

	items := make([]string, 0, len(allItems))
	for k := range allItems {
		items = append(items, k)
	}
	sort.Strings(items)

	for _, item := range items {
		prod := p.Production[item]
		cons := p.Consumption[item]
		net := cons - prod
		isFluid := false
		if _, ok := root.Fluids[item]; ok {
			isFluid = true
		}
		if net > 1e-9 {
			p.Ports = append(p.Ports, Port{
				ID: portID, Item: item, Rate: net, IsInput: true, IsFluid: isFluid, Side: "west",
			})
			p.Nets = append(p.Nets, Net{ID: netID, Item: item, Rate: net, IsFluid: isFluid})
			portID++
			netID++
		} else if net < -1e-9 {
			p.Ports = append(p.Ports, Port{
				ID: portID, Item: item, Rate: -net, IsInput: false, IsFluid: isFluid, Side: "east",
			})
			p.Nets = append(p.Nets, Net{ID: netID, Item: item, Rate: -net, IsFluid: isFluid})
			portID++
			netID++
		}
	}

	// Ensure every explicit target appears as an output port.
	for _, t := range cfg.Targets {
		item := t.ResolvedItem()
		rate := t.ResolvedRate()
		found := false
		for _, port := range p.Ports {
			if port.Item == item && !port.IsInput {
				found = true
				break
			}
		}
		if !found {
			isFluid := false
			if _, ok := root.Fluids[item]; ok {
				isFluid = true
			}
			p.Ports = append(p.Ports, Port{
				ID: portID, Item: item, Rate: rate, IsInput: false, IsFluid: isFluid, Side: "east",
			})
			portID++
		}
	}

	RefineMachineCounts(root, cfg, p)
	Rebalance(p, 0.08)
	return p, nil
}

// buildProducerIndex ranks recipes that produce each item (best first).
func buildProducerIndex(root *data.RootData, cfg *config.RunConfiguration) map[string][]string {
	type scored struct {
		name  string
		score int
	}
	byItem := map[string][]scored{}

	allowSpaceAge := cfg.Tricks["space-age"] // optional opt-in; default off for cleaner base plans
	_ = allowSpaceAge

	for name, r := range root.Recipes {
		if _, bad := data.ExcludedCategories[r.Category]; bad {
			continue
		}
		if _, bad := data.ExcludedSubgroups[r.Subgroup]; bad {
			continue
		}
		// Skip Space Age–only categories unless the recipe is the only path.
		cat := r.Category
		if strings.Contains(cat, "crushing") || strings.Contains(cat, "metallurgy") ||
			strings.Contains(cat, "organic") || strings.Contains(cat, "cryogenics") ||
			strings.Contains(cat, "electromagnetics") {
			// still allowed but scored worse
		}

		for _, p := range r.Products {
			if p.Type != "item" && p.Type != "fluid" {
				continue
			}
			score := 100
			// Prefer enabled recipes slightly.
			if r.Enabled {
				score += 5
			}
			// Prefer shorter ingredient lists (less intermediate complexity).
			score -= len(r.Ingredients)
			// Prefer base categories.
			if _, ok := data.PreferBaseCategories[r.Category]; ok {
				score += 20
			}
			// Penalise Space Age heavy recipes.
			if strings.Contains(name, "asteroid") || strings.Contains(name, "bacteria") ||
				strings.Contains(name, "bioflux") || strings.Contains(name, "yumako") ||
				strings.Contains(name, "jellynut") || strings.Contains(name, "melting") ||
				strings.Contains(name, "casting-") || strings.Contains(name, "lava") {
				score -= 50
			}
			// Prefer recipes whose main product is this item.
			if r.MainProduct != nil && r.MainProduct.Name == p.Name {
				score += 10
			}
			byItem[p.Name] = append(byItem[p.Name], scored{name, score})
		}
	}

	out := map[string][]string{}
	for item, list := range byItem {
		sort.Slice(list, func(i, j int) bool {
			if list[i].score != list[j].score {
				return list[i].score > list[j].score
			}
			return list[i].name < list[j].name
		})
		names := make([]string, len(list))
		for i, s := range list {
			names[i] = s.name
		}
		out[item] = names
	}
	return out
}

func pickMachine(root *data.RootData, cfg *config.RunConfiguration, r data.Recipe) string {
	candidates := cfg.Machines
	if len(candidates) == 0 {
		candidates = append([]string{}, data.BaseMachines...)
		candidates = append(candidates, data.SpaceAgeMachines...)
	}
	cats := map[string]struct{}{}
	for _, c := range r.AllCategories() {
		cats[c] = struct{}{}
	}
	for _, name := range candidates {
		ent, ok := root.Entities[name]
		if !ok {
			continue
		}
		for _, ec := range ent.CraftingCategories {
			if _, ok := cats[ec]; ok {
				return name
			}
		}
	}
	for name, ent := range root.Entities {
		for _, ec := range ent.CraftingCategories {
			if _, ok := cats[ec]; ok {
				return name
			}
		}
	}
	return ""
}

// Summary returns a human-readable summary of the plan.
func (p *Plan) Summary() string {
	return fmt.Sprintf("%d machines, %d nets, %d ports",
		len(p.Machines), len(p.Nets), len(p.Ports))
}

// BoundingAreaEstimate returns a rough area estimate for the initial search box.
func (p *Plan) BoundingAreaEstimate(factor float64) (w, h int) {
	area := 0
	for _, m := range p.Machines {
		area += m.W * m.H
	}
	// Extra space for belts, inserters, poles (~ factor).
	area = int(float64(area) * factor)
	if area < 16 {
		area = 16
	}
	side := int(math.Ceil(math.Sqrt(float64(area))))
	// Prefer slightly wider than tall for belt flow west→east.
	w = side + side/5
	h = side
	if w < 8 {
		w = 8
	}
	if h < 8 {
		h = 8
	}
	return w, h
}
