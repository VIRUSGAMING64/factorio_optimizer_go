package validate

import (
	"fmt"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// InserterReport summarises pickup/drop validation.
type InserterReport struct {
	OK       bool
	Errors   []LayoutError
	Warnings []string
}

func (r InserterReport) String() string {
	s := "Inserters: "
	if r.OK {
		s += "OK"
	} else {
		s += fmt.Sprintf("%d errors", len(r.Errors))
	}
	s += "\n"
	for _, e := range r.Errors {
		s += "  - " + e.String() + "\n"
	}
	for _, w := range r.Warnings {
		s += "  · " + w + "\n"
	}
	return s
}

// CheckInserters validates every inserter has a plausible pickup and drop target.
// hard=true promotes weak ends to errors; hard=false keeps them as warnings.
func CheckInserters(entities []route.Entity, hard bool) InserterReport {
	rep := InserterReport{OK: true}
	at := map[geom.Point][]route.Entity{}
	for _, e := range entities {
		w, h := 1, 1
		if e.Kind == route.KindMachine {
			s := data.Footprint(e.Name, nil)
			w, h = s.W, s.H
		}
		for dy := 0; dy < h; dy++ {
			for dx := 0; dx < w; dx++ {
				p := geom.Point{e.X + dx, e.Y + dy}
				at[p] = append(at[p], e)
			}
		}
	}
	valid := func(k route.EntityKind) bool {
		switch k {
		case route.KindMachine, route.KindBelt, route.KindUndergroundIn, route.KindUndergroundOut,
			route.KindChest, route.KindSplitter:
			return true
		}
		return false
	}
	for _, e := range entities {
		if e.Kind != route.KindInserter {
			continue
		}
		reach := 1
		if spec, ok := data.BaseInserters[e.Name]; ok {
			reach = spec.Reach
		}
		pick := geom.Point{e.X + e.Dir.Delta().X*reach, e.Y + e.Dir.Delta().Y*reach}
		drop := geom.Point{e.X + e.Dir.Opposite().Delta().X*reach, e.Y + e.Dir.Opposite().Delta().Y*reach}
		okPick, okDrop := false, false
		for _, t := range at[pick] {
			if valid(t.Kind) {
				okPick = true
				break
			}
		}
		for _, t := range at[drop] {
			if valid(t.Kind) {
				okDrop = true
				break
			}
		}
		if !okPick {
			msg := fmt.Sprintf("%s pickup empty at (%d,%d)", e.Name, pick.X, pick.Y)
			if hard {
				rep.OK = false
				rep.Errors = append(rep.Errors, LayoutError{Type: ErrInserterBadPick, X: e.X, Y: e.Y, Message: msg})
			} else {
				rep.Warnings = append(rep.Warnings, msg)
			}
		}
		if !okDrop {
			msg := fmt.Sprintf("%s drop empty at (%d,%d)", e.Name, drop.X, drop.Y)
			if hard {
				rep.OK = false
				rep.Errors = append(rep.Errors, LayoutError{Type: ErrInserterBadDrop, X: e.X, Y: e.Y, Message: msg})
			} else {
				rep.Warnings = append(rep.Warnings, msg)
			}
		}
	}
	return rep
}
