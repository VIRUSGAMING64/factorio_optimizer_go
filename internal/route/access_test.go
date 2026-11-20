package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

func TestAccessOptions(t *testing.T) {
	rect := geom.NewRect(5, 5, 3, 3)
	free := func(p geom.Point) bool {
		return p.X >= 0 && p.Y >= 0 && p.X < 20 && p.Y < 20
	}
	opts := AccessOptions(0, rect, true, free, true)
	if len(opts) == 0 {
		t.Fatal("no options")
	}
	long := 0
	for _, o := range opts {
		if o.Long {
			long++
		}
	}
	if long == 0 {
		t.Fatal("expected long-handed options")
	}
}

func TestDirBetween(t *testing.T) {
	if dirBetween(geom.Point{0, 0}, geom.Point{2, 0}) != geom.East {
		t.Fatal("east")
	}
	if dirBetween(geom.Point{0, 0}, geom.Point{0, 3}) != geom.South {
		t.Fatal("south")
	}
}
