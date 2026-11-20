package validate

import "testing"

func TestMaxFlowSimple(t *testing.T) {
	// s=0 -> 1 (cap 10) -> t=2 (cap 5)  and 0 -> 2 (cap 3)
	mf := NewMaxFlow(3)
	mf.Edge(0, 1, 10)
	mf.Edge(1, 2, 5)
	mf.Edge(0, 2, 3)
	flow := mf.Run(0, 2)
	if flow < 7.99 || flow > 8.01 {
		t.Fatalf("flow=%v want 8", flow)
	}
}

func TestMaxFlowNoPath(t *testing.T) {
	mf := NewMaxFlow(2)
	mf.Edge(0, 1, 0)
	if mf.Run(0, 1) != 0 {
		t.Fatal("expected 0 flow")
	}
}

func TestMaxFlowDiamond(t *testing.T) {
	// classic: s-a-t and s-b-t
	mf := NewMaxFlow(4)
	s, a, b, sink := 0, 1, 2, 3
	mf.Edge(s, a, 10)
	mf.Edge(s, b, 10)
	mf.Edge(a, sink, 6)
	mf.Edge(b, sink, 7)
	flow := mf.Run(s, sink)
	if flow < 12.99 || flow > 13.01 {
		t.Fatalf("flow=%v want 13", flow)
	}
}
