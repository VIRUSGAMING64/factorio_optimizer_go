package place

import (
	"math"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// NetTerm is one weighted half-perimeter wire-length contribution.
type NetTerm struct {
	Producers []int // machine indices
	Consumers []int
	Weight    float64 // typically plan net rate
	Item      string
}

// BuildNetTerms derives producer/consumer groups for each non-fluid net.
func BuildNetTerms(p *plan.Plan) []NetTerm {
	if p == nil {
		return nil
	}
	var terms []NetTerm
	for _, net := range p.Nets {
		if net.IsFluid || net.Rate < 1e-12 {
			continue
		}
		var prods, cons []int
		for i, mi := range p.Machines {
			for _, pr := range mi.Products {
				if pr == net.Item {
					prods = append(prods, i)
					break
				}
			}
			for _, ing := range mi.Ingredients {
				if ing == net.Item {
					cons = append(cons, i)
					break
				}
			}
		}
		if len(prods) == 0 || len(cons) == 0 {
			continue
		}
		w := net.Rate
		if w < 0.1 {
			w = 0.1
		}
		terms = append(terms, NetTerm{Producers: prods, Consumers: cons, Weight: w, Item: net.Item})
	}
	return terms
}

// HPWL returns sum over nets of weight * half-perimeter of the bounding box of
// all terminals (producer + consumer centres).
func (m *Model) HPWL(terms []NetTerm) float64 {
	if len(terms) == 0 {
		return 0
	}
	var total float64
	for _, t := range terms {
		minX, minY := math.MaxFloat64, math.MaxFloat64
		maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
		n := 0
		for _, ids := range [][]int{t.Producers, t.Consumers} {
			for _, id := range ids {
				if id < 0 || id >= len(m.px) {
					continue
				}
				cx := float64(m.px[id]) + float64(m.mw[id])/2
				cy := float64(m.py[id]) + float64(m.mh[id])/2
				if cx < minX {
					minX = cx
				}
				if cy < minY {
					minY = cy
				}
				if cx > maxX {
					maxX = cx
				}
				if cy > maxY {
					maxY = cy
				}
				n++
			}
		}
		if n < 2 {
			continue
		}
		hpwl := (maxX - minX) + (maxY - minY)
		total += t.Weight * hpwl
	}
	return total
}
