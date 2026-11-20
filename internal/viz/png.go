package viz

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// RenderPNG writes a simple colored PNG of the layout.
// Machines = green, belts = yellow, poles = blue, undergrounds = magenta, inserters = cyan, beacons = orange.
func RenderPNG(path string, w, h int, entities []route.Entity, scale int) error {
	if scale < 1 {
		scale = 4
	}
	img := image.NewRGBA(image.Rect(0, 0, w*scale, h*scale))
	bg := color.RGBA{30, 30, 30, 255}
	for y := 0; y < h*scale; y++ {
		for x := 0; x < w*scale; x++ {
			img.Set(x, y, bg)
		}
	}

	set := func(x, y int, c color.Color) {
		for dy := 0; dy < scale; dy++ {
			for dx := 0; dx < scale; dx++ {
				px, py := x*scale+dx, y*scale+dy
				if px >= 0 && py >= 0 && px < w*scale && py < h*scale {
					img.Set(px, py, c)
				}
			}
		}
	}

	for _, e := range entities {
		if e.X < 0 || e.Y < 0 || e.X >= w || e.Y >= h {
			continue
		}
		var c color.Color
		switch e.Kind {
		case route.KindMachine:
			c = color.RGBA{50, 180, 50, 255}
		case route.KindBelt:
			c = color.RGBA{220, 180, 40, 255}
		case route.KindPole:
			c = color.RGBA{80, 120, 220, 255}
		case route.KindUndergroundIn, route.KindUndergroundOut:
			c = color.RGBA{200, 50, 200, 255}
		case route.KindInserter:
			c = color.RGBA{40, 200, 200, 255}
		case route.KindSplitter:
			c = color.RGBA{180, 80, 40, 255}
		case route.KindBeacon:
			c = color.RGBA{255, 140, 0, 255} // orange
		default:
			c = color.RGBA{100, 100, 100, 255}
		}
		set(e.X, e.Y, c)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
