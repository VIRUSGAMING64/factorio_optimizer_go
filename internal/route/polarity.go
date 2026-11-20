// Package route – lane polarity helpers.
//
// PreferredLane decides which belt lane (left/right relative to travel)
// an inserter is more likely to interact with, given the direction the
// belt is travelling and the side from which the inserter approaches.
package route

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"

// PreferredLane returns the lane that lies on the approach side of a belt.
//
// Convention (Factorio): when facing the travel direction of the belt,
// the left side is LaneLeft and the right side is LaneRight.
// If the inserter approaches from the left side of that facing, it prefers LaneLeft.
func PreferredLane(beltDir, approachSide geom.Direction) Lane {
	// Left side when facing beltDir is the counter-clockwise direction.
	leftSide := beltDir.CCW()
	if approachSide == leftSide {
		return LaneLeft
	}
	if approachSide == leftSide.Opposite() {
		return LaneRight
	}
	// Approach is parallel to travel or undefined – no strong preference.
	return LaneLeft
}

// ApproachSide derives the cardinal direction from a belt tile toward the
// inserter tile of an AccessOption. Used as the approachSide argument to
// PreferredLane.
func ApproachSide(ao AccessOption) geom.Direction {
	dx := ao.InsTile.X - ao.BeltTile.X
	dy := ao.InsTile.Y - ao.BeltTile.Y
	if absInt(dx) >= absInt(dy) {
		if dx > 0 {
			return geom.East
		}
		return geom.West
	}
	if dy > 0 {
		return geom.South
	}
	return geom.North
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
