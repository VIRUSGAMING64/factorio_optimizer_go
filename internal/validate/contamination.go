package validate

import "github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"

// Contamination reports belts that carry more than one distinct item (or empty
// mixed with a named item), which Factorio would treat as blocked.
type ContaminationReport struct {
	OK       bool
	Conflicts []string
}

func CheckContamination(ents []route.Entity) ContaminationReport {
	// Map tile → set of items on belts/UG at that tile
	type key struct{ x, y int }
	items := map[key]map[string]bool{}
	for _, e := range ents {
		if e.Kind != route.KindBelt && e.Kind != route.KindUndergroundIn && e.Kind != route.KindUndergroundOut && e.Kind != route.KindSplitter {
			continue
		}
		if e.Item == "" {
			continue
		}
		k := key{e.X, e.Y}
		if items[k] == nil {
			items[k] = map[string]bool{}
		}
		items[k][e.Item] = true
	}
	rep := ContaminationReport{OK: true}
	for k, set := range items {
		if len(set) > 1 {
			rep.OK = false
			names := ""
			for n := range set {
				if names != "" {
					names += ","
				}
				names += n
			}
			rep.Conflicts = append(rep.Conflicts, names+" @ "+itoa(k.x)+","+itoa(k.y))
		}
	}
	return rep
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
