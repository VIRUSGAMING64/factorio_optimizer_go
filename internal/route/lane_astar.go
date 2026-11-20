package route

import (
	"container/heap"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

// LaneNode is a search state with explicit belt lane and heading.
type LaneNode struct {
	X, Y int
	Dir  geom.Direction
	Lane Lane
}

func (n LaneNode) key(w int) int {
	return ((((n.Y*w + n.X) << 2) | int(n.Dir)) << 1) | int(n.Lane)
}

type laneItem struct {
	node  LaneNode
	g, f  float64
	index int
}

type lanePQ []*laneItem

func (h lanePQ) Len() int            { return len(h) }
func (h lanePQ) Less(i, j int) bool  { return h[i].f < h[j].f }
func (h lanePQ) Swap(i, j int)       { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *lanePQ) Push(x interface{}) { *h = append(*h, x.(*laneItem)) }
func (h *lanePQ) Pop() interface{} {
	old := *h
	n := len(old)
	it := old[n-1]
	*h = old[:n-1]
	return it
}

// FindPathLaneAStar runs A* over (tile, direction, lane).
// Lane switches and turns are penalised; same-lane straight runs are preferred.
// Residual capacity is not considered; see FindPathLaneAStarCapacity.
func (pf *PathFinder) FindPathLaneAStar(src, dst geom.Point, netID int, startLane Lane) ([]geom.Point, []Lane) {
	return pf.FindPathLaneAStarCapacity(src, dst, netID, startLane, nil)
}

// FindPathLaneAStarCapacity is the capacity-aware variant of FindPathLaneAStar.
// When lt is non-nil the step cost is increased for lanes that already have low
// residual capacity, so the search prefers freer lanes of the same net.
func (pf *PathFinder) FindPathLaneAStarCapacity(src, dst geom.Point, netID int, startLane Lane, lt *LaneTracker) ([]geom.Point, []Lane) {
	src = pf.nearestFree(src, netID)
	dst = pf.nearestFree(dst, netID)
	if src.X < 0 || dst.X < 0 {
		return nil, nil
	}
	start := LaneNode{X: src.X, Y: src.Y, Dir: geom.East, Lane: startLane}
	pq := &lanePQ{}
	heap.Init(pq)
	heap.Push(pq, &laneItem{node: start, g: 0, f: float64(src.Manhattan(dst))})
	gscore := map[int]float64{start.key(pf.W): 0}
	parent := map[int]LaneNode{}
	parentKey := map[int]int{}
	closed := map[int]bool{}

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*laneItem)
		ck := cur.node.key(pf.W)
		if closed[ck] {
			continue
		}
		closed[ck] = true
		if cur.node.X == dst.X && cur.node.Y == dst.Y {
			path := []geom.Point{{cur.node.X, cur.node.Y}}
			lanes := []Lane{cur.node.Lane}
			k := ck
			for {
				pk, ok := parentKey[k]
				if !ok {
					break
				}
				prev := parent[k]
				path = append(path, geom.Point{prev.X, prev.Y})
				lanes = append(lanes, prev.Lane)
				k = pk
			}
			for a, b := 0, len(path)-1; a < b; a, b = a+1, b-1 {
				path[a], path[b] = path[b], path[a]
				lanes[a], lanes[b] = lanes[b], lanes[a]
			}
			return path, lanes
		}
		for _, d := range geom.AllDirections {
			nx := cur.node.X + d.Delta().X
			ny := cur.node.Y + d.Delta().Y
			if !pf.Walkable(nx, ny, netID) {
				continue
			}
			idx := pf.idx(nx, ny)
			for _, lane := range []Lane{LaneLeft, LaneRight} {
				nn := LaneNode{X: nx, Y: ny, Dir: d, Lane: lane}
				nk := nn.key(pf.W)
				if closed[nk] {
					continue
				}
				step := pf.TileCost(nx, ny, netID) + pf.ChanCost(cur.node.X, cur.node.Y, d)
				if d != cur.node.Dir {
					step += 0.35
				}
				if lane != cur.node.Lane {
					step += 0.55
				}
				if lane == cur.node.Lane && d == cur.node.Dir {
					step *= 0.92
				}
				// Prefer lanes with residual capacity when a tracker is supplied.
				if lt != nil {
					rem := lt.RemainingLane(idx, lane)
					cap := lt.CapPerLane()
					if cap > 0 {
						// Penalty grows as residual fraction drops (0..~1.5).
						usedFrac := 1.0 - rem/cap
						if usedFrac < 0 {
							usedFrac = 0
						}
						step += usedFrac * 1.2
					}
				}
				ng := cur.g + step
				if old, ok := gscore[nk]; ok && ng >= old {
					continue
				}
				gscore[nk] = ng
				parent[nk] = cur.node
				parentKey[nk] = ck
				h := float64(absI(nx-dst.X) + absI(ny-dst.Y))
				heap.Push(pq, &laneItem{node: nn, g: ng, f: ng + h})
			}
		}
	}
	return pf.FindPathWithLanes(src, dst, netID, startLane)
}

func absI(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
