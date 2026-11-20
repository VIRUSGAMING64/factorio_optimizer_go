package validate

import (
	"fmt"
	"math"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// RateReport is the result of plan-level rate validation via max-flow.
type RateReport struct {
	OK       bool
	Messages []string
	// Item → required vs delivered at sinks/ports
	Shortfall map[string]float64
}

// CheckPlanRates builds a flow network from the production plan and verifies
// that producers can supply consumers and output ports at the planned rates.
//
// Graph:
//
//	source → producers (cap = crafts * output per craft)
//	producers → item nodes (same)
//	item nodes → consumers (cap = demand)
//	item nodes → output ports
//	consumers / ports → sink
func CheckPlanRates(p *plan.Plan, cfg *config.RunConfiguration, root *data.RootData) RateReport {
	rep := RateReport{OK: true, Shortfall: map[string]float64{}}

	// Collect unique items involved
	items := map[string]int{} // item → node id
	next := 0
	node := func(item string) int {
		if id, ok := items[item]; ok {
			return id
		}
		id := next
		next++
		items[item] = id
		return id
	}
	for _, mi := range p.Machines {
		for _, pr := range mi.Products {
			node(pr)
		}
		for _, ing := range mi.Ingredients {
			node(ing)
		}
	}
	for _, port := range p.Ports {
		node(port.Item)
	}

	nItems := next
	// Node layout:
	// 0..nItems-1: item nodes
	// then one node per machine (producer side)
	// then SOURCE, SINK
	machineBase := nItems
	source := machineBase + len(p.Machines)
	sink := source + 1
	n := sink + 1

	mf := NewMaxFlow(n)

	// Machine → products
	for i, mi := range p.Machines {
		mNode := machineBase + i
		// How much does this machine produce of each product?
		// Crafts/s * expected amount per craft
		r, ok := root.Recipes[mi.Recipe]
		if !ok {
			continue
		}
		for _, pr := range r.Products {
			amt := pr.ExpectedAmount() * mi.Crafts
			if amt <= 0 {
				continue
			}
			itemN := node(pr.Name)
			mf.Edge(mNode, itemN, amt)
		}
		// Source supplies the machine's production capacity (sum of outputs)
		totalOut := 0.0
		for _, pr := range r.Products {
			totalOut += pr.ExpectedAmount() * mi.Crafts
		}
		if totalOut > 0 {
			mf.Edge(source, mNode, totalOut)
		}
		// Ingredients demand: item → sink through consumer edge
		for _, ing := range r.Ingredients {
			need := ing.Amount * mi.Crafts
			if need <= 0 {
				continue
			}
			itemN := node(ing.Name)
			mf.Edge(itemN, sink, need)
		}
	}

	// External inputs (input ports): source → item
	for _, port := range p.Ports {
		if !port.IsInput {
			continue
		}
		itemN := node(port.Item)
		mf.Edge(source, itemN, port.Rate*1.001) // tiny slack
	}

	// Output ports: item → sink
	for _, port := range p.Ports {
		if port.IsInput {
			continue
		}
		itemN := node(port.Item)
		mf.Edge(itemN, sink, port.Rate)
	}

	// Raw resources unlimited from source if still short
	for item, id := range items {
		if data.IsRaw(item) {
			mf.Edge(source, id, 1e9)
		}
	}

	totalDemand := 0.0
	for _, port := range p.Ports {
		if !port.IsInput {
			totalDemand += port.Rate
		}
	}
	for _, mi := range p.Machines {
		r, ok := root.Recipes[mi.Recipe]
		if !ok {
			continue
		}
		for _, ing := range r.Ingredients {
			totalDemand += ing.Amount * mi.Crafts
		}
	}

	flow := mf.Run(source, sink)

	// Check each output target
	for _, t := range cfg.Targets {
		item := t.ResolvedItem()
		rate := t.ResolvedRate()
		// Compare production of that item
		prod := p.Production[item]
		if prod+1e-6 < rate {
			rep.OK = false
			short := rate - prod
			rep.Shortfall[item] = short
			rep.Messages = append(rep.Messages,
				fmt.Sprintf("target %s: need %.4f/s, plan produces %.4f/s (short %.4f)", item, rate, prod, short))
		}
	}

	// Machine craft feasibility: crafts/s vs machine speed
	for _, mi := range p.Machines {
		ent, ok := root.Entities[mi.Entity]
		if !ok {
			continue
		}
		r, ok := root.Recipes[mi.Recipe]
		if !ok {
			continue
		}
		speed := ent.CraftingSpeed.Float()
		if speed <= 0 {
			speed = 1
		}
		maxCrafts := speed / math.Max(r.Energy, 0.01)
		if mi.Crafts > maxCrafts*1.01 {
			rep.Messages = append(rep.Messages,
				fmt.Sprintf("machine %d (%s): crafts %.3f/s > max %.3f/s", mi.ID, mi.Recipe, mi.Crafts, maxCrafts))
		}
	}

	util := 0.0
	if totalDemand > 1e-9 {
		util = flow / totalDemand
	}
	rep.Messages = append(rep.Messages,
		fmt.Sprintf("max-flow %.4f / demand-proxy %.4f (util %.0f%%)", flow, totalDemand, util*100))

	if len(rep.Shortfall) == 0 && rep.OK {
		rep.Messages = append(rep.Messages, "all targets covered by plan production")
	}
	return rep
}

func (r RateReport) String() string {
	s := "Rate check: "
	if r.OK {
		s += "OK\n"
	} else {
		s += "FAILED\n"
	}
	for _, m := range r.Messages {
		s += "  - " + m + "\n"
	}
	return s
}
