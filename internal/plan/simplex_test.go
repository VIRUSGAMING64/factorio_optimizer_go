package plan

import (
	"math"
	"testing"
)

func TestSimplexBasic(t *testing.T) {
	// min 2x + 3y
	// s.t. x + y >= 4
	//      2x + y >= 5
	//      x,y >= 0
	// Optimal: x=4, y=0, obj=8
	s := &SimplexSolver{}
	c := []float64{2, 3}
	A := [][]float64{{1, 1}, {2, 1}}
	b := []float64{4, 5}
	x := s.SolveMin(c, A, b)
	if x == nil {
		t.Fatal("expected feasible solution")
	}
	if math.Abs(x[0]-4) > 0.1 || math.Abs(x[1]) > 0.1 {
		t.Fatalf("got %v, want ~[1,3]", x)
	}
}

func TestSimplexInfeasible(t *testing.T) {
	s := &SimplexSolver{}
	// x >= 2, -x >= 1 → impossible
	c := []float64{1}
	A := [][]float64{{1}, {-1}}
	b := []float64{2, 1}
	x := s.SolveMin(c, A, b)
	if x != nil {
		t.Fatal("expected infeasible")
	}
}
