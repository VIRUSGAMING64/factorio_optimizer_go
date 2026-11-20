package viz

import (
	"os"
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestRenderPNG(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindMachine, X: 2, Y: 2},
		{Kind: route.KindBelt, X: 3, Y: 2},
		{Kind: route.KindPole, X: 1, Y: 1},
	}
	path := "test_layout.png"
	defer os.Remove(path)
	err := RenderPNG(path, 8, 8, ents, 4)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() < 100 {
		t.Fatal("png not written or too small")
	}
}
