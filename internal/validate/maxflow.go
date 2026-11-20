package validate

import "math"

const mfEPS = 1e-12

// MaxFlow is Dinic's algorithm on a directed graph with real capacities.
type MaxFlow struct {
	n    int
	head []int
	to   []int
	next []int
	cap  []float64
	level []int
	iter  []int
}

// NewMaxFlow creates a flow network with n nodes.
func NewMaxFlow(n int) *MaxFlow {
	head := make([]int, n)
	for i := range head {
		head[i] = -1
	}
	return &MaxFlow{n: n, head: head}
}

// Edge adds u→v with capacity c and returns the forward edge id.
func (m *MaxFlow) Edge(u, v int, c float64) int {
	id := len(m.to)
	m.to = append(m.to, v)
	m.next = append(m.next, m.head[u])
	m.cap = append(m.cap, c)
	m.head[u] = id
	// reverse
	m.to = append(m.to, u)
	m.next = append(m.next, m.head[v])
	m.cap = append(m.cap, 0)
	m.head[v] = id + 1
	return id
}

// Raise adds extra capacity to edge id.
func (m *MaxFlow) Raise(id int, extra float64) {
	m.cap[id] += extra
}

// FlowOn returns flow pushed through forward edge id after Run.
func (m *MaxFlow) FlowOn(id int) float64 {
	return m.cap[id^1]
}

// Run computes max flow from s to t.
func (m *MaxFlow) Run(s, t int) float64 {
	m.level = make([]int, m.n)
	m.iter = make([]int, m.n)
	total := 0.0
	for m.bfs(s, t) {
		copy(m.iter, m.head)
		for {
			f := m.dfs(s, t, math.MaxFloat64)
			if f <= mfEPS {
				break
			}
			total += f
		}
	}
	return total
}

func (m *MaxFlow) bfs(s, t int) bool {
	for i := range m.level {
		m.level[i] = -1
	}
	m.level[s] = 0
	q := []int{s}
	for len(q) > 0 {
		u := q[0]
		q = q[1:]
		for e := m.head[u]; e >= 0; e = m.next[e] {
			if m.cap[e] > mfEPS && m.level[m.to[e]] < 0 {
				m.level[m.to[e]] = m.level[u] + 1
				q = append(q, m.to[e])
			}
		}
	}
	return m.level[t] >= 0
}

func (m *MaxFlow) dfs(u, t int, f float64) float64 {
	if u == t {
		return f
	}
	for m.iter[u] >= 0 {
		e := m.iter[u]
		v := m.to[e]
		if m.cap[e] > mfEPS && m.level[v] == m.level[u]+1 {
			d := m.dfs(v, t, math.Min(f, m.cap[e]))
			if d > mfEPS {
				m.cap[e] -= d
				m.cap[e^1] += d
				return d
			}
		}
		m.iter[u] = m.next[e]
	}
	return 0
}
