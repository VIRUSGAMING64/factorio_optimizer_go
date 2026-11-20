package validate

import (
	"fmt"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// LayoutRateReport checks whether belt connectivity can carry planned item rates.
type LayoutRateReport struct {
	OK       bool
	Messages []string
}

func (r LayoutRateReport) String() string {
	s := "Layout rates: "
	if r.OK {
		s += "OK\n"
	} else {
		s += "ISSUES\n"
	}
	for _, m := range r.Messages {
		s += "  - " + m + "\n"
	}
	return s
}

// CheckLayoutRates builds a flow network:
//
//	source → producer belt nodes (cap = production rate)
//	belt graph edges (cap = belt throughput)
//	consumer belt nodes → sink (cap = consumption rate)
//
// and verifies max-flow meets demand per item.
func CheckLayoutRates(p *plan.Plan, entities []route.Entity, beltTier string) LayoutRateReport {
	rep := LayoutRateReport{OK: true}
	tier := data.BeltTierLookup(beltTier)
	cap := tier.Throughput * 0.95
	if cap < 1 {
		cap = 15
	}

	g := BuildBeltGraph(entities)
	g.PropagateItems()

	// Group belt nodes by item
	itemNodes := map[string][]int{}
	for i, n := range g.Nodes {
		if n.Item != "" {
			itemNodes[n.Item] = append(itemNodes[n.Item], i)
		}
		for it := range g.Items[i] {
			itemNodes[it] = append(itemNodes[it], i)
		}
	}

	for _, net := range p.Nets {
		if net.IsFluid || net.Rate < 1e-9 {
			continue
		}
		nodes := uniqueInts(itemNodes[net.Item])
		if len(nodes) == 0 {
			rep.Messages = append(rep.Messages, fmt.Sprintf("%s: no belt nodes for item", net.Item))
			// not necessarily a hard fail if ports only
			continue
		}

		// Flow network: source, nodes..., sink
		source := len(nodes)
		sink := source + 1
		mf := NewMaxFlow(sink + 1)
		idxOf := map[int]int{}
		for i, ni := range nodes {
			idxOf[ni] = i
		}
		// edges between belt nodes that are successors in graph
		for _, ni := range nodes {
			u := idxOf[ni]
			for _, sj := range g.Succ[ni] {
				if v, ok := idxOf[sj]; ok {
					mf.Edge(u, v, cap)
				}
			}
			// source supplies every node (producers inject anywhere the item is seeded)
			mf.Edge(source, u, net.Rate)
			// sink drains every node
			mf.Edge(u, sink, net.Rate)
		}
		flow := mf.Run(source, sink)
		if flow+1e-6 < net.Rate {
			rep.OK = false
			rep.Messages = append(rep.Messages,
				fmt.Sprintf("%s: belt flow %.3f < plan rate %.3f", net.Item, flow, net.Rate))
		} else {
			rep.Messages = append(rep.Messages,
				fmt.Sprintf("%s: belt flow %.3f >= %.3f OK", net.Item, flow, net.Rate))
		}
	}
	if len(rep.Messages) == 0 {
		rep.Messages = append(rep.Messages, "no item nets to check")
	}
	return rep
}

func uniqueInts(a []int) []int {
	seen := map[int]bool{}
	var out []int
	for _, v := range a {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
