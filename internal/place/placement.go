// Package place positions machines inside a bounding box using simulated annealing.
package place

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
)

// PlacedMachine is a machine at a concrete tile position.
type PlacedMachine struct {
	ID  int
	X   int
	Y   int
	Rot int // 0..3 quarter turns clockwise
}

// Placement is the result of the placement stage.
type Placement struct {
	W        int
	H        int
	Machines []PlacedMachine
	Cost     float64
}

// Model holds the mutable state for annealing.
type Model struct {
	plan *plan.Plan
	w, h int

	mw, mh []int // machine widths / heights
	px, py []int // current positions
	rot    []int

	occ []int // 0 free, -1 obstacle, id+1 occupied by machine

	netTerms []NetTerm // weighted producer/consumer groups for HPWL
	bins     *CongestionBins
	linked   [][2]int

	rng *rand.Rand
}

// NewModel builds a placement model for the given plan and box size.
func NewModel(p *plan.Plan, w, h int) *Model {
	n := len(p.Machines)
	m := &Model{
		plan:     p,
		w:        w,
		h:        h,
		mw:       make([]int, n),
		mh:       make([]int, n),
		px:       make([]int, n),
		py:       make([]int, n),
		rot:      make([]int, n),
		occ:      make([]int, w*h),
		netTerms: BuildNetTerms(p),
		bins:     NewCongestionBins(w, h, 4),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	m.linked = LinkedPairs(m.netTerms)
	for i, mi := range p.Machines {
		m.mw[i] = mi.W
		m.mh[i] = mi.H
		if m.mw[i] < 1 {
			m.mw[i] = 1
		}
		if m.mh[i] < 1 {
			m.mh[i] = 1
		}
	}
	return m
}

func (m *Model) idx(x, y int) int { return y*m.w + x }

func (m *Model) clearOcc() {
	for i := range m.occ {
		m.occ[i] = 0
	}
}

func (m *Model) canPlace(id, x, y int) bool {
	w, h := m.mw[id], m.mh[id]
	if x < 0 || y < 0 || x+w > m.w || y+h > m.h {
		return false
	}
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			v := m.occ[m.idx(x+dx, y+dy)]
			if v != 0 && v != id+1 {
				return false
			}
		}
	}
	return true
}

func (m *Model) stamp(id, x, y int, mark int) {
	w, h := m.mw[id], m.mh[id]
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			m.occ[m.idx(x+dx, y+dy)] = mark
		}
	}
}

// RandomInitial places every machine at a random non-overlapping position.
func (m *Model) RandomInitial() bool {
	m.clearOcc()
	n := len(m.px)
	order := m.rng.Perm(n)
	for _, id := range order {
		placed := false
		for try := 0; try < 2000; try++ {
			x := m.rng.Intn(m.w - m.mw[id] + 1)
			y := m.rng.Intn(m.h - m.mh[id] + 1)
			if m.canPlace(id, x, y) {
				m.px[id], m.py[id] = x, y
				m.stamp(id, x, y, id+1)
				placed = true
				break
			}
		}
		if !placed {
			return false
		}
	}
	return true
}


