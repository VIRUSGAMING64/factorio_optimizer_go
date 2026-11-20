package route

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

// SpoilageDrain is a filter inserter + chest that removes spoiled items from a belt end.
type SpoilageDrain struct {
	Inserter Entity
	Chest    Entity
	Item     string
}

// PlaceSpoilageDrains finds belt tiles carrying spoilable items that look like
// dead-ends and attaches a drain inserter + wooden chest.
func PlaceSpoilageDrains(entities []Entity, occ []int, w, h int) (extra []Entity, drains []SpoilageDrain) {
	type key struct{ X, Y int }
	beltAt := map[key]Entity{}
	for _, e := range entities {
		if e.Kind == KindBelt && e.Item != "" {
			if _, ok := data.Spoils[e.Item]; ok {
				beltAt[key{e.X, e.Y}] = e
			}
		}
	}
	used := map[key]bool{}
	for k, belt := range beltAt {
		// Dead-end heuristic: no forward belt of same net
		d := belt.Dir.Delta()
		fwd := key{k.X + d.X, k.Y + d.Y}
		if _, ok := beltAt[fwd]; ok {
			continue
		}
		// Place inserter beside the belt
		for _, side := range []geom.Direction{belt.Dir.CW(), belt.Dir.CCW()} {
			sd := side.Delta()
			ix, iy := k.X+sd.X, k.Y+sd.Y
			if ix < 0 || iy < 0 || ix >= w || iy >= h {
				continue
			}
			idx := iy*w + ix
			if occ != nil && idx < len(occ) && occ[idx] != 0 {
				continue
			}
			ik := key{ix, iy}
			if used[ik] {
				continue
			}
			// Chest one more step out
			cx, cy := ix+sd.X, iy+sd.Y
			if cx < 0 || cy < 0 || cx >= w || cy >= h {
				continue
			}
			cidx := cy*w + cx
			if occ != nil && cidx < len(occ) && occ[cidx] != 0 {
				continue
			}
			ins := Entity{
				Name: "inserter", Kind: KindInserter,
				X: ix, Y: iy, Dir: side.Opposite(), Item: belt.Item, NetID: belt.NetID,
			}
			chest := Entity{
				Name: "wooden-chest", Kind: KindChest,
				X: cx, Y: cy, Item: belt.Item, NetID: belt.NetID,
			}
			extra = append(extra, ins, chest)
			drains = append(drains, SpoilageDrain{Inserter: ins, Chest: chest, Item: belt.Item})
			used[ik] = true
			if occ != nil {
				if idx < len(occ) {
					occ[idx] = -3
				}
				if cidx < len(occ) {
					occ[cidx] = -3
				}
			}
			break
		}
	}
	return extra, drains
}
