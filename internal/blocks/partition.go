// Package blocks implements factory-of-blocks: splitting large plans into
// placeable blocks, floorplanning them, and stitching ports.
package blocks

import (
	"sort"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// Block is a subset of machines that should be placed together.
type Block struct {
	ID       int
	Machines []plan.MachineInstance
	// Items that enter / leave this block
	Inputs  []string
	Outputs []string
}

// Partition splits a plan into blocks of at most maxMachines each.
// Strategy: group by recipe first, then pack groups into blocks (first-fit decreasing).
func Partition(p *plan.Plan, maxMachines int) []Block {
	if maxMachines < 1 {
		maxMachines = 15
	}
	if len(p.Machines) <= maxMachines {
		return []Block{{ID: 0, Machines: append([]plan.MachineInstance{}, p.Machines...)}}
	}

	// Group by recipe
	type grp struct {
		recipe string
		list   []plan.MachineInstance
	}
	byRecipe := map[string]*grp{}
	for _, mi := range p.Machines {
		g, ok := byRecipe[mi.Recipe]
		if !ok {
			g = &grp{recipe: mi.Recipe}
			byRecipe[mi.Recipe] = g
		}
		g.list = append(g.list, mi)
	}
	groups := make([]*grp, 0, len(byRecipe))
	for _, g := range byRecipe {
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i].list) > len(groups[j].list)
	})

	var blocks []Block
	var cur []plan.MachineInstance
	flush := func() {
		if len(cur) == 0 {
			return
		}
		b := Block{ID: len(blocks), Machines: append([]plan.MachineInstance{}, cur...)}
		inSet, outSet := map[string]bool{}, map[string]bool{}
		for _, mi := range cur {
			for _, ing := range mi.Ingredients {
				inSet[ing] = true
			}
			for _, pr := range mi.Products {
				outSet[pr] = true
			}
		}
		// Inputs = ingredients not produced inside the block
		for ing := range inSet {
			if !outSet[ing] {
				b.Inputs = append(b.Inputs, ing)
			}
		}
		for pr := range outSet {
			if !inSet[pr] {
				b.Outputs = append(b.Outputs, pr)
			}
		}
		sort.Strings(b.Inputs)
		sort.Strings(b.Outputs)
		blocks = append(blocks, b)
		cur = nil
	}

	for _, g := range groups {
		for _, mi := range g.list {
			if len(cur) >= maxMachines {
				flush()
			}
			cur = append(cur, mi)
		}
	}
	flush()

	// Re-number machine IDs per block for local placement
	for bi := range blocks {
		for i := range blocks[bi].Machines {
			blocks[bi].Machines[i].ID = i
		}
	}
	return blocks
}

// BlockPlan converts a Block into a standalone Plan for place+route.
func BlockPlan(b Block) *plan.Plan {
	p := &plan.Plan{
		Machines:    append([]plan.MachineInstance{}, b.Machines...),
		Production:  map[string]float64{},
		Consumption: map[string]float64{},
	}
	for _, mi := range b.Machines {
		for _, pr := range mi.Products {
			p.Production[pr] += mi.Crafts // approximate
		}
		for _, ing := range mi.Ingredients {
			p.Consumption[ing] += mi.Crafts
		}
	}
	portID, netID := 0, 0
	for _, item := range b.Inputs {
		rate := p.Consumption[item] - p.Production[item]
		if rate < 0 {
			rate = 0.1
		}
		p.Ports = append(p.Ports, plan.Port{ID: portID, Item: item, Rate: rate, IsInput: true, Side: "west"})
		p.Nets = append(p.Nets, plan.Net{ID: netID, Item: item, Rate: rate})
		portID++
		netID++
	}
	for _, item := range b.Outputs {
		rate := p.Production[item] - p.Consumption[item]
		if rate < 0 {
			rate = 0.1
		}
		p.Ports = append(p.Ports, plan.Port{ID: portID, Item: item, Rate: rate, IsInput: false, Side: "east"})
		p.Nets = append(p.Nets, plan.Net{ID: netID, Item: item, Rate: rate})
		portID++
		netID++
	}
	return p
}
