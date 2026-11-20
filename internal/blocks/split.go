package blocks

import (
	"sort"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// SplitByRecipe cuts a plan into one block per recipe group (ordered by size desc).
// Useful as a coarser alternative to Partition when each recipe should be its own cell.
func SplitByRecipe(p *plan.Plan) []Block {
	groups := map[string][]plan.MachineInstance{}
	for _, mi := range p.Machines {
		groups[mi.Recipe] = append(groups[mi.Recipe], mi)
	}
	type kv struct {
		recipe string
		n      int
	}
	var order []kv
	for r, list := range groups {
		order = append(order, kv{r, len(list)})
	}
	sort.Slice(order, func(i, j int) bool {
		if order[i].n != order[j].n {
			return order[i].n > order[j].n
		}
		return order[i].recipe < order[j].recipe
	})
	var blocks []Block
	for _, o := range order {
		list := groups[o.recipe]
		b := Block{ID: len(blocks), Machines: append([]plan.MachineInstance{}, list...)}
		inSet, outSet := map[string]bool{}, map[string]bool{}
		for i := range b.Machines {
			b.Machines[i].ID = i
			for _, ing := range b.Machines[i].Ingredients {
				inSet[ing] = true
			}
			for _, pr := range b.Machines[i].Products {
				outSet[pr] = true
			}
		}
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
	}
	return blocks
}

// MergeTinyBlocks merges blocks smaller than minSize into neighbours until all
// are large enough or only one remains.
func MergeTinyBlocks(blocks []Block, minSize int) []Block {
	if minSize < 1 || len(blocks) <= 1 {
		return blocks
	}
	changed := true
	for changed {
		changed = false
		for i := 0; i < len(blocks); i++ {
			if len(blocks[i].Machines) >= minSize {
				continue
			}
			// merge into largest other block
			best := -1
			bestN := -1
			for j := range blocks {
				if j == i {
					continue
				}
				if len(blocks[j].Machines) > bestN {
					bestN = len(blocks[j].Machines)
					best = j
				}
			}
			if best < 0 {
				continue
			}
			blocks[best].Machines = append(blocks[best].Machines, blocks[i].Machines...)
			// reindex
			for k := range blocks[best].Machines {
				blocks[best].Machines[k].ID = k
			}
			blocks = append(blocks[:i], blocks[i+1:]...)
			changed = true
			break
		}
	}
	for i := range blocks {
		blocks[i].ID = i
	}
	return blocks
}
