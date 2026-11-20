package viz

import (
	"strings"
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestRenderASCII(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindMachine, X: 1, Y: 1},
		{Kind: route.KindBelt, X: 2, Y: 1, Dir: geom.East},
		{Kind: route.KindPole, X: 0, Y: 0},
	}
	out := RenderASCII(5, 3, ents)
	if !strings.Contains(out, "M") || !strings.Contains(out, ">") || !strings.Contains(out, "P") {
		t.Fatalf("expected symbols in render:\n%s", out)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
}
