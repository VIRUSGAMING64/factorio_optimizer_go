package blocks

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// Compaction is the result of removing empty rows/columns from a layout.
type Compaction struct {
	Entities []route.Entity
	W, H     int
	RowsRemoved, ColsRemoved int
}

// Compact removes empty rows and columns that only contain belts/pipes in a
// straight pass-through (simplified port of blocks/Compact.kt).
func Compact(entities []route.Entity, w, h int) Compaction {
	if len(entities) == 0 {
		return Compaction{W: w, H: h}
	}
	// Build occupancy
	type cell struct {
		hasMachine bool
		hasOther   bool
	}
	grid := make([][]cell, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]cell, w)
	}
	for _, e := range entities {
		if e.X < 0 || e.Y < 0 || e.X >= w || e.Y >= h {
			continue
		}
		if e.Kind == route.KindMachine || e.Kind == route.KindPole {
			// machines may be multi-tile; mark origin and expand via footprint later
			grid[e.Y][e.X].hasMachine = true
			if e.Kind == route.KindMachine {
				for dy := 0; dy < 3; dy++ {
					for dx := 0; dx < 3; dx++ {
						nx, ny := e.X+dx, e.Y+dy
						if nx < w && ny < h {
							grid[ny][nx].hasMachine = true
						}
					}
				}
			}
		} else {
			grid[e.Y][e.X].hasOther = true
		}
	}

	rowKeep := make([]bool, h)
	colKeep := make([]bool, w)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if grid[y][x].hasMachine {
				rowKeep[y] = true
				colKeep[x] = true
			}
		}
	}
	// Also keep rows/cols that have belts if neighbouring a kept machine row
	for y := 0; y < h; y++ {
		if rowKeep[y] {
			continue
		}
		for x := 0; x < w; x++ {
			if grid[y][x].hasOther && ((y > 0 && rowKeep[y-1]) || (y+1 < h && rowKeep[y+1])) {
				rowKeep[y] = true
				break
			}
		}
	}
	for x := 0; x < w; x++ {
		if colKeep[x] {
			continue
		}
		for y := 0; y < h; y++ {
			if grid[y][x].hasOther && ((x > 0 && colKeep[x-1]) || (x+1 < w && colKeep[x+1])) {
				colKeep[x] = true
				break
			}
		}
	}

	newY := make([]int, h)
	newX := make([]int, w)
	ny, nx := 0, 0
	rowsRm, colsRm := 0, 0
	for y := 0; y < h; y++ {
		if rowKeep[y] {
			newY[y] = ny
			ny++
		} else {
			newY[y] = -1
			rowsRm++
		}
	}
	for x := 0; x < w; x++ {
		if colKeep[x] {
			newX[x] = nx
			nx++
		} else {
			newX[x] = -1
			colsRm++
		}
	}

	var out []route.Entity
	for _, e := range entities {
		if e.X < 0 || e.Y < 0 || e.X >= w || e.Y >= h {
			continue
		}
		if newX[e.X] < 0 || newY[e.Y] < 0 {
			continue
		}
		e.X = newX[e.X]
		e.Y = newY[e.Y]
		out = append(out, e)
	}
	return Compaction{Entities: out, W: nx, H: ny, RowsRemoved: rowsRm, ColsRemoved: colsRm}
}

// CompactFloorplan tightens a floorplan's block boxes by removing gaps.
func CompactFloorplan(fp FloorPlan) FloorPlan {
	if len(fp.Boxes) == 0 {
		return fp
	}
	minX, minY := fp.Boxes[0].Min.X, fp.Boxes[0].Min.Y
	for _, b := range fp.Boxes {
		if b.Min.X < minX {
			minX = b.Min.X
		}
		if b.Min.Y < minY {
			minY = b.Min.Y
		}
	}
	boxes := make([]geom.Rect, len(fp.Boxes))
	maxX, maxY := 0, 0
	for i, b := range fp.Boxes {
		nb := geom.NewRect(b.Min.X-minX, b.Min.Y-minY, b.Width(), b.Height())
		boxes[i] = nb
		if nb.Max.X > maxX {
			maxX = nb.Max.X
		}
		if nb.Max.Y > maxY {
			maxY = nb.Max.Y
		}
	}
	return FloorPlan{W: maxX, H: maxY, Boxes: boxes, Cost: fp.Cost}
}
