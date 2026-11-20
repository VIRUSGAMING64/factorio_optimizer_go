package route

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

func TestPlaceRingPorts(t *testing.T) {
	p := &plan.Plan{
		Ports: []plan.Port{
			{ID: 0, Item: "iron-ore", Rate: 1, IsInput: true, Side: "west"},
			{ID: 1, Item: "copper-ore", Rate: 1, IsInput: true, Side: "west"},
			{ID: 2, Item: "iron-plate", Rate: 1, IsInput: false, Side: "east"},
		},
	}
	pl := place.Placement{W: 20, H: 15}
	ports := PlaceRingPorts(p, pl)
	if len(ports) != 3 {
		t.Fatalf("ports %d", len(ports))
	}
	ents := RingPortEntities(ports, "fast-transport-belt")
	if len(ents) != 3 {
		t.Fatal("entities")
	}
	for _, e := range ents {
		if e.Kind != KindBelt {
			t.Fatal("kind")
		}
	}
}
