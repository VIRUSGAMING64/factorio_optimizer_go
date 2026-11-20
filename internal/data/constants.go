package data

// Blueprint / wire constants (Factorio 2.0).
const (
	BlueprintVersion  int64   = (2 << 48) | (0 << 32) | (10 << 16) | 1
	WirePoleCopper    = 5
	WireRed           = 1
	WireGreen         = 2
	CircuitWireReach  = 9.0
	Pipe              = "pipe"
	PipeToGround      = "pipe-to-ground"
	PipeToGroundGap   = 9
	Beacon            = "beacon"
	BeaconSupply      = 3
	InfinityChest     = "infinity-chest"
	InfinityPipe      = "infinity-pipe"
	HeatPipe          = "heat-pipe"
	StorageTank       = "storage-tank"
	StorageTankCap    = 25000.0
	PipeCapacity      = 100.0
	MaxBeltStack      = 4
	StackInserter     = "stack-inserter"
	ParallelDropLane  = 1
)

// MarkerTiles usable as floor markers.
var MarkerTiles = []string{
	"stone-path", "concrete", "hazard-concrete-left", "hazard-concrete-right",
	"refined-concrete", "refined-hazard-concrete-left", "refined-hazard-concrete-right",
}

// BeltTier describes a transport-belt family.
type BeltTier struct {
	Key               string
	Belt              string
	Underground       string
	Splitter          string
	MaxUndergroundGap int
	Throughput        float64
}

// BaseBeltTiers are the vanilla tiers.
var BaseBeltTiers = map[string]BeltTier{
	"basic":   {Key: "basic", Belt: "transport-belt", Underground: "underground-belt", Splitter: "splitter", MaxUndergroundGap: 4, Throughput: 15},
	"fast":    {Key: "fast", Belt: "fast-transport-belt", Underground: "fast-underground-belt", Splitter: "fast-splitter", MaxUndergroundGap: 6, Throughput: 30},
	"express": {Key: "express", Belt: "express-transport-belt", Underground: "express-underground-belt", Splitter: "express-splitter", MaxUndergroundGap: 8, Throughput: 45},
	"turbo":   {Key: "turbo", Belt: "turbo-transport-belt", Underground: "turbo-underground-belt", Splitter: "turbo-splitter", MaxUndergroundGap: 10, Throughput: 60},
}

// BeltTierLookup returns a tier by key, defaulting to fast.
func BeltTierLookup(key string) BeltTier {
	if t, ok := BaseBeltTiers[key]; ok {
		return t
	}
	return BaseBeltTiers["fast"]
}

// TierOfEntity returns the belt tier for a belt/underground/splitter entity name.
func TierOfEntity(name string) *BeltTier {
	for k, t := range BaseBeltTiers {
		if t.Belt == name || t.Underground == name || t.Splitter == name {
			tt := BaseBeltTiers[k]
			return &tt
		}
	}
	return nil
}

// InserterSpec describes an inserter prototype.
type InserterSpec struct {
	Name      string
	Reach     int
	Filterable bool
	BaseRate  float64 // swings/s chest-to-chest at hand size 1
	Bulk      bool
	TierCost  int
}

// BaseInserters from vanilla.
var BaseInserters = map[string]InserterSpec{
	"burner-inserter":      {Name: "burner-inserter", Reach: 1, Filterable: false, BaseRate: 0.6, TierCost: 1},
	"inserter":             {Name: "inserter", Reach: 1, Filterable: false, BaseRate: 0.83, TierCost: 2},
	"long-handed-inserter": {Name: "long-handed-inserter", Reach: 2, Filterable: false, BaseRate: 1.15, TierCost: 3},
	"fast-inserter":        {Name: "fast-inserter", Reach: 1, Filterable: false, BaseRate: 2.31, TierCost: 4},
	"bulk-inserter":        {Name: "bulk-inserter", Reach: 1, Filterable: false, BaseRate: 2.31, Bulk: true, TierCost: 5},
	"stack-inserter":       {Name: "stack-inserter", Reach: 1, Filterable: false, BaseRate: 2.31, Bulk: true, TierCost: 6},
}

// HandSizes per inserter-capacity research level 0..7 (non-bulk, bulk).
var HandSizes = [][2]int{{1, 2}, {1, 3}, {2, 4}, {2, 5}, {2, 6}, {2, 8}, {2, 10}, {3, 12}}

// HandSizeAt returns hand sizes for a research level.
func HandSizeAt(level int) (nonBulk, bulk int) {
	if level < 0 {
		level = 0
	}
	if level >= len(HandSizes) {
		level = len(HandSizes) - 1
	}
	return HandSizes[level][0], HandSizes[level][1]
}

