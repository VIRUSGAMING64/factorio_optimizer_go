package route

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// RingPort is a belt tile on the outer ring of the box used as an input/output port.
type RingPort struct {
	Tile   geom.Point
	Item   string
	IsInput bool
	NetID  int
	Dir    geom.Direction
}

// PlaceRingPorts assigns concrete tiles on the box perimeter for each plan port.
// Inputs prefer west/north; outputs prefer east/south. Spreads ports along the side.
func PlaceRingPorts(p *plan.Plan, pl place.Placement) []RingPort {
	var out []RingPort
	// Group ports by side
	bySide := map[string][]plan.Port{}
	for _, port := range p.Ports {
		side := port.Side
		if side == "" {
			if port.IsInput {
				side = "west"
			} else {
				side = "east"
			}
		}
		bySide[side] = append(bySide[side], port)
	}
	for side, ports := range bySide {
		tiles := sideTiles(pl, side)
		if len(tiles) == 0 {
			continue
		}
		for i, port := range ports {
			idx := (i + 1) * len(tiles) / (len(ports) + 1)
			if idx >= len(tiles) {
				idx = len(tiles) - 1
			}
			if idx < 0 {
				idx = 0
			}
			dir := geom.East
			switch side {
			case "west":
				dir = geom.East
			case "east":
				dir = geom.West
			case "north":
				dir = geom.South
			case "south":
				dir = geom.North
			}
			out = append(out, RingPort{
				Tile: tiles[idx], Item: port.Item, IsInput: port.IsInput,
				NetID: port.ID, Dir: dir,
			})
		}
	}
	return out
}

func sideTiles(pl place.Placement, side string) []geom.Point {
	var out []geom.Point
	switch side {
	case "west":
		for y := 0; y < pl.H; y++ {
			out = append(out, geom.Point{0, y})
		}
	case "east":
		for y := 0; y < pl.H; y++ {
			out = append(out, geom.Point{pl.W - 1, y})
		}
	case "north":
		for x := 0; x < pl.W; x++ {
			out = append(out, geom.Point{x, 0})
		}
	case "south":
		for x := 0; x < pl.W; x++ {
			out = append(out, geom.Point{x, pl.H - 1})
		}
	}
	return out
}

// RingPortEntities turns ring ports into belt entities at the boundary.
func RingPortEntities(ports []RingPort, beltName string) []Entity {
	var out []Entity
	for _, rp := range ports {
		out = append(out, Entity{
			Name: beltName, Kind: KindBelt,
			X: rp.Tile.X, Y: rp.Tile.Y, Dir: rp.Dir,
			Item: rp.Item, NetID: rp.NetID,
		})
	}
	return out
}
