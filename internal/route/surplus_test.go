package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

func TestAddSurplusControlEmpty(t *testing.T) {
	r := AddSurplusControl(nil, nil, 0.5)
	if len(r.Wires) != 0 {
		t.Fatal("expected empty")
	}
}

func TestAddSurplusControlWithEntities(t *testing.T) {
	ents := []Entity{
		{Kind: KindPipe, X: 0, Y: 0, Item: "heavy-oil"},
		{Kind: KindMachine, X: 2, Y: 0, Recipe: "heavy-oil-cracking"},
	}
	r := AddSurplusControl(ents, map[string]string{"heavy-oil-cracking": "heavy-oil"}, 0.5)
	t.Logf("wires=%d notes=%v", len(r.Wires), r.Notes)
	_ = geom.North
}
