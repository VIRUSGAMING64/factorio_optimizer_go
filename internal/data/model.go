// Package data holds Factorio game data models ported from the Kotlin original.
package data

import "encoding/json"

// RootData is the top-level structure loaded from recipes.json.
type RootData struct {
	GameVersion  string              `json:"game_version"`
	Exporter     string              `json:"exporter,omitempty"`
	Mods         map[string]string   `json:"mods,omitempty"`
	Groups       map[string]Group    `json:"groups"`
	Quality      map[string]Quality  `json:"quality"`
	QualityNames []string            `json:"quality_names"`
	Recipes      map[string]Recipe   `json:"recipes"`
	Items        map[string]Item     `json:"items"`
	Fluids       map[string]Fluid    `json:"fluids"`
	Entities     map[string]Entity   `json:"entities"`
	Logistics    *LogisticsData      `json:"logistics,omitempty"`
}

// Group represents a recipe/item group.
type Group struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Order         string `json:"order"`
	OrderInRecipe string `json:"order_in_recipe"`
}

// Quality tier data (normal, uncommon, rare, epic, legendary).
type Quality struct {
	Name                                string   `json:"name"`
	Level                               int      `json:"level"`
	NextProbability                     float64  `json:"next_probability"`
	BeaconPowerUsageMultiplier          float64  `json:"beacon_power_usage_multiplier"`
	MiningDrillResourceDrainMultiplier  float64  `json:"mining_drill_resource_drain_multiplier"`
	Group                               string   `json:"group"`
	Subgroup                            string   `json:"subgroup"`
	Next                                *string  `json:"next,omitempty"`
	TranslatedName                      string   `json:"translated_name"`
	ChainProbability                    *float64 `json:"chain_probability,omitempty"`
}

// Chain returns the chance a craft that reached this tier goes one further.
func (q Quality) Chain() float64 {
	if q.ChainProbability != nil {
		return *q.ChainProbability
	}
	return q.NextProbability
}

// Recipe describes a crafting recipe.
type Recipe struct {
	Name                string              `json:"name"`
	Category            string              `json:"category"`
	Ingredients         FlexibleIngredients `json:"ingredients"`
	Products            FlexibleProducts    `json:"products"`
	MainProduct         *Product            `json:"main_product,omitempty"`
	AllowedEffects      AllowedEffects      `json:"allowed_effects"`
	MaximumProductivity int                 `json:"maximum_productivity"`
	Energy              float64             `json:"energy"`
	Order               string              `json:"order"`
	Group               string              `json:"group"`
	Subgroup            string              `json:"subgroup"`
	Enabled             bool                `json:"enabled"`
	ProductivityBonus   int                 `json:"productivity_bonus"`
	TranslatedName      string              `json:"translated_name"`
	Categories          []string            `json:"categories,omitempty"`
}

// AllCategories returns every category the recipe belongs to.
func (r Recipe) AllCategories() []string {
	if len(r.Categories) > 0 {
		return r.Categories
	}
	return []string{r.Category}
}

// InAny reports whether the recipe is in at least one of the given categories.
func (r Recipe) InAny(cats map[string]struct{}) bool {
	for _, c := range r.AllCategories() {
		if _, ok := cats[c]; ok {
			return true
		}
	}
	return false
}

// FlexibleIngredients accepts either a list or an empty object {} from the exporter.
type FlexibleIngredients []Ingredient

func (f *FlexibleIngredients) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == "{}" {
		*f = nil
		return nil
	}
	var list []Ingredient
	if err := json.Unmarshal(b, &list); err != nil {
		var obj map[string]any
		if err2 := json.Unmarshal(b, &obj); err2 == nil {
			*f = nil
			return nil
		}
		return err
	}
	*f = list
	return nil
}

// FlexibleProducts accepts either a list or an empty object {}.
type FlexibleProducts []Product

