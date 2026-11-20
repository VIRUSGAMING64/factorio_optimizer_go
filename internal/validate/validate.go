// Package validate performs checks on plan, placement and route results.
package validate

import (
	"fmt"
	"strings"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// Report summarises validation findings.
type Report struct {
	OK       bool
	Messages []string
}

func (r Report) String() string {
	var b strings.Builder
	if r.OK {
		b.WriteString("OK")
		if len(r.Messages) > 0 {
			b.WriteString(" with notes:\n")
		} else {
			b.WriteByte('\n')
		}
	} else {
		b.WriteString("FAILED:\n")
	}
	for _, m := range r.Messages {
		b.WriteString("  - ")
		b.WriteString(m)
		b.WriteByte('\n')
	}
	return b.String()
}

// Check runs validation.
func Check(p *plan.Plan, pl place.Placement, rt route.Result) Report {
	var msgs []string
	ok := true

	if len(p.Machines) == 0 {
		msgs = append(msgs, "plan has no machines")
		ok = false
	}
	if len(pl.Machines) > 0 && len(pl.Machines) != len(p.Machines) {
		msgs = append(msgs, fmt.Sprintf("placement has %d machines, plan has %d", len(pl.Machines), len(p.Machines)))
		ok = false
	}
	// Factory-of-blocks may leave placement empty; machines live in route entities.

	// Overlap check (single-block placement only)
	occ := map[[2]int]int{}
	for _, m := range pl.Machines {
		mi := p.Machines[m.ID]
		for dy := 0; dy < mi.H; dy++ {
			for dx := 0; dx < mi.W; dx++ {
				key := [2]int{m.X + dx, m.Y + dy}
				if prev, used := occ[key]; used {
					msgs = append(msgs, fmt.Sprintf("overlap at (%d,%d) machines %d and %d", key[0], key[1], prev, m.ID))
					ok = false
				}
				occ[key] = m.ID
			}
		}
		if m.X < 0 || m.Y < 0 || m.X+mi.W > pl.W || m.Y+mi.H > pl.H {
			msgs = append(msgs, fmt.Sprintf("machine %d out of box at (%d,%d)", m.ID, m.X, m.Y))
			ok = false
		}
	}

	// Every machine should produce or consume something
	for _, mi := range p.Machines {
		if len(mi.Products) == 0 && len(mi.Ingredients) == 0 {
			msgs = append(msgs, fmt.Sprintf("machine %d (%s) has empty I/O", mi.ID, mi.Recipe))
		}
	}

	// Route stats
	if !rt.Success {
		msgs = append(msgs, "routing incomplete: "+rt.Message)
	}
	if rt.Belts == 0 && len(p.Nets) > 0 {
		msgs = append(msgs, "no belts placed despite nets")
	}
	if len(rt.FailedNets) > 0 {
		msgs = append(msgs, fmt.Sprintf("%d nets had routing failures: %v", len(rt.FailedNets), rt.FailedNets))
	}

	machines := 0
	belts, ugs, ins, poles := 0, 0, 0, 0
	for _, e := range rt.Entities {
		switch e.Kind {
		case route.KindMachine:
			machines++
		case route.KindBelt:
			belts++
		case route.KindUndergroundIn, route.KindUndergroundOut:
			ugs++
		case route.KindInserter:
			ins++
		case route.KindPole:
			poles++
		}
	}
	if machines != len(p.Machines) {
		msgs = append(msgs, fmt.Sprintf("route has %d machines, expected %d", machines, len(p.Machines)))
	}

	// Coverage: share of nets with at least one belt
	netsWithBelt := map[int]bool{}
	for _, e := range rt.Entities {
		if e.Kind == route.KindBelt || e.Kind == route.KindUndergroundIn {
			netsWithBelt[e.NetID] = true
		}
	}
	covered := 0
	for _, n := range p.Nets {
		if !n.IsFluid && netsWithBelt[n.ID] {
			covered++
		}
	}
	itemNets := 0
	for _, n := range p.Nets {
		if !n.IsFluid {
			itemNets++
		}
	}
	if itemNets > 0 {
		msgs = append(msgs, fmt.Sprintf("belt coverage %d/%d item nets; %d belts %d ug %d inserters %d poles",
			covered, itemNets, belts, ugs, ins, poles))
	}

	if ok && len(msgs) == 0 {
		msgs = append(msgs, "no issues")
	}
	return Report{OK: ok, Messages: msgs}
}
