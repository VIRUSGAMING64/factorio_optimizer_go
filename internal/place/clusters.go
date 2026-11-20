package place

import (
	"sort"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// Cluster is a rigid group of machines placed with fixed relative offsets
// so that inserters can hand items directly (direct insertion).
type Cluster struct {
	Members []int // machine IDs
	DX, DY  []int // offsets from cluster origin
	Kind    string
}

// DIClusters builds direct-insertion clusters from a plan.
// Simplified port of place/Clusters.kt:
//   - sandwich: consumer | producer | consumer in a row
//   - duo: producer | consumer | producer (consumer of two inputs)
//   - pair: producer | consumer side by side
func DIClusters(p *plan.Plan, longHanded bool) []Cluster {
	n := len(p.Machines)
	if n == 0 {
		return nil
	}
	used := make([]bool, n)
	producersOf := map[string][]int{}
	for i, mi := range p.Machines {
		for _, pr := range mi.Products {
			producersOf[pr] = append(producersOf[pr], i)
		}
	}
	plain := func(m int) bool {
		mi := p.Machines[m]
		return mi.W == mi.H // square only; skip fluid-oriented layouts
	}
	freeProducer := func(item string, size int) int {
		for _, id := range producersOf[item] {
			if !used[id] && plain(id) && p.Machines[id].W == size {
				return id
			}
		}
		return -1
	}
	mates := func(c int) []int {
		var out []int
		for i := 0; i < n; i++ {
			if i == c || used[i] {
				continue
			}
			if p.Machines[i].Recipe == p.Machines[c].Recipe && plain(i) {
				out = append(out, i)
			}
		}
		return out
	}

	var out []Cluster
	gap := 1
	if longHanded {
		gap = 2
	}

	// Build consumer demand of internal items
	type demand struct {
		machine int
		item    string
	}
	var consumers []demand
	for i, mi := range p.Machines {
		for _, ing := range mi.Ingredients {
			if len(producersOf[ing]) > 0 {
				consumers = append(consumers, demand{i, ing})
			}
		}
	}
	// Prefer consumers deeper in the chain first (more ingredients)
	sort.SliceStable(consumers, func(i, j int) bool {
		return len(p.Machines[consumers[i].machine].Ingredients) > len(p.Machines[consumers[j].machine].Ingredients)
	})

	for _, dem := range consumers {
		c := dem.machine
		if used[c] || !plain(c) {
			continue
		}
		size := p.Machines[c].W
		item := dem.item

		// sandwich: mate | producer | c  or  c | producer | mate
		mt := mates(c)
		if len(mt) > 0 {
			prod := freeProducer(item, size)
			if prod >= 0 {
				m := mt[0]
				// consumer, producer, consumer in a row
				step := size + gap
				out = append(out, Cluster{
					Members: []int{c, prod, m},
					DX:      []int{0, step, 2 * step},
					DY:      []int{0, 0, 0},
					Kind:    "sandwich",
				})
				used[c], used[prod], used[m] = true, true, true
				continue
			}
		}

		// duo: two producers feeding one consumer that needs both products
		if len(p.Machines[c].Ingredients) >= 2 {
			ings := p.Machines[c].Ingredients
			p0 := freeProducer(ings[0], size)
			p1 := freeProducer(ings[1], size)
			if p0 >= 0 && p1 >= 0 && p0 != p1 {
				step := size + gap
				out = append(out, Cluster{
					Members: []int{p0, c, p1},
					DX:      []int{0, step, 2 * step},
					DY:      []int{0, 0, 0},
					Kind:    "duo",
				})
				used[c], used[p0], used[p1] = true, true, true
				continue
			}
		}

		// pair: single producer beside consumer
		prod := freeProducer(item, size)
		if prod >= 0 {
			step := size + gap
			out = append(out, Cluster{
				Members: []int{prod, c},
				DX:      []int{0, step},
				DY:      []int{0, 0},
				Kind:    "pair",
			})
			used[c], used[prod] = true, true
		}
	}
	return out
}

// ApplyClusters places cluster members at fixed offsets given a cluster origin.
func ApplyClusters(pl *Placement, clusters []Cluster, originX, originY []int) {
	for i, c := range clusters {
		if i >= len(originX) {
			break
		}
		ox, oy := originX[i], originY[i]
		for j, mid := range c.Members {
			for k := range pl.Machines {
				if pl.Machines[k].ID == mid {
					pl.Machines[k].X = ox + c.DX[j]
					pl.Machines[k].Y = oy + c.DY[j]
					break
				}
			}
		}
	}
}

// ClusterMemberSet returns the set of machine IDs that belong to any cluster.
func ClusterMemberSet(clusters []Cluster) map[int]bool {
	m := map[int]bool{}
	for _, c := range clusters {
		for _, id := range c.Members {
			m[id] = true
		}
	}
	return m
}