func (f *FlexibleProducts) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == "{}" {
		*f = nil
		return nil
	}
	var list []Product
	if err := json.Unmarshal(b, &list); err != nil {
		var obj map[string]any
		if err2 := json.Unmarshal(b, &obj); err2 == nil {
			*f = nil
			return nil
		}
		return err
	}
	*f = list
	return nil
}


// FlexibleFloat accepts a plain number or a quality→speed map (Factorio 2.0+).
// When it is a map we keep the "normal" value as the default speed.
type FlexibleFloat float64

func (f *FlexibleFloat) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		*f = 0
		return nil
	}
	var v float64
	if err := json.Unmarshal(b, &v); err == nil {
		*f = FlexibleFloat(v)
		return nil
	}
	var m map[string]float64
	if err := json.Unmarshal(b, &m); err == nil {
		if n, ok := m["normal"]; ok {
			*f = FlexibleFloat(n)
		} else {
			for _, val := range m {
				*f = FlexibleFloat(val)
				break
			}
		}
		return nil
	}
	return json.Unmarshal(b, &v) // return original error shape
}

func (f FlexibleFloat) Float() float64 { return float64(f) }


// FlexibleEffects accepts either the object form (recipes) or the string-list form (entities).
type FlexibleEffects struct {
	Consumption  bool
	Speed        bool
	Productivity bool
	Pollution    bool
	Quality      bool
}

func (f *FlexibleEffects) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		return nil
	}
	// list of strings
	var list []string
	if err := json.Unmarshal(b, &list); err == nil {
		for _, name := range list {
			switch name {
			case "consumption":
				f.Consumption = true
			case "speed":
				f.Speed = true
			case "productivity":
				f.Productivity = true
			case "pollution":
				f.Pollution = true
			case "quality":
				f.Quality = true
			}
		}
		return nil
	}
	// object form
	var obj AllowedEffects
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}
	*f = FlexibleEffects{
		Consumption:  obj.Consumption,
		Speed:        obj.Speed,
		Productivity: obj.Productivity,
		Pollution:    obj.Pollution,
		Quality:      obj.Quality,
	}
	return nil
}

// ProbabilityRange is a share of a common random roll (Factorio 2.1).
type ProbabilityRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Ingredient of a recipe.
type Ingredient struct {
	Type               string   `json:"type"`
	Name               string   `json:"name"`
	Amount             float64  `json:"amount"`
	FluidboxMultiplier *int     `json:"fluidbox_multiplier,omitempty"`
	IgnoredByStats     *float64 `json:"ignored_by_stats,omitempty"`
	FluidboxIndex      *int     `json:"fluidbox_index,omitempty"`
}

// IsFluid returns true if this ingredient is a fluid.
func (i Ingredient) IsFluid() bool { return i.Type == "fluid" }

// Product of a recipe.
type Product struct {
	Type                   string            `json:"type"`
	Name                   string            `json:"name"`
	Amount                 float64           `json:"amount"`
	Probability            *float64          `json:"probability,omitempty"`
	Temperature            *float64          `json:"temperature,omitempty"`
	ExtraCountFraction     *float64          `json:"extra_count_fraction,omitempty"`
	FluidboxIndex          *int              `json:"fluidbox_index,omitempty"`
	FluidboxMultiplier     *int              `json:"fluidbox_multiplier,omitempty"`
	PercentSpoiled         *float64          `json:"percent_spoiled,omitempty"`
	IgnoredByStats         *float64          `json:"ignored_by_stats,omitempty"`
	IgnoredByProductivity  *float64          `json:"ignored_by_productivity,omitempty"`
	IndependentProbability *float64          `json:"independent_probability,omitempty"`
	SharedProbability      *ProbabilityRange `json:"shared_probability,omitempty"`
	AmountMin              *float64          `json:"amount_min,omitempty"`
	AmountMax              *float64          `json:"amount_max,omitempty"`
}

