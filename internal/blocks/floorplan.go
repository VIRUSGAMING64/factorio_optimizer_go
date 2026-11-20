package blocks

import (
	"math"
	"math/rand"
	"time"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

// Shape is one possible size for a block.
type Shape struct {
	W, H int
}

// FloorBlock is a block to pack in the factory floorplan.
type FloorBlock struct {
	ID     int
	Shapes []Shape // alternatives (e.g. rotated)
	Pad    int     // corridor padding around the block
}

// FloorNet connects blocks (by ID) and optionally factory edges.
type FloorNet struct {
	BlockIDs []int
	Weight   float64
}

// FloorPlan is the result of packing blocks.
type FloorPlan struct {
	W, H   int
	Boxes  []geom.Rect // box of each block (indexed by block ID order)
	Cost   float64
}

type skylineSeg struct {
	x, y, w int
}

// packSkyline places rectangles using a bottom-left skyline heuristic.
// Returns positions for each block index and the overall bounding box.
func packSkyline(ws, hs []int, order []int, targetW int) (x, y []int, maxX, maxY int) {
	n := len(ws)
	x = make([]int, n)
	y = make([]int, n)
	// Initial wide skyline
	sky := []skylineSeg{{0, 0, targetW * 4}}
	maxX, maxY = 0, 0

	for _, bi := range order {
		bw, bh := ws[bi], hs[bi]
		bestX, bestY, bestIdx, bestWaste := 0, 1<<30, -1, 1<<30
		for i, seg := range sky {
			if seg.w < bw {
				continue
			}
			py := seg.y
			waste := seg.w - bw
			if py < bestY || (py == bestY && waste < bestWaste) {
				bestY = py
				bestX = seg.x
				bestIdx = i
				bestWaste = waste
			}
		}
		if bestIdx < 0 {
			// Force a new row
			bestY = maxY
			bestX = 0
			sky = append(sky, skylineSeg{0, maxY, targetW * 4})
			bestIdx = len(sky) - 1
		}
		x[bi] = bestX
		y[bi] = bestY
		if bestX+bw > maxX {
			maxX = bestX + bw
		}
		if bestY+bh > maxY {
			maxY = bestY + bh
		}
		// Split / raise the chosen skyline segment
		seg := sky[bestIdx]
		var newSky []skylineSeg
		newSky = append(newSky, sky[:bestIdx]...)
		if bestX > seg.x {
			newSky = append(newSky, skylineSeg{seg.x, seg.y, bestX - seg.x})
		}
		newSky = append(newSky, skylineSeg{bestX, bestY + bh, bw})
		if bestX+bw < seg.x+seg.w {
			newSky = append(newSky, skylineSeg{bestX + bw, seg.y, (seg.x + seg.w) - (bestX + bw)})
		}
		newSky = append(newSky, sky[bestIdx+1:]...)
		// Merge adjacent equal-height segments
		merged := make([]skylineSeg, 0, len(newSky))
		for _, s := range newSky {
			if len(merged) > 0 && merged[len(merged)-1].y == s.y && merged[len(merged)-1].x+merged[len(merged)-1].w == s.x {
				merged[len(merged)-1].w += s.w
			} else {
				merged = append(merged, s)
			}
		}
		sky = merged
	}
	return
}

// Pack packs blocks into a factory using a mix of skyline and row-major random search.
func Pack(blocks []FloorBlock, nets []FloorNet, aspect float64, attempts int) FloorPlan {
	if len(blocks) == 0 {
		return FloorPlan{}
	}
	if attempts < 1 {
		attempts = 8
	}
	if aspect < 0.2 {
		aspect = 1.0
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	best := FloorPlan{Cost: math.Inf(1)}
	for a := 0; a < attempts; a++ {
		// Choose a shape per block
		ws := make([]int, len(blocks))
		hs := make([]int, len(blocks))
		for i, b := range blocks {
			sh := b.Shapes[rng.Intn(len(b.Shapes))]
			pad := b.Pad
			ws[i] = sh.W + 2*pad
			hs[i] = sh.H + 2*pad
		}
		order := rng.Perm(len(blocks))
		area := 0
		for i := range blocks {
			area += ws[i] * hs[i]
		}
		targetW := int(math.Ceil(math.Sqrt(float64(area) * aspect)))
		if targetW < 10 {
			targetW = 10
		}

		var x, y []int
		var maxX, maxY int
		if a%3 == 0 {
			x, y, maxX, maxY = packSkyline(ws, hs, order, targetW)
		} else if a%3 == 1 {
			gp, gm := randomSequencePair(rng, len(blocks))
			x, y, maxX, maxY = packSequencePair(ws, hs, gp, gm)
		} else {
			// original row-major fallback
			x = make([]int, len(blocks))
			y = make([]int, len(blocks))
			cx, cy, rowH := 0, 0, 0
			maxX, maxY = 0, 0
			for _, bi := range order {
				if cx > 0 && cx+ws[bi] > targetW {
					cx = 0
					cy += rowH
					rowH = 0
				}
				x[bi] = cx
				y[bi] = cy
				cx += ws[bi]
				if hs[bi] > rowH {
					rowH = hs[bi]
				}
				if cx > maxX {
					maxX = cx
				}
				if cy+hs[bi] > maxY {
					maxY = cy + hs[bi]
				}
			}
			if maxY < cy+rowH {
				maxY = cy + rowH
			}
		}

		boxes := make([]geom.Rect, len(blocks))
		for i := range blocks {
			pad := blocks[i].Pad
			boxes[i] = geom.NewRect(x[i]+pad, y[i]+pad, ws[i]-2*pad, hs[i]-2*pad)
		}

		// Wire cost: HPWL of nets
		wire := 0.0
		for _, net := range nets {
			if len(net.BlockIDs) < 2 {
				continue
			}
			minX, minY := 1e9, 1e9
			maxXX, maxYY := -1e9, -1e9
			for _, id := range net.BlockIDs {
				if id < 0 || id >= len(boxes) {
					continue
				}
				cx, cy := boxes[id].Centre()
				if cx < minX {
					minX = cx
				}
				if cy < minY {
					minY = cy
				}
				if cx > maxXX {
					maxXX = cx
				}
				if cy > maxYY {
					maxYY = cy
				}
			}
			w := net.Weight
			if w <= 0 {
				w = 1
			}
			wire += w * ((maxXX - minX) + (maxYY - minY))
		}
		cost := float64(maxX*maxY) + 2*wire
		// Aspect penalty
		ratio := float64(maxX) / math.Max(float64(maxY), 1)
		if ratio < aspect {
			ratio = aspect / math.Max(ratio, 0.01)
		} else {
			ratio = ratio / aspect
		}
		cost *= 1 + 0.5*(ratio-1)*(ratio-1)

		if cost < best.Cost {
			best = FloorPlan{W: maxX, H: maxY, Boxes: boxes, Cost: cost}
		}
	}
	return best
}

// StitchOffsets returns the world-space origin of each block from a floorplan.
func StitchOffsets(fp FloorPlan) []geom.Point {
	out := make([]geom.Point, len(fp.Boxes))
	for i, b := range fp.Boxes {
		out[i] = b.Min
	}
	return out
}
