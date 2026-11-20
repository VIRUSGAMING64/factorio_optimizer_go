package plan

import (
	"math"
	"sort"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
)

// SimplexSolver is a minimal pure-Go two-phase simplex for small LPs
// of the form min c·x s.t. A x >= b, x >= 0.
// Intended only for the recipe-rate refinement (dozens of variables).
type SimplexSolver struct {
	// Internal tableau is built on demand.
}

// SolveMin returns an optimal x for min c·x s.t. A·x >= b, x >= 0.
// Returns nil if infeasible or numerical failure (caller should fall back).
func (s *SimplexSolver) SolveMin(c []float64, A [][]float64, b []float64) []float64 {
	m := len(A)
	n := len(c)
	if m == 0 || n == 0 || len(b) != m {
		return nil
	}
	for _, row := range A {
		if len(row) != n {
			return nil
		}
	}

	return s.solveBigM(c, A, b)
}

// solveBigM implements a rudimentary Big-M method.
// It is intentionally conservative and returns nil on any pivot failure.
func (s *SimplexSolver) solveBigM(c []float64, A [][]float64, b []float64) []float64 {
	const bigM = 1e6
	const eps = 1e-8
	m := len(A)
	n := len(c)

	// Variables: x[0..n-1], surplus s[0..m-1], artificial a[0..m-1]
	totalVars := n + 2*m
	tab := make([][]float64, m+1)
	for i := range tab {
		tab[i] = make([]float64, totalVars+1)
	}

	for i := 0; i < m; i++ {
		if b[i] < -eps {
			for j := 0; j < n; j++ {
				tab[i][j] = -A[i][j]
			}
			tab[i][n+i] = 1
			tab[i][totalVars] = -b[i]
		} else {
			for j := 0; j < n; j++ {
				tab[i][j] = A[i][j]
			}
			tab[i][n+i] = -1  // surplus
			tab[i][n+m+i] = 1 // artificial
			tab[i][totalVars] = b[i]
		}
	}

	for j := 0; j < n; j++ {
		tab[m][j] = -c[j]
	}
	for j := 0; j < m; j++ {
		tab[m][n+m+j] = -bigM
	}

	for i := 0; i < m; i++ {
		artCol := n + m + i
		if math.Abs(tab[i][artCol]) > eps {
			factor := tab[m][artCol] / tab[i][artCol]
			for j := 0; j <= totalVars; j++ {
				tab[m][j] -= factor * tab[i][j]
			}
		}
	}

	const maxIter = 200
	basic := make([]int, m)
	for i := 0; i < m; i++ {
		basic[i] = n + m + i
	}

	for iter := 0; iter < maxIter; iter++ {
		enter := -1
		best := eps
		for j := 0; j < totalVars; j++ {
			if tab[m][j] > best {
				best = tab[m][j]
				enter = j
			}
		}
		if enter < 0 {
			break
		}

		leave := -1
		minRatio := math.Inf(1)
		for i := 0; i < m; i++ {
			if tab[i][enter] > eps {
				ratio := tab[i][totalVars] / tab[i][enter]
				if ratio < minRatio {
					minRatio = ratio
					leave = i
				}
			}
		}
		if leave < 0 {
			return nil
		}

		piv := tab[leave][enter]
		for j := 0; j <= totalVars; j++ {
			tab[leave][j] /= piv
		}
		for i := 0; i <= m; i++ {
			if i == leave {
				continue
			}
			f := tab[i][enter]
			if math.Abs(f) < eps {
				continue
			}
			for j := 0; j <= totalVars; j++ {
				tab[i][j] -= f * tab[leave][j]
			}
		}
		basic[leave] = enter
	}

	x := make([]float64, n)
	for i := 0; i < m; i++ {
		bv := basic[i]
		if bv < n {
			x[bv] = tab[i][totalVars]
		}
	}
	for i := 0; i < m; i++ {
		if basic[i] >= n+m && tab[i][totalVars] > 1e-4 {
			return nil
		}
	}
	return x
}

// RefineRatesWithLP attempts to re-solve crafts for a fixed set of recipes
// using the pure-Go simplex, minimising total machine area subject to
// meeting all targets. Returns the adjusted crafts map or nil on failure.
func RefineRatesWithLP(root *data.RootData, cfg *config.RunConfiguration, selected map[string]float64, targets map[string]float64) map[string]float64 {
	if len(selected) == 0 || len(selected) > 40 {
		return nil
	}

	recipes := make([]string, 0, len(selected))
	for r := range selected {
		recipes = append(recipes, r)
	}
	sort.Strings(recipes)

	items := map[string]struct{}{}
	for _, rn := range recipes {
		r := root.Recipes[rn]
		for _, p := range r.Products {
			items[p.Name] = struct{}{}
		}
		for _, ing := range r.Ingredients {
			items[ing.Name] = struct{}{}
		}
	}
	for it := range targets {
		items[it] = struct{}{}
	}
	itemList := make([]string, 0, len(items))
	for it := range items {
		itemList = append(itemList, it)
	}
	sort.Strings(itemList)
	itemIdx := map[string]int{}
	for i, it := range itemList {
		itemIdx[it] = i
	}

	n := len(recipes)
	m := len(itemList)
	A := make([][]float64, m)
	b := make([]float64, m)
	c := make([]float64, n)
	for i := range A {
		A[i] = make([]float64, n)
	}

	for j, rn := range recipes {
		r := root.Recipes[rn]
		entName := pickMachine(root, cfg, r)
		fp := data.Footprint(entName, root.Entities)
		area := float64(fp.W * fp.H)
		if area <= 0 {
			area = 1
		}
		c[j] = area

		for _, p := range r.Products {
			if idx, ok := itemIdx[p.Name]; ok {
				A[idx][j] += p.ExpectedAmount()
			}
		}
		for _, ing := range r.Ingredients {
			if idx, ok := itemIdx[ing.Name]; ok {
				A[idx][j] -= ing.Amount
			}
		}
	}

	for i, it := range itemList {
		if data.IsRaw(it) {
			b[i] = -1e6
			continue
		}
		b[i] = targets[it]
	}

	solver := &SimplexSolver{}
	x := solver.SolveMin(c, A, b)
	if x == nil {
		return nil
	}

	out := make(map[string]float64, n)
	for j, rn := range recipes {
		if x[j] > 1e-8 {
			out[rn] = x[j]
		}
	}
	return out
}
