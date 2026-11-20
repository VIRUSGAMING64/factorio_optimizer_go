package route

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"

type aNode struct {
	p    geom.Point
	g, f float64
	prev int
}

type minHeap []aNode

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].f < h[j].f }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(aNode)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
