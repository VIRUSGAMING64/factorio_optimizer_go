package route

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

// EndpointPair is a source→target demand for one net.
type EndpointPair struct {
	Src, Dst geom.Point
	NetID    int
	Item     string
}

// NegotiateResult is the outcome of multi-iteration negotiated congestion.
type NegotiateResult struct {
	Paths    [][]geom.Point // parallel to pairs (nil if failed)
	Failed   []int          // net IDs that failed
	Iters    int
}

// Negotiate routes all pairs jointly: each iteration routes every pair with
// current history, commits paths, then decays and retries failed ones.
// This is a simplified PathFinder negotiated-congestion loop.
func (pf *PathFinder) Negotiate(pairs []EndpointPair, maxIters int) NegotiateResult {
	if maxIters < 1 {
		maxIters = 3
	}
	paths := make([][]geom.Point, len(pairs))
	var failed []int

	for it := 0; it < maxIters; it++ {
		failed = nil
		// Rip up previous commitments for this negotiate session by resetting occ from machines only
		// (caller should have marked machines). We only clear belt marks (>0).
		if it > 0 {
			pf.ClearBeltMarks()
			pf.DecayHistory()
			pf.PresentBoost(1.35)
		}
		for i, pair := range pairs {
			path := pf.FindPath(pair.Src, pair.Dst, pair.NetID)
			paths[i] = path
			if path == nil {
				failed = append(failed, pair.NetID)
				continue
			}
			pf.CommitPath(path, pair.NetID)
		}
		if len(failed) == 0 {
			return NegotiateResult{Paths: paths, Failed: nil, Iters: it + 1}
		}
	}
	return NegotiateResult{Paths: paths, Failed: uniqueInts(failed), Iters: maxIters}
}
