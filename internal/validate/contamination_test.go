package validate

import (
	"testing"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

func TestContaminationOK(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindBelt, X: 0, Y: 0, Item: "iron-plate"},
		{Kind: route.KindBelt, X: 1, Y: 0, Item: "iron-plate"},
	}
	r := CheckContamination(ents)
	if !r.OK {
		t.Fatal(r.Conflicts)
	}
}

func TestContaminationConflict(t *testing.T) {
	ents := []route.Entity{
		{Kind: route.KindBelt, X: 0, Y: 0, Item: "iron-plate"},
		{Kind: route.KindBelt, X: 0, Y: 0, Item: "copper-plate"},
	}
	r := CheckContamination(ents)
	if r.OK {
		t.Fatal("expected conflict")
	}
	if len(r.Conflicts) == 0 {
		t.Fatal("no conflicts listed")
	}
}
