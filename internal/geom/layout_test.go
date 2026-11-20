package geom

import "testing"

func TestGridDims(t *testing.T) {
	g := NewGridDims(10, 8, 2)
	if g.FullW != 14 || g.FullH != 12 {
		t.Fatalf("full %d %d", g.FullW, g.FullH)
	}
	if !g.InBox(0, 0) || g.InBox(-1, 0) {
		t.Fatal("inbox")
	}
	if !g.InFull(-2, -2) || g.InFull(-3, 0) {
		t.Fatal("infull")
	}
	i := g.Idx(1, 2)
	if g.XOf(i) != 1 || g.YOf(i) != 2 {
		t.Fatal("roundtrip")
	}
}

func TestSideTilesAndMiddle(t *testing.T) {
	r := NewRect(5, 5, 3, 3)
	n := r.SideTiles(North, 1)
	if len(n) != 3 {
		t.Fatalf("n=%d", len(n))
	}
	m := r.SideMiddle(East, 1)
	if m.X != 5+3 {
		t.Fatalf("middle %v", m)
	}
}

func TestKindOfAndDefaultSize(t *testing.T) {
	if KindOf("transport-belt", "") != KindBelt {
		t.Fatal("belt")
	}
	if KindOf("fast-inserter", "") != KindInserter {
		t.Fatal("ins")
	}
	w, h := DefaultSize("assembling-machine-3", KindMachine, North)
	if w != 3 || h != 3 {
		t.Fatalf("%d %d", w, h)
	}
}

func TestPlacedEntityRect(t *testing.T) {
	e := PlacedEntity{Name: "x", Kind: KindMachine, X: 1, Y: 2, W: 3, H: 3}
	r := e.Rect()
	if r.Width() != 3 || r.Height() != 3 {
		t.Fatal(r)
	}
}

func TestDirectionString(t *testing.T) {
	if North.String() == "" || East.Delta().X != 1 {
		t.Fatal("dir")
	}
}
