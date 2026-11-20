package blocks

import (
	"fmt"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// FactoryResult is a multi-block layout stitched into one coordinate system.
type FactoryResult struct {
	Blocks     []Block
	Floorplan  FloorPlan
	Entities   []route.Entity
	Placements []place.Placement // per-block local placements
	Message    string
}

// BuildFactory partitions a large plan, places+routes each block, floorplans them, and stitches entities.
func BuildFactory(p *plan.Plan, cfg *config.RunConfiguration, tuning config.Tuning, maxMachines int, saIters int) (FactoryResult, error) {
	if maxMachines < 1 {
		maxMachines = 15
	}
	parts := Partition(p, maxMachines)
	fr := FactoryResult{Blocks: parts}

	beltTier := cfg.BeltTier
	if beltTier == "" {
		beltTier = "fast"
	}
	areaFactor := tuning.AreaFactor
	if areaFactor < 1.1 {
		areaFactor = 1.4
	}
	if saIters < 1 {
		saIters = tuning.SAIterations
	}
	if saIters < 1 {
		saIters = 15000
	}

	floorBlocks := make([]FloorBlock, len(parts))
	var allLocal [][]route.Entity
	var placements []place.Placement

	for i, b := range parts {
		bp := BlockPlan(b)
		// Prefer column template for single-recipe homogeneous blocks
		var pl place.Placement
		var err error
		if len(uniqueRecipes(b)) == 1 {
			if tr, ok := TryBestTemplate(b); ok && tr.Placement.W > 0 {
				pl = tr.Placement
			}
		}
		if len(pl.Machines) != len(b.Machines) {
			pl, err = place.Place(bp, areaFactor, saIters, tuning.SAInitialTemp, tuning.SACoolingRate)
		}
		if err != nil {
			// try larger area
			pl, err = place.Place(bp, areaFactor*1.3, saIters, tuning.SAInitialTemp, tuning.SACoolingRate)
			if err != nil {
				return fr, fmt.Errorf("block %d place: %w", i, err)
			}
		}
		rt := route.Route(bp, pl, beltTier)
		placements = append(placements, pl)
		allLocal = append(allLocal, rt.Entities)
		floorBlocks[i] = FloorBlock{
			ID:     i,
			Shapes: []Shape{{pl.W, pl.H}, {pl.H, pl.W}},
			Pad:    2,
		}
	}
	fr.Placements = placements

	// Nets between blocks: shared items
	var nets []FloorNet
	itemBlocks := map[string][]int{}
	for i, b := range parts {
		for _, item := range append(b.Inputs, b.Outputs...) {
			itemBlocks[item] = append(itemBlocks[item], i)
		}
	}
	for _, blks := range itemBlocks {
		if len(blks) >= 2 {
			nets = append(nets, FloorNet{BlockIDs: blks, Weight: 1})
		}
	}

	fp := Pack(floorBlocks, nets, 1.2, 12)
	fr.Floorplan = fp
	offsets := StitchOffsets(fp)

	// Stitch entities into world coordinates
	var world []route.Entity
	for i, local := range allLocal {
		off := geom.Point{}
		if i < len(offsets) {
			off = offsets[i]
		}
		for _, e := range local {
			e.X += off.X
			e.Y += off.Y
			world = append(world, e)
		}
	}
	// Compact empty margins in the stitched layout
	cw, ch := fp.W, fp.H
	if cw < 1 {
		cw = 1
	}
	if ch < 1 {
		ch = 1
	}
	comp := Compact(world, cw, ch)
	fr.Entities = comp.Entities
	if comp.W > 0 {
		fr.Floorplan.W = comp.W
	}
	if comp.H > 0 {
		fr.Floorplan.H = comp.H
	}
	fr.Message = fmt.Sprintf("%d blocks, factory %d×%d, %d entities (compacted -%d rows -%d cols)",
		len(parts), fr.Floorplan.W, fr.Floorplan.H, len(fr.Entities), comp.RowsRemoved, comp.ColsRemoved)
	return fr, nil
}
