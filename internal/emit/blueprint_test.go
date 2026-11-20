package emit

import (
	"strings"
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestCompressRoundTrip(t *testing.T) {
	json := `{"blueprint":{"item":"blueprint","label":"test","entities":[],"version":1}}`
	enc := CompressAndEncode(json)
	if !strings.HasPrefix(enc, "0") {
		t.Fatal("must start with 0")
	}
	dec, err := DecodeToJSON(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.Contains(dec, "blueprint") {
		t.Fatalf("decoded: %s", dec)
	}
}

func TestFromRouteMinimal(t *testing.T) {
	ents := []route.Entity{
		{Name: "assembling-machine-3", Kind: route.KindMachine, X: 0, Y: 0, Recipe: "iron-gear-wheel"},
		{Name: "fast-transport-belt", Kind: route.KindBelt, X: 3, Y: 1, Dir: geom.East, Item: "iron-plate", NetID: 0},
		{Name: "fast-inserter", Kind: route.KindInserter, X: 2, Y: 1, Dir: geom.West, Item: "iron-plate"},
		{Name: "medium-electric-pole", Kind: route.KindPole, X: 1, Y: 3},
	}
	s, err := FromRoute(ents, "test")
	if err != nil {
		t.Fatalf("fromroute: %v", err)
	}
	if !strings.HasPrefix(s, "0") {
		t.Fatal("prefix")
	}
	js, err := DecodeToJSON(s)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, want := range []string{"assembling-machine-3", "fast-transport-belt", "iron-gear-wheel"} {
		if !strings.Contains(js, want) {
			t.Fatalf("json missing %s: %s", want, js[:min(200, len(js))])
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
