package emit

import (
	"fmt"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// VerifyKit is a small set of blueprints used to validate game constants in-game
// (inserter rates, belt throughput, underground gaps). Simplified VerifyKit.kt.
type VerifyKit struct {
	Name       string
	Entities   []route.Entity
	Blueprint  string
	Description string
}

// BuildInserterRateKit places a chest → inserter → chest pair for timing tests.
func BuildInserterRateKit(inserter string) (VerifyKit, error) {
	ents := []route.Entity{
		{Name: "iron-chest", Kind: route.KindChest, X: 0, Y: 0, Item: "iron-plate"},
		{Name: inserter, Kind: route.KindInserter, X: 1, Y: 0, Dir: geom.West, Item: "iron-plate"},
		{Name: "iron-chest", Kind: route.KindChest, X: 2, Y: 0},
	}
	bp, err := FromRoute(ents, "verify-inserter-"+inserter)
	if err != nil {
		return VerifyKit{}, err
	}
	return VerifyKit{
		Name: "inserter-rate-" + inserter,
		Entities: ents, Blueprint: bp,
		Description: fmt.Sprintf("chest-%s-chest for rate timing", inserter),
	}, nil
}

// BuildBeltThroughputKit places a line of belts for throughput checks.
func BuildBeltThroughputKit(belt string, length int) (VerifyKit, error) {
	if length < 3 {
		length = 10
	}
	var ents []route.Entity
	for i := 0; i < length; i++ {
		ents = append(ents, route.Entity{
			Name: belt, Kind: route.KindBelt, X: i, Y: 0, Dir: geom.East, Item: "iron-plate",
		})
	}
	ents = append(ents,
		route.Entity{Name: "iron-chest", Kind: route.KindChest, X: -1, Y: 0},
		route.Entity{Name: "fast-inserter", Kind: route.KindInserter, X: -1, Y: 1, Dir: geom.North},
		route.Entity{Name: "iron-chest", Kind: route.KindChest, X: length, Y: 0},
		route.Entity{Name: "fast-inserter", Kind: route.KindInserter, X: length, Y: 1, Dir: geom.North},
	)
	bp, err := FromRoute(ents, "verify-belt-"+belt)
	if err != nil {
		return VerifyKit{}, err
	}
	return VerifyKit{Name: "belt-" + belt, Entities: ents, Blueprint: bp,
		Description: fmt.Sprintf("%d-tile %s line", length, belt)}, nil
}

// BuildUndergroundGapKit places entrance/exit pairs at increasing gaps.
func BuildUndergroundGapKit(ug string, maxGap int) (VerifyKit, error) {
	if maxGap < 2 {
		maxGap = 4
	}
	var ents []route.Entity
	y := 0
	for gap := 2; gap <= maxGap; gap++ {
		ents = append(ents,
			route.Entity{Name: ug, Kind: route.KindUndergroundIn, X: 0, Y: y, Dir: geom.East, UGType: "input"},
			route.Entity{Name: ug, Kind: route.KindUndergroundOut, X: gap, Y: y, Dir: geom.East, UGType: "output"},
		)
		y += 2
	}
	bp, err := FromRoute(ents, "verify-ug-"+ug)
	if err != nil {
		return VerifyKit{}, err
	}
	return VerifyKit{Name: "ug-" + ug, Entities: ents, Blueprint: bp,
		Description: fmt.Sprintf("underground gaps 2..%d", maxGap)}, nil
}
