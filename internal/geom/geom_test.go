package geom

import "testing"

func TestPointManhattan(t *testing.T) {
	a := Point{0, 0}
	b := Point{3, 4}
	if a.Manhattan(b) != 7 {
		t.Fatalf("manhattan=%d want 7", a.Manhattan(b))
	}
	if a.Dist(b) < 4.9 || a.Dist(b) > 5.1 {
		t.Fatalf("dist=%v want ~5", a.Dist(b))
	}
}

func TestDirectionCycle(t *testing.T) {
	d := North
	if d.Opposite() != South {
		t.Fatal("N opposite")
	}
	if d.CW() != East {
		t.Fatal("N CW")
	}
	if d.CCW() != West {
		t.Fatal("N CCW")
	}
	if d.CW().CW().CW().CW() != North {
		t.Fatal("4 CW")
	}
	if East.Value16() != 4 || South.Value16() != 8 || West.Value16() != 12 {
		t.Fatalf("Value16 E=%d S=%d W=%d", East.Value16(), South.Value16(), West.Value16())
	}
	if DirectionFrom16(4) != East || DirectionFrom16(12) != West {
		t.Fatal("From16")
	}
}

func TestRectOps(t *testing.T) {
	r := NewRect(2, 3, 4, 5)
	if r.Width() != 4 || r.Height() != 5 || r.Area() != 20 {
		t.Fatalf("dims %d %d area %d", r.Width(), r.Height(), r.Area())
	}
	if !r.Contains(Point{2, 3}) || !r.Contains(Point{5, 7}) {
		t.Fatal("contains edges")
	}
	if r.Contains(Point{6, 3}) || r.Contains(Point{2, 8}) {
		t.Fatal("should not contain outside")
	}
	o := NewRect(5, 6, 3, 3)
	if !r.Intersects(o) {
		t.Fatal("should intersect")
	}
	far := NewRect(20, 20, 1, 1)
	if r.Intersects(far) {
		t.Fatal("should not intersect far")
	}
	cx, cy := r.Centre()
	if cx != 4 || cy != 5.5 {
		t.Fatalf("centre %v %v", cx, cy)
	}
	sides := r.SideTiles(North, 1)
	if len(sides) != 4 {
		t.Fatalf("side tiles N len=%d", len(sides))
	}
}
