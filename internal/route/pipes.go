package route

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// PipeResult is the fluid-routing output (simplified).
type PipeResult struct {
	Entities []Entity
	Pipes    int
	Failed   []int
}

// RoutePipes lays pipes between fluid-producing/consuming machines and fluid ports.
// This is a simplified BFS pipe router; the original Pipes.kt is more complete.
func RoutePipes(p *plan.Plan, pl place.Placement, occ []int, w, h int) PipeResult {
	var ents []Entity
	pipes := 0
	var failed []int

	walkable := func(t geom.Point) bool {
		if t.X < 0 || t.X >= w || t.Y < 0 || t.Y >= h {
			return false
		}
		v := occ[t.Y*w+t.X]
		return v == 0 || v > 0 // free or belt (pipes can share conceptual space for foundation)
	}

	for _, net := range p.Nets {
		if !net.IsFluid {
			continue
		}
		var sources, targets []geom.Point
		posOf := map[int]place.PlacedMachine{}
		for _, m := range pl.Machines {
			posOf[m.ID] = m
		}
		for i, mi := range p.Machines {
			for _, pr := range mi.Products {
				if pr == net.Item {
					m := posOf[i]
					rect := geom.NewRect(m.X, m.Y, mi.W, mi.H)
					// Prefer fluid connection side if known
					conns, _ := data.FluidsOf(mi.Entity)
					if len(conns) > 0 {
						for _, c := range conns {
							if c.Flow == data.FlowOutput {
								side := geom.Direction(c.Side)
								for _, t := range rect.SideTiles(side, 1) {
									if walkable(t) {
										sources = append(sources, t)
									}
								}
							}
						}
					} else {
						for _, side := range geom.AllDirections {
							for _, t := range rect.SideTiles(side, 1) {
								if walkable(t) {
									sources = append(sources, t)
								}
							}
						}
					}
				}
			}
			for _, ing := range mi.Ingredients {
				if ing == net.Item {
					m := posOf[i]
					rect := geom.NewRect(m.X, m.Y, mi.W, mi.H)
					for _, side := range geom.AllDirections {
						for _, t := range rect.SideTiles(side, 1) {
							if walkable(t) {
								targets = append(targets, t)
							}
						}
					}
				}
			}
		}
		for _, port := range p.Ports {
			if port.Item != net.Item {
				continue
			}
			tiles := edgePortTiles(pl, port)
			if port.IsInput {
				sources = append(sources, tiles...)
			} else {
				targets = append(targets, tiles...)
			}
		}
		if len(sources) == 0 || len(targets) == 0 {
			continue
		}
		// BFS from first source to nearest target
		src := sources[0]
		best := []geom.Point(nil)
		for _, tgt := range targets {
			path := bfsPath(src, tgt, walkable, w, h)
			if path != nil && (best == nil || len(path) < len(best)) {
				best = path
			}
		}
		if best == nil {
			failed = append(failed, net.ID)
			continue
		}
		for _, t := range best {
			if t.X < 0 || t.X >= w || t.Y < 0 || t.Y >= h {
				continue
			}
			idx := t.Y*w + t.X
			if occ[idx] == -1 {
				continue
			}
			ents = append(ents, Entity{
				Name: data.Pipe, Kind: KindPipe,
				X: t.X, Y: t.Y, Item: net.Item, NetID: net.ID,
			})
			pipes++
		}
	}
	return PipeResult{Entities: ents, Pipes: pipes, Failed: failed}
}

func bfsPath(src, dst geom.Point, walkable func(geom.Point) bool, w, h int) []geom.Point {
	if !walkable(src) || !walkable(dst) {
		return nil
	}
	type node struct {
		p    geom.Point
		prev int
	}
	queue := []node{{src, -1}}
	came := map[geom.Point]geom.Point{}
	seen := map[geom.Point]bool{src: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.p == dst {
			path := []geom.Point{dst}
			for {
				prev, ok := came[path[len(path)-1]]
				if !ok {
					break
				}
				path = append(path, prev)
			}
			for a, b := 0, len(path)-1; a < b; a, b = a+1, b-1 {
				path[a], path[b] = path[b], path[a]
			}
			return path
		}
		for _, d := range geom.AllDirections {
			np := cur.p.Add(d.Delta())
			if np.X < 0 || np.X >= w || np.Y < 0 || np.Y >= h {
				continue
			}
			if seen[np] || !walkable(np) {
				continue
			}
			seen[np] = true
			came[np] = cur.p
			queue = append(queue, node{np, 0})
		}
	}
	return nil
}