// InserterRate estimates items/s for an inserter configuration.
func InserterRate(name string, fromBelt bool, capacityLevel int, stackSize int) float64 {
	spec, ok := BaseInserters[name]
	if !ok {
		return 1.0
	}
	nb, b := HandSizeAt(capacityLevel)
	hand := nb
	if spec.Bulk {
		hand = b
	}
	if stackSize > 1 && (name == StackInserter || name == "bulk-inserter") {
		hand *= stackSize
	}
	rate := spec.BaseRate * float64(hand)
	if fromBelt {
		// belt pickup is slower
		rate *= 0.85
	}
	return rate
}

// PoleSpec describes an electric pole.
type PoleSpec struct {
	Name       string
	Size       int
	SupplyHalf float64
	WireReach  float64
}

// BasePoles from vanilla.
var BasePoles = map[string]PoleSpec{
	"small-electric-pole":  {Name: "small-electric-pole", Size: 1, SupplyHalf: 2.5, WireReach: 7.5},
	"medium-electric-pole": {Name: "medium-electric-pole", Size: 1, SupplyHalf: 3.5, WireReach: 9},
	"big-electric-pole":    {Name: "big-electric-pole", Size: 2, SupplyHalf: 2, WireReach: 30},
	"substation":           {Name: "substation", Size: 2, SupplyHalf: 9, WireReach: 18},
}

// Chests that can hold items.
var Chests = map[string]bool{
	"wooden-chest": true, "iron-chest": true, "steel-chest": true,
	"active-provider-chest": true, "passive-provider-chest": true,
	"storage-chest": true, "buffer-chest": true, "requester-chest": true,
	InfinityChest: true,
}

// Spoils maps item → (result, spoil_ticks). nil result = hatches/disappears.
var Spoils = map[string]struct {
	Result string
	Ticks  int
}{
	"raw-fish":              {"", 20 * 60 * 60},
	"yumako":                {"spoilage", 30 * 60 * 60},
	"jellynut":              {"spoilage", 30 * 60 * 60},
	"yumako-mash":           {"spoilage", 30 * 60 * 60},
	"jelly":                 {"spoilage", 30 * 60 * 60},
	"bioflux":               {"spoilage", 2 * 60 * 60},
	"agricultural-science-pack": {"spoilage", 30 * 60 * 60},
	"nutrients":             {"spoilage", 30 * 60 * 60},
	"biter-egg":             {"", 30 * 60 * 60},
}

// Qualities in order.
var Qualities = []string{"normal", "uncommon", "rare", "epic", "legendary"}

// ModuleBonusMultiplier for quality level 0..5.
func ModuleBonusMultiplier(level int) float64 {
	return 1.0 + 0.3*float64(level)
}

// QualityLevel returns 0 for normal … 5 for legendary.
func QualityLevel(name string) int {
	for i, q := range Qualities {
		if q == name {
			return i
		}
	}
	return 0
}

// MachineSize is tile footprint.
type MachineSize struct{ W, H int }

// BaseMachineSizes from Factorio 2.0 / Space Age.
var BaseMachineSizes = map[string]MachineSize{
	"assembling-machine-1":  {3, 3},
	"assembling-machine-2":  {3, 3},
	"assembling-machine-3":  {3, 3},
	"stone-furnace":         {2, 2},
	"steel-furnace":         {2, 2},
	"electric-furnace":      {3, 3},
	"chemical-plant":        {3, 3},
	"oil-refinery":          {5, 5},
	"centrifuge":            {3, 3},
	"rocket-silo":           {9, 9},
	"foundry":               {5, 5},
	"electromagnetic-plant": {4, 4},
	"cryogenic-plant":       {5, 5},
	"biochamber":            {3, 3},
	"crusher":               {2, 3},
	"recycler":              {2, 4},
	"beacon":                {3, 3},
	"lab":                   {3, 3},
	"boiler":                {2, 3},
	"steam-engine":          {3, 5},
	"steam-turbine":         {3, 5},
	"solar-panel":           {3, 3},
	"accumulator":           {2, 2},
	"radar":                 {3, 3},
	"roboport":              {4, 4},
}

// Footprint returns tile size of an entity.
func Footprint(name string, entities map[string]Entity) MachineSize {
	if e, ok := entities[name]; ok {
		w, h := int(e.Width), int(e.Height)
		if w > 0 && h > 0 {
			return MachineSize{w, h}
		}
	}
	if s, ok := BaseMachineSizes[name]; ok {
		return s
	}
	if p, ok := BasePoles[name]; ok {
		return MachineSize{p.Size, p.Size}
	}
	if t := TierOfEntity(name); t != nil && t.Splitter == name {
		return MachineSize{2, 1}
	}
	return MachineSize{1, 1}
}

// Flow direction for fluid connections.
type Flow int

const (
	FlowInput Flow = iota
	FlowOutput
)

// FluidConnection of a machine facing north.
type FluidConnection struct {
	Flow Flow
	Index int // 1-based among same flow
	DX, DY int
	Side   int // 0=N 1=E 2=S 3=W
}

// MachineGeom describes a crafting machine's geometry.
type MachineGeom struct {
	Name                         string
	W, H                         int
	Connections                  []FluidConnection
	ConnectionsOnlyWithFluidRecipe bool
	Burner                       bool
}

