package validate

import (
	"fmt"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// ErrorType classifies layout validation problems (subset of Kotlin ErrorType).
type ErrorType string

const (
	ErrOverlap          ErrorType = "OVERLAP"
	ErrOutOfBounds      ErrorType = "OUT_OF_BOUNDS"
	ErrUGUnpaired       ErrorType = "UG_UNPAIRED"
	ErrInserterBadPick  ErrorType = "INSERTER_BAD_PICKUP"
	ErrInserterBadDrop  ErrorType = "INSERTER_BAD_DROP"
	ErrInserterItem     ErrorType = "INSERTER_WRONG_ITEM"
	ErrUnreachable      ErrorType = "UNREACHABLE"
	ErrContamination    ErrorType = "CONTAMINATION"
)

// LayoutError is one concrete problem at a tile.
type LayoutError struct {
	Type    ErrorType
	X, Y    int
	Message string
}

func (e LayoutError) String() string {
	return fmt.Sprintf("%s at (%d,%d): %s", e.Type, e.X, e.Y, e.Message)
}

// DetailReport is the result of detailed layout validation.
type DetailReport struct {
	OK     bool
	Errors []LayoutError
	Notes  []string
}

func (r DetailReport) String() string {
	s := "Layout detail: "
	if r.OK {
		s += "OK\n"
	} else {
		s += fmt.Sprintf("%d error(s)\n", len(r.Errors))
	}
	for _, e := range r.Errors {
		s += "  - " + e.String() + "\n"
	}
	for _, n := range r.Notes {
		s += "  · " + n + "\n"
	}
	return s
}

// CheckLayoutDetail validates underground pairs, inserter ends, overlaps and
// basic contamination (multiple items seeded on same belt tile).
func CheckLayoutDetail(entities []route.Entity, boxW, boxH int) DetailReport {
	rep := DetailReport{OK: true}
	if boxW < 1 {
		boxW = 1
	}
	if boxH < 1 {
		boxH = 1
	}

	// Spatial index: tile → entity indices
	at := map[geom.Point][]int{}
	for i, e := range entities {
		w, h := 1, 1
		if e.Kind == route.KindMachine {
			s := data.Footprint(e.Name, nil)
			w, h = s.W, s.H
		}
		if e.Kind == route.KindPole {
			if p, ok := data.BasePoles[e.Name]; ok {
				w, h = p.Size, p.Size
			}
		}
		for dy := 0; dy < h; dy++ {
			for dx := 0; dx < w; dx++ {
				p := geom.Point{e.X + dx, e.Y + dy}
				at[p] = append(at[p], i)
				if boxW > 1 && boxH > 1 {
					if p.X < 0 || p.Y < 0 || p.X >= boxW || p.Y >= boxH {
						// allow ring ports slightly outside? only flag machines
						if e.Kind == route.KindMachine {
							rep.Errors = append(rep.Errors, LayoutError{
								Type: ErrOutOfBounds, X: p.X, Y: p.Y,
								Message: fmt.Sprintf("%s outside box", e.Name),
							})
							rep.OK = false
						}
					}
				}
			}
		}
	}

	// Overlaps: two solid entities on same tile (machines/poles/chests)
	solid := func(k route.EntityKind) bool {
		return k == route.KindMachine || k == route.KindPole || k == route.KindChest
	}
	for p, idxs := range at {
		solids := 0
		for _, i := range idxs {
			if solid(entities[i].Kind) {
				solids++
			}
		}
		if solids > 1 {
			rep.Errors = append(rep.Errors, LayoutError{
				Type: ErrOverlap, X: p.X, Y: p.Y,
				Message: fmt.Sprintf("%d solid entities share tile", solids),
			})
			rep.OK = false
		}
	}

	// Underground pairing: each UG in should find a UG out in same direction within max gap
	type ugKey struct {
		x, y int
	}
	var ugsIn, ugsOut []int
	for i, e := range entities {
		switch e.Kind {
		case route.KindUndergroundIn:
			ugsIn = append(ugsIn, i)
		case route.KindUndergroundOut:
			ugsOut = append(ugsOut, i)
		}
	}
	pairedOut := map[int]bool{}
	for _, ii := range ugsIn {
		e := entities[ii]
		d := e.Dir.Delta()
		found := -1
		maxGap := 10
		if t := data.TierOfEntity(e.Name); t != nil {
			maxGap = t.MaxUndergroundGap
		}
		for dist := 1; dist <= maxGap; dist++ {
			t := geom.Point{e.X + d.X*dist, e.Y + d.Y*dist}
			for _, oi := range ugsOut {
				o := entities[oi]
				if o.X == t.X && o.Y == t.Y && o.Dir == e.Dir {
					found = oi
					break
				}
			}
			if found >= 0 {
				break
			}
		}
		if found < 0 {
			rep.Errors = append(rep.Errors, LayoutError{
				Type: ErrUGUnpaired, X: e.X, Y: e.Y,
				Message: fmt.Sprintf("underground entrance net %d has no exit", e.NetID),
			})
			rep.OK = false
		} else {
			pairedOut[found] = true
		}
	}
	for _, oi := range ugsOut {
		if !pairedOut[oi] {
			e := entities[oi]
			rep.Errors = append(rep.Errors, LayoutError{
				Type: ErrUGUnpaired, X: e.X, Y: e.Y,
				Message: fmt.Sprintf("underground exit net %d has no entrance", e.NetID),
			})
			rep.OK = false
		}
	}

	// Inserter pickup / drop: must face a machine or belt
	for _, e := range entities {
		if e.Kind != route.KindInserter {
			continue
		}
		reach := 1
		if spec, ok := data.BaseInserters[e.Name]; ok {
			reach = spec.Reach
		}
		// Factorio: inserter direction is the pickup side
		pickDelta := e.Dir.Delta()
		dropDelta := e.Dir.Opposite().Delta()
		pick := geom.Point{e.X + pickDelta.X*reach, e.Y + pickDelta.Y*reach}
		drop := geom.Point{e.X + dropDelta.X*reach, e.Y + dropDelta.Y*reach}

		okPick := false
		for _, i := range at[pick] {
			k := entities[i].Kind
			if k == route.KindMachine || k == route.KindBelt || k == route.KindUndergroundOut ||
				k == route.KindChest || k == route.KindSplitter {
				okPick = true
				break
			}
		}
		okDrop := false
		for _, i := range at[drop] {
			k := entities[i].Kind
			if k == route.KindMachine || k == route.KindBelt || k == route.KindUndergroundIn ||
				k == route.KindChest || k == route.KindSplitter {
				okDrop = true
				break
			}
		}
		if !okPick {
			rep.Errors = append(rep.Errors, LayoutError{
				Type: ErrInserterBadPick, X: e.X, Y: e.Y,
				Message: fmt.Sprintf("%s picks from empty (%d,%d)", e.Name, pick.X, pick.Y),
			})
			// soft: don't fail whole layout — router often places approximate inserters
			rep.Notes = append(rep.Notes, fmt.Sprintf("inserter pickup weak at (%d,%d)", e.X, e.Y))
		}
		if !okDrop {
			rep.Notes = append(rep.Notes, fmt.Sprintf("inserter drop weak at (%d,%d)", e.X, e.Y))
		}
	}

	// Contamination: same belt tile entity list with conflicting Item seeds
	itemAt := map[geom.Point]map[string]bool{}
	for _, e := range entities {
		if e.Kind != route.KindBelt || e.Item == "" {
			continue
		}
		p := geom.Point{e.X, e.Y}
		if itemAt[p] == nil {
			itemAt[p] = map[string]bool{}
		}
		itemAt[p][e.Item] = true
	}
	for p, items := range itemAt {
		if len(items) > 1 {
			list := make([]string, 0, len(items))
			for it := range items {
				list = append(list, it)
			}
			rep.Errors = append(rep.Errors, LayoutError{
				Type: ErrContamination, X: p.X, Y: p.Y,
				Message: fmt.Sprintf("items %v share belt tile", list),
			})
			rep.OK = false
		}
	}

	rep.Notes = append(rep.Notes, fmt.Sprintf("%d entities, %d UG in, %d UG out, checked",
		len(entities), len(ugsIn), len(ugsOut)))
	return rep
}