// MeanAmount returns items per successful craft.
func (p Product) MeanAmount() float64 {
	if p.AmountMin != nil && p.AmountMax != nil && p.Amount == 0 {
		return (*p.AmountMin + *p.AmountMax) / 2
	}
	return p.Amount
}

// Chance returns the probability a craft yields this product.
func (p Product) Chance() float64 {
	base := 1.0
	if p.IndependentProbability != nil {
		base = *p.IndependentProbability
	} else if p.Probability != nil {
		base = *p.Probability
	}
	if p.SharedProbability != nil {
		base *= p.SharedProbability.Max - p.SharedProbability.Min
	}
	return base
}

// ExpectedAmount returns expected items per craft (including extra fraction).
func (p Product) ExpectedAmount() float64 {
	extra := 0.0
	if p.ExtraCountFraction != nil {
		extra = *p.ExtraCountFraction
	}
	return p.MeanAmount()*p.Chance() + extra
}

// AllowedEffects describes which module effects a recipe accepts.
type AllowedEffects struct {
	Consumption  bool `json:"consumption"`
	Speed        bool `json:"speed"`
	Productivity bool `json:"productivity"`
	Pollution    bool `json:"pollution"`
	Quality      bool `json:"quality"`
}

// Item in the game.
type Item struct {
	Name           string         `json:"name"`
	Type           string         `json:"type"`
	Order          string         `json:"order"`
	Group          string         `json:"group"`
	Subgroup       string         `json:"subgroup"`
	StackSize      int            `json:"stack_size"`
	Weight         float64        `json:"weight"`
	FuelValue      int64          `json:"fuel_value"`
	TranslatedName string         `json:"translated_name"`
	FuelCategory   *string        `json:"fuel_category,omitempty"`
	Flags          []string       `json:"flags,omitempty"`
	ModuleEffects  *ModuleEffects `json:"module_effects,omitempty"`
	FuelCategories []string       `json:"fuel_categories,omitempty"`
}

// AllFuelCategories returns every fuel category of the item.
func (i Item) AllFuelCategories() []string {
	if len(i.FuelCategories) > 0 {
		return i.FuelCategories
	}
	if i.FuelCategory != nil {
		return []string{*i.FuelCategory}
	}
	return nil
}

// ModuleEffects from modules.
type ModuleEffects struct {
	Consumption  *float64 `json:"consumption,omitempty"`
	Speed        *float64 `json:"speed,omitempty"`
	Productivity *float64 `json:"productivity,omitempty"`
	Pollution    *float64 `json:"pollution,omitempty"`
	Quality      *float64 `json:"quality,omitempty"`
}

// Fluid in the game.
type Fluid struct {
	Name           string `json:"name"`
	Order          string `json:"order"`
	Group          string `json:"group"`
	Subgroup       string `json:"subgroup"`
	FuelValue      int64  `json:"fuel_value"`
	TranslatedName string `json:"translated_name"`
}

// Entity (assembling machine, furnace, chemical plant, etc.).
type Entity struct {
	Name                   string          `json:"name"`
	Type                   string          `json:"type"`
	Order                  string          `json:"order"`
	Group                  string          `json:"group"`
	Subgroup               string          `json:"subgroup"`
	TranslatedName         string          `json:"translated_name"`
	CraftingCategories     []string        `json:"crafting_categories,omitempty"`
	CraftingSpeed          FlexibleFloat   `json:"crafting_speed"`
	ModuleSlots            int             `json:"module_slots"`
	AllowedEffectsList     FlexibleEffects `json:"allowed_effects,omitempty"`
	EnergyUsage            FlexibleFloat   `json:"energy_usage,omitempty"`
	EnergySource           string          `json:"energy_source,omitempty"`
	Width                  float64         `json:"width,omitempty"`
	Height                 float64         `json:"height,omitempty"`
	CollisionBox           *CollisionBox   `json:"collision_box,omitempty"`
	SelectionBox           *CollisionBox   `json:"selection_box,omitempty"`
	FluidBoxes             []FluidBoxData  `json:"fluid_boxes,omitempty"`
	ProductivityBonus      float64         `json:"productivity_bonus,omitempty"`
	SupplyAreaDistance     *FlexibleFloat  `json:"supply_area_distance,omitempty"`
	MaxWireDistance        *FlexibleFloat  `json:"max_wire_distance,omitempty"`
	DistributionEffectivity *FlexibleFloat `json:"distribution_effectivity,omitempty"`
	ModuleInventorySize    *int            `json:"module_inventory_size,omitempty"`
	Hidden                 *bool           `json:"hidden,omitempty"`
}

