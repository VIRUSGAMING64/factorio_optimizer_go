package place

import "math"

// CongestionBins divides the box into cells and penalises overcrowding
// (too many machine centres in one bin) — PlaceModel bin term.
type CongestionBins struct {
	Cell   int
	Cols   int
	Rows   int
	Counts []int
}

// NewCongestionBins builds empty bins for a box.
func NewCongestionBins(w, h, cell int) *CongestionBins {
	if cell < 2 {
		cell = 4
	}
	cols := (w + cell - 1) / cell
	rows := (h + cell - 1) / cell
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	return &CongestionBins{Cell: cell, Cols: cols, Rows: rows, Counts: make([]int, cols*rows)}
}

func (b *CongestionBins) idx(x, y int) int {
	cx, cy := x/b.Cell, y/b.Cell
	if cx >= b.Cols {
		cx = b.Cols - 1
	}
	if cy >= b.Rows {
		cy = b.Rows - 1
	}
	if cx < 0 {
		cx = 0
	}
	if cy < 0 {
		cy = 0
	}
	return cy*b.Cols + cx
}

// Rebuild recounts machine centres.
func (b *CongestionBins) Rebuild(px, py, mw, mh []int) {
	for i := range b.Counts {
		b.Counts[i] = 0
	}
	for i := range px {
		cx := px[i] + mw[i]/2
		cy := py[i] + mh[i]/2
		b.Counts[b.idx(cx, cy)]++
	}
}

// Penalty is sum of max(0, count - softCap)^2.
func (b *CongestionBins) Penalty(softCap int) float64 {
	if softCap < 1 {
		softCap = 2
	}
	var p float64
	for _, c := range b.Counts {
		if c > softCap {
			d := float64(c - softCap)
			p += d * d
		}
	}
	return p
}

// DICredit estimates direct-insertion savings: pairs of linked machines
// within inserter reach of each other reduce net routing need.
func DICredit(px, py, mw, mh []int, linked [][2]int, reach int) float64 {
	if reach < 1 {
		reach = 1
	}
	var credit float64
	for _, pair := range linked {
		i, j := pair[0], pair[1]
		if i < 0 || j < 0 || i >= len(px) || j >= len(px) {
			continue
		}
		// Chebyshev distance between footprints
		ax0, ay0 := px[i], py[i]
		ax1, ay1 := px[i]+mw[i], py[i]+mh[i]
		bx0, by0 := px[j], py[j]
		bx1, by1 := px[j]+mw[j], py[j]+mh[j]
		dx := 0
		if ax1 < bx0 {
			dx = bx0 - ax1
		} else if bx1 < ax0 {
			dx = ax0 - bx1
		}
		dy := 0
		if ay1 < by0 {
			dy = by0 - ay1
		} else if by1 < ay0 {
			dy = ay0 - by1
		}
		gap := dx
		if dy > gap {
			gap = dy
		}
		if gap <= reach {
			credit += 8 // strong credit for true DI
		} else if gap <= reach+1 {
			credit += 3 // long-handed
		}
	}
	return credit
}

// LinkedPairs from net terms: each producer-consumer pair of machine indices.
func LinkedPairs(terms []NetTerm) [][2]int {
	seen := map[[2]int]bool{}
	var out [][2]int
	for _, t := range terms {
		for _, p := range t.Producers {
			for _, c := range t.Consumers {
				a, b := p, c
				if a > b {
					a, b = b, a
				}
				key := [2]int{a, b}
				if !seen[key] {
					seen[key] = true
					out = append(out, key)
				}
			}
		}
	}
	return out
}

// SoftCapDefault scales with cell size.
func SoftCapDefault(cell int) int {
	return int(math.Max(2, float64(cell)/2))
}
