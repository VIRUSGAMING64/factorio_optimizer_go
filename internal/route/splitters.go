package route

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// SplitterPlacement is a candidate splitter on a belt chain.
type SplitterPlacement struct {
	X, Y int
	Dir  geom.Direction
	Item string
	NetID int
}

// PlaceSplittersOnBranch looks for tiles where two outgoing belt paths diverge
// from a shared prefix and inserts a splitter entity.
//
// Simplified heuristic: if a belt tile has two different successor directions
// used by entities of the same net, place a splitter facing the dominant direction.
func PlaceSplittersOnBranch(entities []Entity, tier data.BeltTier, net plan.Net) []Entity {
	// Map tile → directions of belts leaving it (next belt of same net)
	type key struct{ X, Y int }
	belts := map[key]Entity{}
	for _, e := range entities {
		if e.Kind == KindBelt && e.NetID == net.ID {
			belts[key{e.X, e.Y}] = e
		}
	}
	if len(belts) < 3 {
		return nil
	}

	// For each belt, see if both forward-left and forward-right (or two orthogonal)
	// neighbours are also belts of this net → branch point.
	var splitters []Entity
	seen := map[key]bool{}
	for k, e := range belts {
		if seen[k] {
			continue
		}
		// Count neighbouring belts (4-connected)
		var dirs []geom.Direction
		for _, d := range geom.AllDirections {
			nk := key{k.X + d.Delta().X, k.Y + d.Delta().Y}
			if _, ok := belts[nk]; ok {
				dirs = append(dirs, d)
			}
		}
		if len(dirs) < 3 {
			// degree 3+ in the belt graph ≈ junction
			continue
		}
		// Place splitter at this tile, facing the entity's direction
		seen[k] = true
		splitters = append(splitters, Entity{
			Name:  tier.Splitter,
			Kind:  KindSplitter,
			X:     e.X,
			Y:     e.Y,
			Dir:   e.Dir,
			Item:  net.Item,
			NetID: net.ID,
		})
	}
	return splitters
}

// MergeBeltPaths finds shared prefixes of paths and marks branch points.
func MergeBeltPaths(paths [][]geom.Point) (shared []geom.Point, branches []geom.Point) {
	if len(paths) == 0 {
		return nil, nil
	}
	if len(paths) == 1 {
		return paths[0], nil
	}
	// Longest common prefix
	minLen := len(paths[0])
	for _, p := range paths[1:] {
		if len(p) < minLen {
			minLen = len(p)
		}
	}
	prefix := 0
	for i := 0; i < minLen; i++ {
		same := true
		for _, p := range paths[1:] {
			if p[i] != paths[0][i] {
				same = false
				break
			}
		}
		if !same {
			break
		}
		prefix = i + 1
	}
	if prefix > 0 {
		shared = paths[0][:prefix]
		if prefix < len(paths[0]) {
			branches = append(branches, paths[0][prefix])
		}
		for _, p := range paths[1:] {
			if prefix < len(p) {
				branches = append(branches, p[prefix])
			}
		}
	}
	return shared, branches
}
