package blocks

import (
	"math"
	"math/rand"
)

// packSequencePair places rectangles using a simple sequence-pair representation.
//
// A sequence pair (Γ+, Γ-) encodes relative constraints:
//   - if a appears before b in both sequences, a is left of b
//   - if a appears before b in Γ+ and after in Γ-, a is below b
//
// We derive a topological order for x-coordinates from the "left-of" relation
// and for y-coordinates from the "below" relation, then place each block at
// the earliest feasible coordinate. This is a lightweight pure-Go approximation
// of classical sequence-pair floorplanning; it is used as one of the candidate
// packers inside Pack.
func packSequencePair(ws, hs []int, gammaPlus, gammaMinus []int) (x, y []int, maxX, maxY int) {
	n := len(ws)
	if n == 0 {
		return nil, nil, 0, 0
	}
	posPlus := make([]int, n)
	posMinus := make([]int, n)
	for i, id := range gammaPlus {
		if id >= 0 && id < n {
			posPlus[id] = i
		}
	}
	for i, id := range gammaMinus {
		if id >= 0 && id < n {
			posMinus[id] = i
		}
	}

	// Build "must be after" lists for left-of and below relations.
	afterX := make([][]int, n) // blocks that must be to the right of i
	afterY := make([][]int, n) // blocks that must be above i
	for a := 0; a < n; a++ {
		for b := 0; b < n; b++ {
			if a == b {
				continue
			}
			// a left of b  <=>  a before b in both sequences
			if posPlus[a] < posPlus[b] && posMinus[a] < posMinus[b] {
				afterX[a] = append(afterX[a], b)
			}
			// a below b  <=>  a before b in Γ+ and after b in Γ-
			if posPlus[a] < posPlus[b] && posMinus[a] > posMinus[b] {
				afterY[a] = append(afterY[a], b)
			}
		}
	}

	x = make([]int, n)
	y = make([]int, n)

	// Longest-path style earliest placement for x
	place1D(ws, afterX, x)
	// and for y
	place1D(hs, afterY, y)

	maxX, maxY = 0, 0
	for i := 0; i < n; i++ {
		if x[i]+ws[i] > maxX {
			maxX = x[i] + ws[i]
		}
		if y[i]+hs[i] > maxY {
			maxY = y[i] + hs[i]
		}
	}
	return x, y, maxX, maxY
}

// place1D computes the earliest start coordinate for each item given
// predecessor constraints (after[i] lists items that must start after i finishes).
func place1D(sizes []int, after [][]int, out []int) {
	n := len(sizes)
	indeg := make([]int, n)
	for i := 0; i < n; i++ {
		for _, j := range after[i] {
			indeg[j]++
		}
	}
	// Kahn + longest path
	q := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if indeg[i] == 0 {
			q = append(q, i)
		}
	}
	for len(q) > 0 {
		i := q[0]
		q = q[1:]
		for _, j := range after[i] {
			if out[i]+sizes[i] > out[j] {
				out[j] = out[i] + sizes[i]
			}
			indeg[j]--
			if indeg[j] == 0 {
				q = append(q, j)
			}
		}
	}
	// Any remaining cycles (shouldn't happen for a valid SP) fall back to 0
}

// randomSequencePair returns two random permutations of 0..n-1.
func randomSequencePair(rng *rand.Rand, n int) (plus, minus []int) {
	plus = rng.Perm(n)
	minus = rng.Perm(n)
	return plus, minus
}

// ScoreSequencePair is a small helper used by tests and by Pack to evaluate
// a candidate sequence pair without mutating external state.
func ScoreSequencePair(ws, hs []int, plus, minus []int) (area float64, maxX, maxY int) {
	_, _, maxX, maxY = packSequencePair(ws, hs, plus, minus)
	area = float64(maxX * maxY)
	if maxX == 0 || maxY == 0 {
		area = math.Inf(1)
	}
	return area, maxX, maxY
}

// improveSequencePair runs a short local-search (pair swaps + block moves)
// on a sequence pair, returning the best found. Used when the number of blocks
// is modest so that relative-order exploration stays cheap.
func improveSequencePair(ws, hs []int, plus, minus []int, iters int, rng *rand.Rand) (bestPlus, bestMinus []int, bestArea float64) {
	n := len(ws)
	if n == 0 || iters <= 0 {
		return plus, minus, math.Inf(1)
	}
	bestPlus = append([]int(nil), plus...)
	bestMinus = append([]int(nil), minus...)
	bestArea, _, _ = ScoreSequencePair(ws, hs, bestPlus, bestMinus)

	curPlus := append([]int(nil), plus...)
	curMinus := append([]int(nil), minus...)
	curArea := bestArea

	for i := 0; i < iters; i++ {
		// Try a random move: swap two positions in one sequence or move a block.
		np := append([]int(nil), curPlus...)
		nm := append([]int(nil), curMinus...)
		if rng.Float64() < 0.5 {
			// pair swap in Γ+
			a, b := rng.Intn(n), rng.Intn(n)
			np[a], np[b] = np[b], np[a]
		} else {
			// pair swap in Γ-
			a, b := rng.Intn(n), rng.Intn(n)
			nm[a], nm[b] = nm[b], nm[a]
		}
		area, _, _ := ScoreSequencePair(ws, hs, np, nm)
		if area < curArea {
			curPlus, curMinus, curArea = np, nm, area
			if area < bestArea {
				bestPlus = append([]int(nil), np...)
				bestMinus = append([]int(nil), nm...)
				bestArea = area
			}
		}
	}
	return bestPlus, bestMinus, bestArea
}