// SeedFromClusters repositions machines that belong to DI clusters so they sit
// at fixed relative offsets. Called after RandomInitial; failures leave the
// random placement for those machines.
func (m *Model) SeedFromClusters(clusters []Cluster) {
	if len(clusters) == 0 {
		return
	}
	for _, c := range clusters {
		if len(c.Members) == 0 {
			continue
		}
		// Compute bounding size of cluster
		maxX, maxY := 0, 0
		for j := range c.Members {
			id := c.Members[j]
			if id < 0 || id >= len(m.mw) {
				continue
			}
			ex := c.DX[j] + m.mw[id]
			ey := c.DY[j] + m.mh[id]
			if ex > maxX {
				maxX = ex
			}
			if ey > maxY {
				maxY = ey
			}
		}
		if maxX > m.w || maxY > m.h {
			continue
		}
		// Unstamp members
		for _, id := range c.Members {
			if id >= 0 && id < len(m.px) {
				m.stamp(id, m.px[id], m.py[id], 0)
			}
		}
		placed := false
		for try := 0; try < 500; try++ {
			ox := m.rng.Intn(m.w - maxX + 1)
			oy := m.rng.Intn(m.h - maxY + 1)
			ok := true
			for j, id := range c.Members {
				if id < 0 || id >= len(m.mw) {
					ok = false
					break
				}
				x, y := ox+c.DX[j], oy+c.DY[j]
				if !m.canPlace(id, x, y) {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			for j, id := range c.Members {
				x, y := ox+c.DX[j], oy+c.DY[j]
				m.px[id], m.py[id] = x, y
				m.stamp(id, x, y, id+1)
			}
			placed = true
			break
		}
		if !placed {
			// restore random positions
			for _, id := range c.Members {
				if id < 0 || id >= len(m.px) {
					continue
				}
				for try := 0; try < 200; try++ {
					x := m.rng.Intn(m.w - m.mw[id] + 1)
					y := m.rng.Intn(m.h - m.mh[id] + 1)
					if m.canPlace(id, x, y) {
						m.px[id], m.py[id] = x, y
						m.stamp(id, x, y, id+1)
						break
					}
				}
			}
		}
	}
}

// Cost is a placement objective combining:
//   - spread from centre of mass
//   - clustering of same-recipe machines
//   - affinity between producer and consumer recipes (shared items)
func (m *Model) Cost() float64 {
	n := len(m.px)
	if n == 0 {
		return 0
	}

	type pt struct{ x, y float64 }
	centres := make([]pt, n)
	var cx, cy float64
	for i := range m.px {
		centres[i] = pt{
			x: float64(m.px[i]) + float64(m.mw[i])/2,
			y: float64(m.py[i]) + float64(m.mh[i])/2,
		}
		cx += centres[i].x
		cy += centres[i].y
	}
	cx /= float64(n)
	cy /= float64(n)

	var spread float64
	for _, c := range centres {
		spread += math.Hypot(c.x-cx, c.y-cy)
	}

	// Same-recipe cluster
	byRecipe := map[string][]int{}
	for i, mi := range m.plan.Machines {
		byRecipe[mi.Recipe] = append(byRecipe[mi.Recipe], i)
	}
	var cluster float64
	for _, ids := range byRecipe {
		for a := 0; a < len(ids); a++ {
			for b := a + 1; b < len(ids); b++ {
				i, j := ids[a], ids[b]
				cluster += math.Hypot(centres[i].x-centres[j].x, centres[i].y-centres[j].y)
			}
		}
	}

	// Producer–consumer affinity: pull machines that share items (product of A = ingredient of B).
	var affinity float64
	for i := 0; i < n; i++ {
		prods := map[string]bool{}
		for _, pr := range m.plan.Machines[i].Products {
			prods[pr] = true
		}
		ings := map[string]bool{}
		for _, ing := range m.plan.Machines[i].Ingredients {
			ings[ing] = true
		}
		for j := i + 1; j < n; j++ {
			linked := false
			for _, pr := range m.plan.Machines[j].Products {
				if ings[pr] {
					linked = true
					break
				}
			}
			if !linked {
				for _, ing := range m.plan.Machines[j].Ingredients {
					if prods[ing] {
						linked = true
						break
					}
				}
			}
			if linked {
				d := math.Hypot(centres[i].x-centres[j].x, centres[i].y-centres[j].y)
				affinity += d * 0.4
			}
		}
	}

	// Spacing: penalize machines closer than 1 tile (no corridor for belts/inserters)
	var spacing float64
	for i := 0; i < n; i++ {
		ri := geom.NewRect(m.px[i], m.py[i], m.mw[i], m.mh[i])
		for j := i + 1; j < n; j++ {
			rj := geom.NewRect(m.px[j], m.py[j], m.mw[j], m.mh[j])
			// gap between rects
			gapX := 0
			if rj.Min.X >= ri.Max.X {
				gapX = rj.Min.X - ri.Max.X
			} else if ri.Min.X >= rj.Max.X {
				gapX = ri.Min.X - rj.Max.X
			}
			gapY := 0
			if rj.Min.Y >= ri.Max.Y {
				gapY = rj.Min.Y - ri.Max.Y
			} else if ri.Min.Y >= rj.Max.Y {
				gapY = ri.Min.Y - rj.Max.Y
			}
			// if they share a row/col band and gap is 0, they're adjacent
			overlapY := ri.Min.Y < rj.Max.Y && rj.Min.Y < ri.Max.Y
			overlapX := ri.Min.X < rj.Max.X && rj.Min.X < ri.Max.X
			if overlapY && gapX == 0 {
				spacing += 4
			}
			if overlapX && gapY == 0 {
				spacing += 4
			}
			if overlapY && gapX == 1 {
				spacing += 1
			}
			if overlapX && gapY == 1 {
				spacing += 1
			}
		}
	}
	hpwl := m.HPWL(m.netTerms)
	if m.bins != nil {
		m.bins.Rebuild(m.px, m.py, m.mw, m.mh)
	}
	binPen := 0.0
	if m.bins != nil {
		binPen = m.bins.Penalty(SoftCapDefault(m.bins.Cell))
	}
	di := DICredit(m.px, m.py, m.mw, m.mh, m.linked, 1)
	nr := NetRemovalCredit(m.px, m.py, m.mw, m.mh, m.netTerms, 2.0)
	return spread + 0.35*cluster + affinity + 0.5*spacing + 0.15*hpwl + 2.0*binPen - di - nr

}

// Anneal runs simulated annealing and returns the best placement found.
func (m *Model) Anneal(iterations int, initialTemp, cooling float64) Placement {
	if iterations <= 0 {
		iterations = 20000
	}
	if initialTemp <= 0 {
		initialTemp = 500
	}
	if cooling <= 0 || cooling >= 1 {
		cooling = 0.997
	}

	if !m.RandomInitial() {
		// Try a slightly larger box by failing gracefully – caller may expand.
		return Placement{W: m.w, H: m.h, Cost: math.Inf(1)}
	}
	m.SeedFromClusters(DIClusters(m.plan, false))

	bestCost := m.Cost()
	bestPX := append([]int(nil), m.px...)
	bestPY := append([]int(nil), m.py...)
	bestRot := append([]int(nil), m.rot...)

	curCost := bestCost
	temp := initialTemp
	n := len(m.px)

	for it := 0; it < iterations; it++ {
		// Pick a random machine and a random new position.
		id := m.rng.Intn(n)
		ox, oy := m.px[id], m.py[id]

		nx := m.rng.Intn(m.w - m.mw[id] + 1)
		ny := m.rng.Intn(m.h - m.mh[id] + 1)
		if nx == ox && ny == oy {
			continue
		}

		// Unstamp, try place, restamp.
		m.stamp(id, ox, oy, 0)
		if !m.canPlace(id, nx, ny) {
			m.stamp(id, ox, oy, id+1)
			continue
		}
		m.px[id], m.py[id] = nx, ny
		m.stamp(id, nx, ny, id+1)

		newCost := m.Cost()
		delta := newCost - curCost
		accept := delta < 0 || m.rng.Float64() < math.Exp(-delta/temp)
		if accept {
			curCost = newCost
			if newCost < bestCost {
				bestCost = newCost
				copy(bestPX, m.px)
				copy(bestPY, m.py)
				copy(bestRot, m.rot)
			}
		} else {
			// Revert.
			m.stamp(id, nx, ny, 0)
			m.px[id], m.py[id] = ox, oy
			m.stamp(id, ox, oy, id+1)
		}

		temp *= cooling
		if temp < 1e-6 {
			temp = 1e-6
		}
	}

	// Restore best.
	m.clearOcc()
	copy(m.px, bestPX)
	copy(m.py, bestPY)
	copy(m.rot, bestRot)
	for i := range m.px {
		m.stamp(i, m.px[i], m.py[i], i+1)
	}

	machines := make([]PlacedMachine, n)
	for i := range machines {
		machines[i] = PlacedMachine{ID: i, X: m.px[i], Y: m.py[i], Rot: m.rot[i]}
	}
	return Placement{W: m.w, H: m.h, Machines: machines, Cost: bestCost}
}

// Place runs placement: starts from an estimated box and grows if packing fails.
func Place(p *plan.Plan, areaFactor float64, iterations int, initialTemp, cooling float64) (Placement, error) {
	if len(p.Machines) == 0 {
		return Placement{}, fmt.Errorf("no machines to place")
	}
	w, h := p.BoundingAreaEstimate(areaFactor)
	if w < 10 {
		w = 10
	}
	if h < 10 {
		h = 10
	}

	for attempt := 0; attempt < 8; attempt++ {
		model := NewModel(p, w, h)
		pl := model.Anneal(iterations, initialTemp, cooling)
		if !math.IsInf(pl.Cost, 1) && len(pl.Machines) == len(p.Machines) {
			return pl, nil
		}
		// Grow box.
		w = int(float64(w) * 1.15)
		h = int(float64(h) * 1.15)
	}
	return Placement{}, fmt.Errorf("could not pack %d machines after growing the box", len(p.Machines))
}

// ASCII renders a simple text view of the placement (useful for small layouts).
func (pl Placement) ASCII(p *plan.Plan) string {
	if pl.W*pl.H > 80*40 {
		return fmt.Sprintf("(box %dx%d too large for ASCII preview)", pl.W, pl.H)
	}
	grid := make([][]byte, pl.H)
	for y := 0; y < pl.H; y++ {
		grid[y] = make([]byte, pl.W)
		for x := 0; x < pl.W; x++ {
			grid[y][x] = '.'
		}
	}
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for i, m := range pl.Machines {
		ch := byte('#')
		if i < len(letters) {
			ch = letters[i]
		}
		mi := p.Machines[m.ID]
		for dy := 0; dy < mi.H; dy++ {
			for dx := 0; dx < mi.W; dx++ {
				x, y := m.X+dx, m.Y+dy
				if y >= 0 && y < pl.H && x >= 0 && x < pl.W {
					grid[y][x] = ch
				}
			}
		}
	}
	var b strings.Builder
	for y := 0; y < pl.H; y++ {
		b.Write(grid[y])
		b.WriteByte('\n')
	}
	return b.String()
}

// Rect returns the axis-aligned rect of a placed machine.
func (pl Placement) Rect(p *plan.Plan, m PlacedMachine) geom.Rect {
	mi := p.Machines[m.ID]
	return geom.Rect{
		Min: geom.Point{X: m.X, Y: m.Y},
		Max: geom.Point{X: m.X + mi.W, Y: m.Y + mi.H},
	}
}
