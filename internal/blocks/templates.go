package blocks

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
)

// TemplateStyle selects a column layout for single-recipe blocks.
type TemplateStyle int

const (
	TemplateNarrow TemplateStyle = iota
	TemplateSplit
	TemplateWide // two mid belts for multi-ingredient
)

// TemplateResult is a fixed column layout.
type TemplateResult struct {
	Placement place.Placement
	BeltTiles []geom.Point
	Style     TemplateStyle
}

// TryColumnTemplate builds a fixed two-column layout for a single-recipe block.
func TryColumnTemplate(b Block, style TemplateStyle) (TemplateResult, bool) {
	if len(b.Machines) == 0 {
		return TemplateResult{}, false
	}
	mw, mh := b.Machines[0].W, b.Machines[0].H
	if mw < 1 {
		mw = 3
	}
	if mh < 1 {
		mh = 3
	}
	for _, m := range b.Machines {
		if m.W != 0 && m.W != mw {
			return TemplateResult{}, false
		}
		if m.H != 0 && m.H != mh {
			return TemplateResult{}, false
		}
	}
	n := len(b.Machines)
	rowPitch := mh + 1
	var placed []place.PlacedMachine
	var belts []geom.Point

	switch style {
	case TemplateNarrow:
		beltCol := mw + 1
		rightX := mw + 3
		w := rightX + mw
		for i := range b.Machines {
			row := i / 2
			y := row * rowPitch
			x := 0
			if i%2 == 1 {
				x = rightX
			}
			placed = append(placed, place.PlacedMachine{ID: i, X: x, Y: y, Rot: 0})
			for dy := 0; dy < mh; dy++ {
				belts = append(belts, geom.Point{beltCol, y + dy})
			}
		}
		pl := place.Placement{W: w, H: ((n+1)/2)*rowPitch + 1, Machines: placed}
		return TemplateResult{Placement: pl, BeltTiles: belts, Style: style}, true

	case TemplateSplit:
		midCol := mw + 1
		rightX := mw + 3
		outR := rightX + mw
		w := outR + 1
		for i := range b.Machines {
			row := i / 2
			y := row * rowPitch
			x := 1
			if i%2 == 1 {
				x = rightX
			}
			placed = append(placed, place.PlacedMachine{ID: i, X: x, Y: y, Rot: 0})
			for dy := 0; dy < mh; dy++ {
				belts = append(belts, geom.Point{midCol, y + dy})
				if i%2 == 0 {
					belts = append(belts, geom.Point{0, y + dy})
				} else {
					belts = append(belts, geom.Point{outR, y + dy})
				}
			}
		}
		pl := place.Placement{W: w, H: ((n+1)/2)*rowPitch + 1, Machines: placed}
		return TemplateResult{Placement: pl, BeltTiles: belts, Style: style}, true

	case TemplateWide:
		// out | machine | beltA | beltB | machine | out
		midA, midB := mw+1, mw+2
		rightX := mw + 4
		outR := rightX + mw
		w := outR + 1
		for i := range b.Machines {
			row := i / 2
			y := row * rowPitch
			x := 1
			if i%2 == 1 {
				x = rightX
			}
			placed = append(placed, place.PlacedMachine{ID: i, X: x, Y: y, Rot: 0})
			for dy := 0; dy < mh; dy++ {
				belts = append(belts, geom.Point{midA, y + dy}, geom.Point{midB, y + dy})
				if i%2 == 0 {
					belts = append(belts, geom.Point{0, y + dy})
				} else {
					belts = append(belts, geom.Point{outR, y + dy})
				}
			}
		}
		pl := place.Placement{W: w, H: ((n+1)/2)*rowPitch + 1, Machines: placed}
		return TemplateResult{Placement: pl, BeltTiles: belts, Style: style}, true
	}
	return TemplateResult{}, false
}

// TryBestTemplate tries narrow then split and returns the denser layout.
func TryBestTemplate(b Block) (TemplateResult, bool) {
	var best TemplateResult
	bestArea := int(^uint(0) >> 1)
	okAny := false
	for _, st := range []TemplateStyle{TemplateNarrow, TemplateSplit, TemplateWide} {
		tr, ok := TryColumnTemplate(b, st)
		if !ok {
			continue
		}
		area := tr.Placement.W * tr.Placement.H
		if area < bestArea {
			bestArea = area
			best = tr
			okAny = true
		}
	}
	return best, okAny
}

func uniqueRecipes(b Block) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range b.Machines {
		if !seen[m.Recipe] {
			seen[m.Recipe] = true
			out = append(out, m.Recipe)
		}
	}
	return out
}
