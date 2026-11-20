package place

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestBuildNetTermsAndHPWL(t *testing.T) {
	p := &plan.Plan{
		Machines: []plan.MachineInstance{
			{ID: 0, W: 3, H: 3, Products: []string{"iron-plate"}, Ingredients: []string{"iron-ore"}},
			{ID: 1, W: 3, H: 3, Products: []string{"iron-gear-wheel"}, Ingredients: []string{"iron-plate"}},
		},
		Nets: []plan.Net{{ID: 0, Item: "iron-plate", Rate: 2}},
	}
	terms := BuildNetTerms(p)
	if len(terms) != 1 {
		t.Fatalf("terms %d", len(terms))
	}
	m := &Model{
		plan: p, w: 20, h: 20,
		mw: []int{3, 3}, mh: []int{3, 3},
		px: []int{0, 10}, py: []int{0, 0},
		netTerms: terms,
	}
	h := m.HPWL(terms)
	if h < 1 {
		t.Fatalf("hpwl %v", h)
	}
}
