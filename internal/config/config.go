// Package config holds job configuration and tuning parameters.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ConfigTarget is a desired production rate of an item.
type ConfigTarget struct {
	Item    string  `json:"item"`
	Name    string  `json:"name"`
	Rate    float64 `json:"rate"`
	Amount  float64 `json:"amount"`
	Quality string  `json:"quality,omitempty"`
}

func (t ConfigTarget) ResolvedItem() string {
	if t.Item != "" {
		return t.Item
	}
	return t.Name
}

func (t ConfigTarget) ResolvedRate() float64 {
	if t.Rate != 0 {
		return t.Rate
	}
	return t.Amount
}

// BoxConfig controls the search box size.
type BoxConfig struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

// BlocksConfig for factory-of-blocks mode.
type BlocksConfig struct {
	MaxMachines int     `json:"max_machines"`
	CompactS    float64 `json:"compact_s,omitempty"`
}

// InserterConfig controls which inserters may be used.
type InserterConfig struct {
	Allowed []string `json:"allowed,omitempty"`
}

// ResearchConfig for inserter capacity etc.
type ResearchConfig struct {
	InserterCapacity int `json:"inserter_capacity,omitempty"`
}

// RecyclingConfig controls recycler usage.
type RecyclingConfig struct {
	Mode      string   `json:"mode"` // off, void, on
	VoidItems []string `json:"void_items,omitempty"`
}

// HeatConfig for Aquilo heating.
type HeatConfig struct {
	Enabled bool `json:"enabled"`
}

// SpoilageConfig for Gleba.
type SpoilageConfig struct {
	Drain bool `json:"drain"`
}

// QualityProductionConfig for quality modules + recyclers.
type QualityProductionConfig struct {
	Enabled bool `json:"enabled"`
	Recycle bool `json:"recycle"`
}

// TileRect is a forbidden or reserved area.
type TileRect struct {
	X, Y, W, H int `json:"x" json:"y" json:"w" json:"h"`
}

// RunConfiguration is the main job description.
type RunConfiguration struct {
	Engine             string                   `json:"engine"`
	Data               string                   `json:"data,omitempty"`
	Quality            string                   `json:"quality"`
	Targets            []ConfigTarget           `json:"targets"`
	ForbiddenRecipes   []string                 `json:"forbidden_recipes"`
	RecipeCategories   []string                 `json:"recipe_categories,omitempty"`
	Machine            string                   `json:"machine"`
	Machines           []string                 `json:"machines,omitempty"`
	Fuel               string                   `json:"fuel"`
	BeltTier           string                   `json:"belt_tier"`
	BeltStackSize      int                      `json:"belt_stack_size"`
	LaneSharing        bool                     `json:"lane_sharing"`
	WeaveTier          string                   `json:"weave_tier"`
	Box                BoxConfig                `json:"box"`
	Blocks             *BlocksConfig            `json:"blocks,omitempty"`
	Tricks             map[string]bool          `json:"tricks,omitempty"`
	Tuning             map[string]any           `json:"tuning,omitempty"`
	Inserters          InserterConfig           `json:"inserters"`
	Research           ResearchConfig           `json:"research"`
	Recycling          RecyclingConfig          `json:"recycling"`
	Heat               HeatConfig               `json:"heat"`
	Spoilage           SpoilageConfig           `json:"spoilage"`
	QualityProduction  *QualityProductionConfig `json:"quality_production,omitempty"`
	ForbiddenTiles     []TileRect               `json:"forbidden_tiles,omitempty"`
}

// DefaultCategories used when recipe_categories is empty.
var DefaultCategories = []string{
	"crafting", "advanced-crafting", "electronics", "crafting-with-fluid",
	"electronics-with-fluid", "pressing",
}

// DefaultTricks from the original (v1 on, later off until implemented).
var DefaultTricks = map[string]bool{
	"I1": true, "I2": true, "I3": true, "I4": true, "I5": true,
	"I6": true, "I7": true, "I8": true, "I9": true,
	"I10": false, "I11": false, "I12": false, "I13": false,
	"B1": true, "B2": true, "B3": true, "B4": true, "B5": true, "B6": true,
	"B7": false, "B8": false, "B10": false, "B11": false, "B12": false, "B13": false,
	"B9": true, "B14": true, "F1": true, "F2": true, "P1": true, "P2": true, "H1": true,
}

// Trick returns whether a trick is enabled.
func (c *RunConfiguration) Trick(name string) bool {
	if c.Tricks != nil {
		if v, ok := c.Tricks[name]; ok {
			return v
		}
	}
	if v, ok := DefaultTricks[name]; ok {
		return v
	}
	return false
}

// Load reads a RunConfiguration from JSON.
func Load(path string) (*RunConfiguration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		data, err = os.ReadFile("data/" + path)
		if err != nil {
			return nil, fmt.Errorf("load config %q: %w", path, err)
		}
	}
	var cfg RunConfiguration
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	return &cfg, nil
}

func (c *RunConfiguration) ApplyDefaults() {
	c.applyDefaults()
}

func (c *RunConfiguration) applyDefaults() {
	if c.Engine == "" {
		c.Engine = "sa"
	}
	if c.Quality == "" {
		c.Quality = "normal"
	}
	if c.Machine == "" {
		c.Machine = "assembling-machine-3"
	}
	if c.Fuel == "" {
		c.Fuel = "coal"
	}
	if c.BeltTier == "" {
		c.BeltTier = "fast"
	}
	if c.BeltStackSize == 0 {
		c.BeltStackSize = 1
	}
	if c.WeaveTier == "" {
		c.WeaveTier = "auto"
	}
	if c.Recycling.Mode == "" {
		c.Recycling.Mode = "off"
	}
	if len(c.Inserters.Allowed) == 0 {
		c.Inserters.Allowed = []string{"inserter", "long-handed-inserter", "fast-inserter", "bulk-inserter"}
	}
}

// Tuning holds algorithm hyperparameters.
type Tuning struct {
	SAInitialTemp float64 `json:"sa_initial_temp"`
	SACoolingRate float64 `json:"sa_cooling_rate"`
	SAIterations  int     `json:"sa_iterations"`
	RouteBudgetMs int     `json:"route_budget_ms"`
	SearchThreads int     `json:"search_threads"`
	AreaFactor    float64 `json:"area_factor"`
	// Placement
	RudyBin       int     `json:"rudy_bin"`
	CongestionWt  float64 `json:"congestion_weight"`
	// Routing
	HistDecay     float64 `json:"hist_decay"`
	PresentFac    float64 `json:"present_factor"`
}

// Defaults returns development defaults.
func Defaults() Tuning {
	return Tuning{
		SAInitialTemp: 1000,
		SACoolingRate: 0.995,
		SAIterations:  50000,
		RouteBudgetMs: 30000,
		SearchThreads: 4,
		AreaFactor:    1.35,
		RudyBin:       4,
		CongestionWt:  1.0,
		HistDecay:     0.99,
		PresentFac:    1.0,
	}
}

// LoadTuning loads tuning from JSON (optional).
func LoadTuning(path string) (Tuning, error) {
	t := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		data, err = os.ReadFile("data/" + path)
		if err != nil {
			return t, nil
		}
	}
	if err := json.Unmarshal(data, &t); err != nil {
		return t, fmt.Errorf("parse tuning: %w", err)
	}
	return t, nil
}
