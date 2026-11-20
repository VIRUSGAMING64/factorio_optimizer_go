// Package validate – inserter throughput guarantees.
//
// CheckInserterRates estimates whether the inserters physically near each
// machine can sustain the machine's required item throughput. The estimate
// is intentionally conservative (uses data.InserterRate with a belt-pickup
// penalty) and reports warnings rather than hard failures so the existing
// pipeline stays green while still surfacing capacity risks.
package validate

import (
	"fmt"
	"math"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// InserterRateReport summarises per-machine inserter capacity checks.
type InserterRateReport struct {
	OK       bool
	Warnings []string
	Notes    []string
}

func (r InserterRateReport) String() string {
	s := "Inserter rates: "
	if r.OK {
		s += "OK"
	} else {
		s += "capacity shortfalls"
	}
	s += "\n"
	for _, w := range r.Warnings {
		s += "  - " + w + "\n"
	}
	for _, n := range r.Notes {
		s += "  · " + n + "\n"
	}
	return s
}

// CheckInserterRates verifies that inserters near each machine can meet its
// required item rate. capacityLevel is the inserter-capacity research level
// (0–7). stackSize is the belt stack size (1 for normal, up to 4).
// Returns OK=false only when a machine has zero nearby inserters but a
// positive crafts rate; otherwise capacity shortfalls are warnings.
func CheckInserterRates(p *plan.Plan, entities []route.Entity, capacityLevel, stackSize int) InserterRateReport {
	rep := InserterRateReport{OK: true}
	if p == nil || len(p.Machines) == 0 {
		return rep
	}
	if stackSize < 1 {
		stackSize = 1
	}

	// Collect inserters and machines by tile for proximity lookup.
	type insRec struct {
		e    route.Entity
		rate float64
	}
	inserters := make([]insRec, 0)
	machineTiles := map[int][]geom.Point{} // machine index → tiles it occupies

	for _, e := range entities {
		switch e.Kind {
		case route.KindInserter:
			rate := data.InserterRate(e.Name, true /*fromBelt*/, capacityLevel, stackSize)
			inserters = append(inserters, insRec{e: e, rate: rate})
		case route.KindMachine:
			size := data.Footprint(e.Name, nil)
			var tiles []geom.Point
			for dy := 0; dy < size.H; dy++ {
				for dx := 0; dx < size.W; dx++ {
					tiles = append(tiles, geom.Point{e.X + dx, e.Y + dy})
				}
			}
			idx := -1
			for i, mi := range p.Machines {
				if mi.Entity == e.Name && mi.Recipe == e.Recipe && machineTiles[i] == nil {
					idx = i
					break
				}
			}
			if idx >= 0 {
				machineTiles[idx] = tiles
			}
		}
	}

	for i, mi := range p.Machines {
		if mi.Crafts <= 0 {
			continue
		}
		tiles := machineTiles[i]
		if len(tiles) == 0 {
			continue
		}
		required := mi.Crafts * 6
		if required < 0.1 {
			required = 0.1
		}

		var capacity float64
		near := 0
		for _, ins := range inserters {
			if inserterNearMachine(ins.e, tiles) {
				capacity += ins.rate
				near++
			}
		}

		if near == 0 {
			rep.OK = false
			rep.Warnings = append(rep.Warnings,
				fmt.Sprintf("machine %d (%s) crafts %.2f/s but has no nearby inserters",
					mi.ID, mi.Recipe, mi.Crafts))
			continue
		}
		if capacity < required*0.8 {
			rep.Warnings = append(rep.Warnings,
				fmt.Sprintf("machine %d (%s) needs ~%.1f items/s, nearby inserters supply %.1f (n=%d)",
					mi.ID, mi.Recipe, required, capacity, near))
		} else {
			rep.Notes = append(rep.Notes,
				fmt.Sprintf("machine %d (%s) capacity %.1f >= required %.1f",
					mi.ID, mi.Recipe, capacity, required))
		}
	}

	if len(rep.Warnings) == 0 && len(rep.Notes) == 0 {
		rep.Notes = append(rep.Notes, "no machines with positive craft rate found in entities")
	}
	return rep
}

func inserterNearMachine(e route.Entity, machineTiles []geom.Point) bool {
	reach := 1
	if spec, ok := data.BaseInserters[e.Name]; ok {
		reach = spec.Reach
	}
	d := e.Dir.Delta()
	candidates := []geom.Point{
		{e.X, e.Y},
		{e.X + d.X*reach, e.Y + d.Y*reach},
		{e.X - d.X*reach, e.Y - d.Y*reach},
	}
	tileSet := map[geom.Point]bool{}
	for _, t := range machineTiles {
		tileSet[t] = true
		for _, dir := range geom.AllDirections {
			dd := dir.Delta()
			tileSet[geom.Point{t.X + dd.X, t.Y + dd.Y}] = true
		}
	}
	for _, c := range candidates {
		if tileSet[c] {
			return true
		}
	}
	return false
}

func EstimatedInserterCapacity(name string, fromBelt bool, capacityLevel, stackSize int) float64 {
	return data.InserterRate(name, fromBelt, capacityLevel, stackSize)
}

func TotalCapacity(names []string, fromBelt bool, capacityLevel, stackSize int) float64 {
	var s float64
	for _, n := range names {
		s += EstimatedInserterCapacity(n, fromBelt, capacityLevel, stackSize)
	}
	return math.Round(s*100) / 100
}
