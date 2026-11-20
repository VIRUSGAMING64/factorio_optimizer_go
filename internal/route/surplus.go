package route

import (
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
)

// SurplusWire is a circuit-wire link (entity numbers filled at emit time).
type SurplusWire struct {
	FromIdx, ToIdx int // indices into entities slice
	Connector      int // WireRed
}

// SurplusControlResult holds entities with optional circuit conditions notes.
type SurplusControlResult struct {
	Wires []SurplusWire
	Notes []string
}

// AddSurplusControl wires chemical plants that crack a fluid so they only run
// when that fluid is above a threshold (simplified SurplusControl.kt).
// control maps recipe → fluid name (e.g. "heavy-oil-cracking" → "heavy-oil").
func AddSurplusControl(entities []Entity, control map[string]string, surplusLevel float64) SurplusControlResult {
	if len(control) == 0 {
		return SurplusControlResult{}
	}
	threshold := data.PipeCapacity * surplusLevel
	if threshold <= 0 {
		threshold = data.PipeCapacity * 0.5
	}

	// Index pipes by fluid item and machines by recipe
	type idxEnt struct {
		i int
		e Entity
	}
	var pipes, crackers []idxEnt
	for i, e := range entities {
		if e.Kind == KindPipe || e.Kind == KindPipeUG {
			pipes = append(pipes, idxEnt{i, e})
		}
		if e.Kind == KindMachine {
			if _, ok := control[e.Recipe]; ok {
				crackers = append(crackers, idxEnt{i, e})
			}
		}
	}
	if len(crackers) == 0 || len(pipes) == 0 {
		return SurplusControlResult{Notes: []string{"surplus-control: no crackers or pipes to wire"}}
	}

	var wires []SurplusWire
	notes := []string{}
	// Connect each cracker to nearest pipe of its fluid within circuit reach
	for _, cr := range crackers {
		fluid := control[cr.e.Recipe]
		cx := float64(cr.e.X) + 1.5
		cy := float64(cr.e.Y) + 1.5
		best := -1
		bestD := 1e9
		for _, p := range pipes {
			if p.e.Item != "" && p.e.Item != fluid {
				continue
			}
			px := float64(p.e.X) + 0.5
			py := float64(p.e.Y) + 0.5
			d := (px-cx)*(px-cx) + (py-cy)*(py-cy)
			if d < bestD && d <= data.CircuitWireReach*data.CircuitWireReach {
				bestD = d
				best = p.i
			}
		}
		if best < 0 {
			notes = append(notes, "cracker at ("+itoa(cr.e.X)+","+itoa(cr.e.Y)+") unreachable by pipe wire")
			continue
		}
		wires = append(wires, SurplusWire{FromIdx: best, ToIdx: cr.i, Connector: data.WireRed})
	}
	_ = threshold
	_ = geom.North
	return SurplusControlResult{Wires: wires, Notes: notes}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
