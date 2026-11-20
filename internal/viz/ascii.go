// Package viz provides simple textual visualizations of layouts for debugging.
package viz

import (
	"strings"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// RenderASCII draws a coarse ASCII map of routed entities.
// Machines are shown as 'M', belts as arrows, poles as 'P', undergrounds as 'U'.
func RenderASCII(w, h int, entities []route.Entity) string {
	if w <= 0 || h <= 0 || w > 200 || h > 200 {
		return "(layout too large or empty for ASCII render)"
	}
	grid := make([][]rune, h)
	for y := range grid {
		grid[y] = make([]rune, w)
		for x := range grid[y] {
			grid[y][x] = '.'
		}
	}

	for _, e := range entities {
		if e.X < 0 || e.Y < 0 || e.X >= w || e.Y >= h {
			continue
		}
		ch := rune('.')
		switch e.Kind {
		case route.KindMachine:
			ch = 'M'
		case route.KindBeacon:
			ch = 'B'
		case route.KindPole:
			ch = 'P'
		case route.KindBelt:
			switch e.Dir {
			case geom.North:
				ch = '^'
			case geom.South:
				ch = 'v'
			case geom.East:
				ch = '>'
			case geom.West:
				ch = '<'
			default:
				ch = '='
			}
		case route.KindUndergroundIn, route.KindUndergroundOut:
			ch = 'U'
		case route.KindInserter:
			ch = 'I'
		case route.KindSplitter:
			ch = 'S'
		case route.KindPipe:
			ch = 'o'
		}
		// Prefer machines over other entities on the same tile
		if grid[e.Y][e.X] != 'M' {
			grid[e.Y][e.X] = ch
		}
	}

	var b strings.Builder
	b.Grow(w * (h + 1))
	for y := 0; y < h; y++ {
		b.WriteString(string(grid[y]))
		b.WriteByte('\n')
	}
	return b.String()
}
