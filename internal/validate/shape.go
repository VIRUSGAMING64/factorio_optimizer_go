package validate

import (
	"fmt"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// ShapeReport flags belts whose predecessor has a different net (shape change).
type ShapeReport struct {
	OK     bool
	Errors []LayoutError
}

func (r ShapeReport) String() string {
	s := "Shape: "
	if r.OK {
		s += "OK\n"
	} else {
		s += fmt.Sprintf("%d issue(s)\n", len(r.Errors))
	}
	for _, e := range r.Errors {
		s += "  - " + e.String() + "\n"
	}
	return s
}

// CheckBeltShapes walks belt/underground entities and reports when a belt of net A
// is fed by a different net without a splitter (simplified shape-change check).
func CheckBeltShapes(entities []route.Entity) ShapeReport {
	rep := ShapeReport{OK: true}
	at := map[geom.Point]route.Entity{}
	for _, e := range entities {
		if e.Kind == route.KindBelt || e.Kind == route.KindUndergroundIn ||
			e.Kind == route.KindUndergroundOut || e.Kind == route.KindSplitter {
			at[geom.Point{e.X, e.Y}] = e
		}
	}
	for _, e := range entities {
		if e.Kind != route.KindBelt {
			continue
		}
		// Tile behind this belt (opposite of travel direction)
		back := geom.Point{e.X - e.Dir.Delta().X, e.Y - e.Dir.Delta().Y}
		prev, ok := at[back]
		if !ok {
			continue
		}
		if prev.Kind == route.KindSplitter {
			continue
		}
		if prev.NetID != 0 && e.NetID != 0 && prev.NetID != e.NetID {
			rep.OK = false
			rep.Errors = append(rep.Errors, LayoutError{
				Type: "SHAPE_CHANGED", X: e.X, Y: e.Y,
				Message: fmt.Sprintf("net %d belt fed by net %d", e.NetID, prev.NetID),
			})
		}
	}
	return rep
}

// CheckLaneContamination reports belt tiles that appear more than once with
// different items (already partially in layout_detail; this is belt-graph aware).
func CheckLaneContamination(entities []route.Entity) []LayoutError {
	type key struct{ x, y int }
	items := map[key]map[string]bool{}
	for _, e := range entities {
		if e.Kind != route.KindBelt || e.Item == "" {
			continue
		}
		k := key{e.X, e.Y}
		if items[k] == nil {
			items[k] = map[string]bool{}
		}
		items[k][e.Item] = true
	}
	var errs []LayoutError
	for k, set := range items {
		if len(set) > 1 {
			list := make([]string, 0, len(set))
			for it := range set {
				list = append(list, it)
			}
			errs = append(errs, LayoutError{
				Type: ErrContamination, X: k.x, Y: k.y,
				Message: fmt.Sprintf("items %v share tile", list),
			})
		}
	}
	return errs
}
