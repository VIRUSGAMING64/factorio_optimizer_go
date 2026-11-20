package place

import "math"

// PinPenalty charges machines that need more inserter slots than free side tiles.
func PinPenalty(px, py, mw, mh []int, needSlots []int) float64 {
	var total float64
	n := len(px)
	for i := 0; i < n; i++ {
		// Free slots ≈ perimeter tiles not blocked by other machines (approx: 2*(w+h))
		free := 2 * (mw[i] + mh[i])
		// Neighbour occupancy reduces free slots roughly by adjacency count
		need := 2
		if i < len(needSlots) && needSlots[i] > 0 {
			need = needSlots[i]
		}
		s := need + 1 - free
		if s < 0 {
			s = 0
		}
		total += float64(s * s)
	}
	return total
}

// NetRemovalCredit awards nets fully covered by direct insertion (producer adjacent to consumer).
func NetRemovalCredit(px, py, mw, mh []int, terms []NetTerm, bonus float64) float64 {
	if bonus <= 0 {
		return 0
	}
	var credit float64
	for _, t := range terms {
		covered := 0
		pairs := 0
		for _, pi := range t.Producers {
			for _, ci := range t.Consumers {
				pairs++
				// bounding boxes touch or are 1-tile apart → DI possible
				ax0, ay0 := px[pi], py[pi]
				ax1, ay1 := ax0+mw[pi], ay0+mh[pi]
				bx0, by0 := px[ci], py[ci]
				bx1, by1 := bx0+mw[ci], by0+mh[ci]
				// gap
				gx := 0
				if bx0 >= ax1 {
					gx = bx0 - ax1
				} else if ax0 >= bx1 {
					gx = ax0 - bx1
				}
				gy := 0
				if by0 >= ay1 {
					gy = by0 - ay1
				} else if ay0 >= by1 {
					gy = ay0 - by1
				}
				if gx <= 1 && gy <= 1 {
					covered++
				}
			}
		}
		if pairs == 0 {
			continue
		}
		share := float64(covered) / float64(pairs)
		if share >= 0.99 {
			credit += bonus * t.Weight
		} else if share > 0 {
			credit += bonus * t.Weight * 0.5 * share * share
		}
	}
	return credit
}

// SoftCap for bin density used by congestion bins.
func SoftCapForBox(w, h, nMachines int) float64 {
	area := float64(w * h)
	if area < 1 {
		return 1
	}
	return math.Max(1, float64(nMachines)*4/area*float64(16))
}
