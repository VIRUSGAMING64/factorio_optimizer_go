package route

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// Role of a terminal on a net.
type Role int

const (
	RoleProducer Role = iota
	RoleConsumer
	RolePortIn
	RolePortOut
	RoleSink
)

// AccessOption describes how an inserter connects a machine to a belt tile.
type AccessOption struct {
	MachineID int
	BeltTile  geom.Point
	InsTile   geom.Point
	InsDir    geom.Direction
	Long      bool
	Side      geom.Direction
}

// AccessOptions returns possible inserter+belt placements around a machine rectangle.
// free reports whether a tile may hold an inserter or belt.
func AccessOptions(machineID int, rect geom.Rect, producer bool, free func(geom.Point) bool, longAllowed bool) []AccessOption {
	var out []AccessOption
	for _, side := range geom.AllDirections {
		d1 := rect.SideTiles(side, 1)
		d2 := rect.SideTiles(side, 2)
		// Consumer picks from belt (outward); producer picks from machine (inward).
		insDir := side
		if producer {
			insDir = side.Opposite()
		}
		for i := range d1 {
			if !free(d1[i]) {
				continue
			}
			// Normal reach-1: inserter on d1, belt on d2
			if i < len(d2) && free(d2[i]) {
				out = append(out, AccessOption{
					MachineID: machineID,
					BeltTile:  d2[i],
					InsTile:   d1[i],
					InsDir:    insDir,
					Long:      false,
					Side:      side,
				})
			}
			// Belt on d1 itself (inserter between machine edge and belt sharing tile conceptually –
			// simplified: belt on d1, inserter needs free neighbor – skip dual occupancy)
			out = append(out, AccessOption{
				MachineID: machineID,
				BeltTile:  d1[i],
				InsTile:   d1[i],
				InsDir:    insDir,
				Long:      false,
				Side:      side,
			})
			// Long-handed: inserter on d1, belt on d3
			if longAllowed {
				d3 := rect.SideTiles(side, 3)
				if i < len(d3) && free(d3[i]) {
					out = append(out, AccessOption{
						MachineID: machineID,
						BeltTile:  d3[i],
						InsTile:   d1[i],
						InsDir:    insDir,
						Long:      true,
						Side:      side,
					})
				}
			}
		}
	}
	return out
}

// MachineRect returns the rect of a placed machine.
func MachineRect(p *plan.Plan, plX, plY, id int) geom.Rect {
	mi := p.Machines[id]
	return geom.NewRect(plX, plY, mi.W, mi.H)
}