func assemblerFluids() []FluidConnection {
	return []FluidConnection{
		{FlowInput, 1, 1, -1, 0},
		{FlowOutput, 1, 1, 3, 2},
	}
}

// BaseFluidTable for machines facing north.
var BaseFluidTable = map[string]struct {
	Conns     []FluidConnection
	OnlyFluid bool
}{
	"assembling-machine-2": {assemblerFluids(), true},
	"assembling-machine-3": {assemblerFluids(), true},
	"chemical-plant": {[]FluidConnection{
		{FlowInput, 1, 0, -1, 0}, {FlowInput, 2, 2, -1, 0},
		{FlowOutput, 1, 0, 3, 2}, {FlowOutput, 2, 2, 3, 2},
	}, false},
	"oil-refinery": {[]FluidConnection{
		{FlowInput, 1, 1, 5, 2}, {FlowInput, 2, 3, 5, 2},
		{FlowOutput, 1, 0, -1, 0}, {FlowOutput, 2, 2, -1, 0}, {FlowOutput, 3, 4, -1, 0},
	}, false},
	"foundry": {[]FluidConnection{
		{FlowInput, 1, 1, -1, 0}, {FlowInput, 2, 3, -1, 0},
		{FlowOutput, 1, 1, 5, 2}, {FlowOutput, 2, 3, 5, 2},
	}, false},
	"biochamber": {[]FluidConnection{
		{FlowInput, 1, 1, -1, 0},
		{FlowOutput, 1, 1, 3, 2},
	}, false},
	"cryogenic-plant": {[]FluidConnection{
		{FlowInput, 1, 1, -1, 0}, {FlowInput, 2, 3, -1, 0},
		{FlowOutput, 1, 1, 5, 2}, {FlowOutput, 2, 3, 5, 2},
	}, false},
	"electromagnetic-plant": {[]FluidConnection{
		{FlowInput, 1, 1, -1, 0},
		{FlowOutput, 1, 1, 4, 2},
	}, false},
}

// FluidsOf returns fluid connections for a machine facing north.
func FluidsOf(name string) ([]FluidConnection, bool) {
	if t, ok := BaseFluidTable[name]; ok {
		return t.Conns, t.OnlyFluid
	}
	return nil, false
}

// BaseMachines preferred order.
var BaseMachines = []string{
	"assembling-machine-3", "assembling-machine-2", "assembling-machine-1",
	"electric-furnace", "steel-furnace", "stone-furnace",
	"chemical-plant", "oil-refinery", "centrifuge", "rocket-silo",
}

// SpaceAgeMachines require Space Age.
var SpaceAgeMachines = []string{
	"foundry", "electromagnetic-plant", "cryogenic-plant", "biochamber", "crusher", "recycler",
}

// SupportedMachines is base + space age.
var SupportedMachines = append(append([]string{}, BaseMachines...), SpaceAgeMachines...)

// PreferBaseCategories for recipe scoring.
var PreferBaseCategories = map[string]struct{}{
	"crafting": {}, "advanced-crafting": {}, "electronics": {},
	"crafting-with-fluid": {}, "electronics-with-fluid": {}, "pressing": {},
	"smelting": {}, "chemistry": {}, "oil-processing": {}, "centrifuging": {},
}

// ExcludedCategories the planner skips by default.
var ExcludedCategories = map[string]struct{}{
	"parameters": {}, "recycling": {}, "recycling-or-hand-crafting": {}, "spoiling": {},
}

// ExcludedSubgroups noise recipes.
var ExcludedSubgroups = map[string]struct{}{
	"fill-barrel": {}, "empty-barrel": {},
}

// BaseRaw resources always available via ports.
var BaseRaw = map[string]struct{}{
	"iron-ore": {}, "copper-ore": {}, "stone": {}, "coal": {}, "wood": {},
	"crude-oil": {}, "water": {}, "uranium-ore": {}, "raw-fish": {},
	"calcite": {}, "tungsten-ore": {}, "scrap": {}, "yumako": {}, "jellynut": {},
	"iron-bacteria": {}, "copper-bacteria": {}, "spoilage": {},
	"lava": {}, "fluorine": {}, "lithium-brine": {}, "ammoniacal-solution": {},
	"metallic-asteroid-chunk": {}, "carbonic-asteroid-chunk": {}, "oxide-asteroid-chunk": {},
}

// IsRaw reports whether an item is typically external.
func IsRaw(name string) bool {
	_, ok := BaseRaw[name]
	return ok
}

// Loaders by belt tier.
var Loaders = map[string]string{
	"basic": "loader", "fast": "fast-loader", "express": "express-loader", "turbo": "turbo-loader",
}

// TakesRecipe reports whether a machine blueprint needs a recipe field.
func TakesRecipe(name string) bool {
	if name == "recycler" {
		return false
	}
	if len(name) >= 7 && name[len(name)-7:] == "furnace" {
		return false
	}
	return true
}