// CollisionBox from Factorio entity data.
type CollisionBox struct {
	LeftTop     Vec `json:"left_top"`
	RightBottom Vec `json:"right_bottom"`
}

// FluidBoxData of a crafting machine (facing north).
type FluidBoxData struct {
	Index          int                  `json:"index"`
	ProductionType string               `json:"production_type"`
	Filter         *string              `json:"filter,omitempty"`
	Connections    []PipeConnectionData `json:"connections,omitempty"`
}

// PipeConnectionData describes one pipe connection of a fluid box.
type PipeConnectionData struct {
	FlowDirection          string  `json:"flow_direction"`
	ConnectionType         string  `json:"connection_type"`
	Direction              *string `json:"direction,omitempty"`
	Positions              []Vec   `json:"positions"`
	MaxUndergroundDistance *int    `json:"max_underground_distance,omitempty"`
}

// Vec is a 2D vector / position.
type Vec struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// LogisticsData holds belts, inserters, poles, pipes, etc.
type LogisticsData struct {
	Belts         []BeltData         `json:"belts,omitempty"`
	Undergrounds  []UndergroundData  `json:"undergrounds,omitempty"`
	Splitters     []SplitterData     `json:"splitters,omitempty"`
	Inserters     []InserterData     `json:"inserters,omitempty"`
	Poles         []PoleData         `json:"poles,omitempty"`
	PipesToGround []PipeToGroundData `json:"pipes_to_ground,omitempty"`
}

type BeltData struct {
	Name           string  `json:"name"`
	Speed          float64 `json:"speed"`
	ItemsPerSecond float64 `json:"items_per_second"`
	Underground    *string `json:"underground,omitempty"`
	Hidden         *bool   `json:"hidden,omitempty"`
}

type UndergroundData struct {
	Name        string  `json:"name"`
	Speed       float64 `json:"speed"`
	MaxDistance int     `json:"max_distance"`
	Hidden      *bool   `json:"hidden,omitempty"`
}

type SplitterData struct {
	Name   string  `json:"name"`
	Speed  float64 `json:"speed"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Hidden *bool   `json:"hidden,omitempty"`
}

type InserterData struct {
	Name           string  `json:"name"`
	RotationSpeed  float64 `json:"rotation_speed"`
	ExtensionSpeed float64 `json:"extension_speed"`
	Pickup         Vec     `json:"pickup"`
	Drop           Vec     `json:"drop"`
	Bulk           *bool   `json:"bulk,omitempty"`
	FilterCount    *int    `json:"filter_count,omitempty"`
	StackSizeBonus *int    `json:"stack_size_bonus,omitempty"`
	EnergySource   *string `json:"energy_source,omitempty"`
	Hidden         *bool   `json:"hidden,omitempty"`
}

type PoleData struct {
	Name               string  `json:"name"`
	Width              int     `json:"width"`
	Height             int     `json:"height"`
	SupplyAreaDistance float64 `json:"supply_area_distance"`
	MaxWireDistance    float64 `json:"max_wire_distance"`
	Hidden             *bool   `json:"hidden,omitempty"`
}

type PipeToGroundData struct {
	Name        string `json:"name"`
	MaxDistance *int   `json:"max_distance,omitempty"`
	Hidden      *bool  `json:"hidden,omitempty"`
}
