package blocks

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
)

// InserterNeed estimates how many inserters a machine side needs for a given rate
// against one belt (items/s). Uses a conservative 7.5 items/s per fast inserter.
func InserterNeed(rate float64, inserterThroughput float64) int {
	if inserterThroughput < 1 {
		inserterThroughput = 7.5
	}
	if rate <= 0 {
		return 1
	}
	n := int(rate/inserterThroughput + 0.999)
	if n < 1 {
		n = 1
	}
	if n > 4 {
		n = 4 // physical limit along a 3-tile side with poles
	}
	return n
}

// TemplateWithInserters builds a narrow column and returns machine placement plus
// suggested inserter counts per machine index.
func TemplateWithInserters(b Block, beltCapacity float64) (place.Placement, []int, bool) {
	tr, ok := TryBestTemplate(b)
	if !ok {
		return place.Placement{}, nil, false
	}
	if beltCapacity < 1 {
		beltCapacity = 15 // fast belt
	}
	needs := make([]int, len(b.Machines))
	for i, mi := range b.Machines {
		// Approximate output rate as crafts (caller should have real rates)
		rate := mi.Crafts
		if rate < 0.1 {
			rate = 0.5
		}
		// Cap by belt so we don't request more inserters than the belt can take
		if rate > beltCapacity {
			rate = beltCapacity
		}
		needs[i] = InserterNeed(rate, 7.5)
	}
	return tr.Placement, needs, true
}
