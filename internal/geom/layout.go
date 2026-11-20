package geom

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"

// Kind classifies a placed entity.
type Kind int

const (
	KindMachine Kind = iota
	KindBelt
	KindUGIn
	KindUGOut
	KindSplitter
	KindInserter
	KindChest
	KindPole
	KindInfinityPipe
	KindInfinityChest
	KindPipe
	KindPipeUG
	KindTank
	KindBeacon
	KindLoaderIn
	KindLoaderOut
	KindHeatPipe
)

// BeltFamily reports whether this kind is part of the belt system.
func (k Kind) BeltFamily() bool {
	return k == KindBelt || k == KindUGIn || k == KindUGOut || k == KindSplitter
}

// PlacedEntity is one entity in a layout (tile top-left coordinates).
type PlacedEntity struct {
	Name    string
	Kind    Kind
	X, Y    int
	Dir     Direction
	Recipe  string
	Item    string // infinity chest / filtered inserter
	Fluid   string
	Net     int
	Mirror  bool
	Drain   bool
	W, H    int
	Modules []string
}

// Rect of this entity.
func (e PlacedEntity) Rect() Rect {
	w, h := e.W, e.H
	if w <= 0 {
		w = 1
	}
	if h <= 0 {
		h = 1
	}
	return NewRect(e.X, e.Y, w, h)
}

// DefaultSize returns footprint for an entity kind/name.
func DefaultSize(name string, kind Kind, dir Direction) (int, int) {
	switch kind {
	case KindMachine, KindBeacon:
		s := data.Footprint(name, nil)
		return s.W, s.H
	case KindSplitter:
		if dir == East || dir == West {
			return 1, 2
		}
		return 2, 1
	case KindPole:
		if p, ok := data.BasePoles[name]; ok {
			return p.Size, p.Size
		}
		return 1, 1
	case KindTank:
		return 3, 3
	case KindLoaderIn, KindLoaderOut:
		if dir == East || dir == West {
			return 2, 1
		}
		return 1, 2
	default:
		return 1, 1
	}
}

// KindOf maps an entity name (+ optional underground type) to Kind.
func KindOf(name string, ugType string) Kind {
	if t := data.TierOfEntity(name); t != nil {
		if name == t.Belt {
			return KindBelt
		}
		if name == t.Underground {
			if ugType == "output" {
				return KindUGOut
			}
			return KindUGIn
		}
		if name == t.Splitter {
			return KindSplitter
		}
	}
	if _, ok := data.BaseInserters[name]; ok {
		return KindInserter
	}
	if name == data.InfinityChest {
		return KindInfinityChest
	}
	if data.Chests[name] {
		return KindChest
	}
	if _, ok := data.BasePoles[name]; ok {
		return KindPole
	}
	if name == data.InfinityPipe {
		return KindInfinityPipe
	}
	if _, ok := data.BaseMachineSizes[name]; ok {
		return KindMachine
	}
	if name == data.Pipe {
		return KindPipe
	}
	if name == data.PipeToGround {
		return KindPipeUG
	}
	if name == data.StorageTank {
		return KindTank
	}
	if name == data.Beacon {
		return KindBeacon
	}
	if name == data.HeatPipe {
		return KindHeatPipe
	}
	for _, l := range data.Loaders {
		if name == l {
			return KindLoaderOut
		}
	}
	return KindMachine
}

// GridDims indexes a box of w×h with a surrounding margin ring.
type GridDims struct {
	W, H, Margin int
	FullW, FullH int
	Size         int
}

// NewGridDims constructs grid dimensions.
func NewGridDims(w, h, margin int) GridDims {
	fw := w + 2*margin
	fh := h + 2*margin
	return GridDims{W: w, H: h, Margin: margin, FullW: fw, FullH: fh, Size: fw * fh}
}

func (g GridDims) InFull(x, y int) bool {
	return x >= -g.Margin && y >= -g.Margin && x < g.W+g.Margin && y < g.H+g.Margin
}
func (g GridDims) InBox(x, y int) bool {
	return x >= 0 && y >= 0 && x < g.W && y < g.H
}
func (g GridDims) Idx(x, y int) int {
	return (y+g.Margin)*g.FullW + (x + g.Margin)
}
func (g GridDims) IdxTile(t Point) int { return g.Idx(t.X, t.Y) }
func (g GridDims) XOf(i int) int      { return i%g.FullW - g.Margin }
func (g GridDims) YOf(i int) int      { return i/g.FullW - g.Margin }
func (g GridDims) TileOf(i int) Point { return Point{g.XOf(i), g.YOf(i)} }

// SideTiles returns tiles at distance d outside a side of a rect (inclusive span, no corners).
func (r Rect) SideTiles(side Direction, d int) []Point {
	var out []Point
	switch side {
	case North:
		for x := r.Min.X; x < r.Max.X; x++ {
			out = append(out, Point{x, r.Min.Y - d})
		}
	case South:
		for x := r.Min.X; x < r.Max.X; x++ {
			out = append(out, Point{x, r.Max.Y + d - 1})
		}
	case West:
		for y := r.Min.Y; y < r.Max.Y; y++ {
			out = append(out, Point{r.Min.X - d, y})
		}
	case East:
		for y := r.Min.Y; y < r.Max.Y; y++ {
			out = append(out, Point{r.Max.X + d - 1, y})
		}
	}
	return out
}

// SideMiddle is the middle tile at distance d outside side.
func (r Rect) SideMiddle(side Direction, d int) Point {
	cx := r.Min.X + r.Width()/2
	cy := r.Min.Y + r.Height()/2
	switch side {
	case North:
		return Point{cx, r.Min.Y - d}
	case South:
		return Point{cx, r.Max.Y + d - 1}
	case West:
		return Point{r.Min.X - d, cy}
	case East:
		return Point{r.Max.X + d - 1, cy}
	}
	return Point{}
}
