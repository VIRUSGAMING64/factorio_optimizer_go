// Package emit produces Factorio 2.0 blueprint strings from a routed layout.
package emit

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/geom"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
)

// --- JSON structures matching Factorio 2.0 blueprint format ---

type bpRoot struct {
	Blueprint bpBlueprint `json:"blueprint"`
}

type bpBlueprint struct {
	Item     string     `json:"item"`
	Label    string     `json:"label"`
	Icons    []bpIcon   `json:"icons,omitempty"`
	Entities []bpEntity `json:"entities"`
	Wires    [][]int    `json:"wires,omitempty"`
	Tiles    []bpTile   `json:"tiles,omitempty"`
	Version  int64      `json:"version"`
}

type bpIcon struct {
	Signal bpSignal `json:"signal"`
	Index  int      `json:"index"`
}

type bpSignal struct {
	Name string `json:"name"`
}

type bpEntity struct {
	EntityNumber  int         `json:"entity_number"`
	Name          string      `json:"name"`
	Position      bpPosition  `json:"position"`
	Direction     *int        `json:"direction,omitempty"`
	Type          *string     `json:"type,omitempty"`
	Recipe        *string     `json:"recipe,omitempty"`
	RecipeQuality *string     `json:"recipe_quality,omitempty"`
	Mirror        *bool       `json:"mirror,omitempty"`
	Items         []bpItemReq `json:"items,omitempty"` // modules
	Filters       []bpFilter  `json:"filters,omitempty"`
	FilterMode    *string     `json:"filter_mode,omitempty"`
}

type bpItemReq struct {
	ID     bpSignal `json:"id"`
	Items  map[string]int `json:"items,omitempty"`
}

type bpFilter struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

type bpPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type bpTile struct {
	Name     string        `json:"name"`
	Position bpTilePosition `json:"position"`
}

type bpTilePosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// FromRoute builds a Factorio blueprint string from routed entities.
func FromRoute(entities []route.Entity, label string) (string, error) {
	if label == "" {
		label = "Go optimizer"
	}
	bp := bpBlueprint{
		Item:    "blueprint",
		Label:   label,
		Version: data.BlueprintVersion,
		Icons: []bpIcon{
			{Signal: bpSignal{Name: "assembling-machine-3"}, Index: 1},
		},
	}

	num := 1
	poleNums := []int{}
	for _, e := range entities {
		cx, cy := entityCentre(e)
		ent := bpEntity{
			EntityNumber: num,
			Name:         e.Name,
			Position:     bpPosition{X: round2(cx), Y: round2(cy)},
		}
		if e.Dir != geom.North {
			d := e.Dir.Value16()
			ent.Direction = &d
		}
		if e.Kind == route.KindMachine && e.Recipe != "" && data.TakesRecipe(e.Name) {
			r := e.Recipe
			ent.Recipe = &r
		}
		if e.Kind == route.KindUndergroundIn || e.UGType == "input" {
			t := "input"
			ent.Type = &t
		}
		if e.Kind == route.KindUndergroundOut || e.UGType == "output" {
			t := "output"
			ent.Type = &t
		}
		if e.Kind == route.KindInserter && e.Item != "" {
			ent.Filters = []bpFilter{{Index: 1, Name: e.Item}}
		}
		if e.Kind == route.KindBeacon {
			// attach two speed-module-3 by default
			ent.Items = []bpItemReq{
				{ID: bpSignal{Name: "speed-module-3"}, Items: map[string]int{"speed-module-3": 1}},
				{ID: bpSignal{Name: "speed-module-3"}, Items: map[string]int{"speed-module-3": 1}},
			}
		}
		if e.Kind == route.KindPole {
			poleNums = append(poleNums, num)
		}
		bp.Entities = append(bp.Entities, ent)
		num++
	}

	// Wire neighbouring poles (simple chain by proximity)
	if len(poleNums) > 1 {
		// Map entity number → position
		pos := map[int]bpPosition{}
		for _, e := range bp.Entities {
			pos[e.EntityNumber] = e.Position
		}
		for i := 0; i < len(poleNums); i++ {
			bestJ := -1
			bestD := 1e9
			for j := i + 1; j < len(poleNums); j++ {
				a, b := pos[poleNums[i]], pos[poleNums[j]]
				d := math.Hypot(a.X-b.X, a.Y-b.Y)
				if d < bestD && d <= data.BasePoles["medium-electric-pole"].WireReach {
					bestD = d
					bestJ = j
				}
			}
			if bestJ >= 0 {
				a, b := poleNums[i], poleNums[bestJ]
				if a > b {
					a, b = b, a
				}
				bp.Wires = append(bp.Wires, []int{a, data.WirePoleCopper, b, data.WirePoleCopper})
			}
		}
	}

	if len(bp.Entities) == 0 {
		return "", fmt.Errorf("no entities to emit")
	}

	root := bpRoot{Blueprint: bp}
	raw, err := json.Marshal(root)
	if err != nil {
		return "", fmt.Errorf("marshal blueprint: %w", err)
	}
	return CompressAndEncode(string(raw)), nil
}

func entityCentre(e route.Entity) (float64, float64) {
	w, h := 1, 1
	if e.Kind == route.KindMachine {
		s := data.Footprint(e.Name, nil)
		w, h = s.W, s.H
	}
	if e.Kind == route.KindPole {
		if p, ok := data.BasePoles[e.Name]; ok {
			w, h = p.Size, p.Size
		}
	}
	if e.Kind == route.KindBeacon {
		w, h = 3, 3
	}
	return float64(e.X) + float64(w)/2, float64(e.Y) + float64(h)/2
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// CompressAndEncode produces a Factorio blueprint string: "0" + base64(zlib(json)).
func CompressAndEncode(jsonString string) string {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, _ = w.Write([]byte(jsonString))
	_ = w.Close()
	return "0" + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// DecodeToJSON reverses CompressAndEncode.
func DecodeToJSON(blueprintString string) (string, error) {
	s := blueprintString
	if len(s) == 0 || s[0] != '0' {
		return "", fmt.Errorf("unknown blueprint version")
	}
	raw, err := base64.StdEncoding.DecodeString(s[1:])
	if err != nil {
		return "", err
	}
	r, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	defer r.Close()
	var out bytes.Buffer
	if _, err := out.ReadFrom(r); err != nil {
		return "", err
	}
	return out.String(), nil
}
