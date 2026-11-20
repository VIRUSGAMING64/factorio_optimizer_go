package validate

import (
	"fmt"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// BeltNode is one belt/underground tile in the layout graph.
type BeltNode struct {
	Idx  int
	X, Y int
	Dir  geom.Direction
	Item string
	Kind route.EntityKind
}

// BeltGraph connects belt tiles by adjacency and direction.
type BeltGraph struct {
	Nodes   []BeltNode
	At      map[geom.Point]int // tile → node index
	Succ    [][]int            // outgoing neighbours
	Pred    [][]int
	Items   []map[string]bool  // items that can reach each node
}

// BuildBeltGraph constructs a graph from routed entities.
func BuildBeltGraph(entities []route.Entity) *BeltGraph {
	g := &BeltGraph{At: map[geom.Point]int{}}
	for _, e := range entities {
		if e.Kind != route.KindBelt && e.Kind != route.KindUndergroundIn &&
			e.Kind != route.KindUndergroundOut && e.Kind != route.KindSplitter {
			continue
		}
		idx := len(g.Nodes)
		g.At[geom.Point{e.X, e.Y}] = idx
		g.Nodes = append(g.Nodes, BeltNode{
			Idx: idx, X: e.X, Y: e.Y, Dir: e.Dir, Item: e.Item, Kind: e.Kind,
		})
	}
	n := len(g.Nodes)
	g.Succ = make([][]int, n)
	g.Pred = make([][]int, n)
	g.Items = make([]map[string]bool, n)
	for i := range g.Items {
		g.Items[i] = map[string]bool{}
	}

	for i, node := range g.Nodes {
		// Forward neighbour in belt direction
		d := node.Dir.Delta()
		fwd := geom.Point{node.X + d.X, node.Y + d.Y}
		if j, ok := g.At[fwd]; ok {
			g.Succ[i] = append(g.Succ[i], j)
			g.Pred[j] = append(g.Pred[j], i)
		}
		// Underground: also link in→out pairs of same net roughly by scanning same dir
		if node.Kind == route.KindUndergroundIn {
			for dist := 1; dist <= 10; dist++ {
				t := geom.Point{node.X + d.X*dist, node.Y + d.Y*dist}
				if j, ok := g.At[t]; ok && g.Nodes[j].Kind == route.KindUndergroundOut {
					g.Succ[i] = append(g.Succ[i], j)
					g.Pred[j] = append(g.Pred[j], i)
					break
				}
			}
		}
		// Splitter: fans out to forward and sides
		if node.Kind == route.KindSplitter {
			for _, side := range []geom.Direction{node.Dir, node.Dir.CW(), node.Dir.CCW()} {
				sd := side.Delta()
				t := geom.Point{node.X + sd.X, node.Y + sd.Y}
				if j, ok := g.At[t]; ok {
					g.Succ[i] = append(g.Succ[i], j)
					g.Pred[j] = append(g.Pred[j], i)
				}
			}
		}
	}
	return g
}

// PropagateItems seeds items from entity.Item on each node and BFS-forwards.
func (g *BeltGraph) PropagateItems() {
	queue := make([]int, 0, len(g.Nodes))
	for i, n := range g.Nodes {
		if n.Item != "" {
			g.Items[i][n.Item] = true
			queue = append(queue, i)
		}
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range g.Succ[u] {
			changed := false
			for item := range g.Items[u] {
				if !g.Items[v][item] {
					g.Items[v][item] = true
					changed = true
				}
			}
			if changed {
				queue = append(queue, v)
			}
		}
	}
}

// Reachable reports whether item can reach any node near point p (within chebyshev 1).
func (g *BeltGraph) Reachable(p geom.Point, item string) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			t := geom.Point{p.X + dx, p.Y + dy}
			if i, ok := g.At[t]; ok && g.Items[i][item] {
				return true
			}
		}
	}
	return false
}

// LayoutReport is the result of layout-level belt validation.
type LayoutReport struct {
	OK       bool
	Messages []string
}

func (r LayoutReport) String() string {
	s := "Layout check: "
	if r.OK {
		s += "OK\n"
	} else {
		s += "ISSUES\n"
	}
	for _, m := range r.Messages {
		s += "  - " + m + "\n"
	}
	return s
}

// CheckLayout runs belt-graph validation on routed entities.
func CheckLayout(entities []route.Entity) LayoutReport {
	rep := LayoutReport{OK: true}
	g := BuildBeltGraph(entities)
	if len(g.Nodes) == 0 {
		rep.Messages = append(rep.Messages, "no belt nodes in layout")
		return rep
	}
	g.PropagateItems()

	// Count items present
	itemNodes := map[string]int{}
	orphans := 0
	for i, n := range g.Nodes {
		if n.Item != "" {
			itemNodes[n.Item]++
		}
		if len(g.Items[i]) == 0 {
			orphans++
		}
	}
	rep.Messages = append(rep.Messages,
		fmt.Sprintf("%d belt nodes, %d distinct items seeded, %d nodes without item after propagate",
			len(g.Nodes), len(itemNodes), orphans))

	// Connectivity: each seeded item should reach >1 node if there are paths
	for item, cnt := range itemNodes {
		reached := 0
		for i := range g.Nodes {
			if g.Items[i][item] {
				reached++
			}
		}
		if reached < cnt {
			rep.Messages = append(rep.Messages,
				fmt.Sprintf("item %s: seeded %d nodes, reachable %d", item, cnt, reached))
		}
	}

	// Degree-0 belts (isolated)
	isolated := 0
	for i := range g.Nodes {
		if len(g.Succ[i]) == 0 && len(g.Pred[i]) == 0 {
			isolated++
		}
	}
	if isolated > 0 {
		rep.Messages = append(rep.Messages, fmt.Sprintf("%d isolated belt tiles (no neighbours)", isolated))
	}

	return rep
}
