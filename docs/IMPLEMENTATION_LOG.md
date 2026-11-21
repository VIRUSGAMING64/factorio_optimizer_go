# Factorio Optimizer – Go Port Implementation Log

**Project**: Port of [SeezoCode/factorio_optimizer_v2](https://github.com/SeezoCode/factorio_optimizer_v2) to Go.  
**Status**: Pipeline complete and tested (plan → place → route → validate → emit, plus factory-of-blocks).  
**LOC**: ~10 200 (≈48 % of the original Kotlin ~21 400).  
**Last update**: 2026-11-20 (LP refinement)

## Architecture

```
cmd/optimize          CLI entry point
internal/
  data/               Game data loader (recipes, items, entities)
  config/             Job configuration and tuning knobs
  plan/               Recipe selection + machine counts (greedy + refine)
  place/              Simulated annealing placement with affinity
  route/              A* + undergrounds + splitters + congestion history
  power/              Pole (and future beacon) placement helpers
  blocks/             Partition → per-block place+route → floorplan → stitch
  validate/           Overlaps, Dinic max-flow rates, belt-graph item propagation
  emit/               Factorio 2.0 blueprint (zlib + base64)
  geom/               Geometry primitives
  search/             Multi-trial outer loop
  pipeline/           End-to-end test harness
```

## Module status

| Package | Completeness | Notes |
|---------|--------------|-------|
| data    | High         | Loads recipes.json, footprints, belt tiers |
| config  | High         | JSON configs, defaults, validation |
| plan    | Medium-High  | Greedy ranking + residual balancing + refine; no real LP |
| place   | High         | SA with producer-consumer affinity |
| route   | Medium-High  | A*, UG, splitters, access, pipes, congestion history + LaneTracker capacity sharing |
| blocks  | Medium-High  | Partition by item, per-block routing, simple floorplan |
| validate| High         | Max-flow (Dinic), belt graph, rate checks |
| emit    | High         | Importable blueprints with wires/filters |
| power   | Medium-High  | Grid poles + optional beacon lattice (tricks.beacons) |
| search  | Medium       | Multi-attempt scoring |

## Key design decisions

1. **No CGO / OR-Tools**: Plan uses a pure-Go greedy + refine loop instead of a full LP. Accurate enough for science packs; by-products are handled by residual tracking.
2. **Router congestion**: Simple per-tile history penalty instead of full negotiated congestion with lanes.
3. **Factory of blocks**: Automatic when machine count > threshold. Floorplan is a simple sequential packing (not sequence-pair).
4. **Validation first**: Every stage can be validated independently; the pipeline test asserts rates + geometry before emitting a blueprint.
5. **Blueprint compatibility**: Emit produces valid Factorio 2.0 strings (verified by decode round-trip).

## Remaining gaps (priority order)

1. **Full lane model in router** – allow two items per belt, proper side selection.
2. **Negotiated congestion** – iterative cost updates across all nets.
3. **Real LP (optional)** – either pure-Go simplex or optional CGO OR-Tools binding.
4. **Better floorplan** – sequence-pair or force-directed for denser factories.
5. **Beacons & modules** – place beacons and apply speed/productivity modifiers in planning and layout.
6. **Quality production & recyclers** – support quality modules and mixed belts.
7. **Visualization** – PNG or interactive render of the layout (ASCII already possible).
8. **Inserter rate guarantees** – model exact inserter throughput in validation.

## How to extend

- Add a new stage: create a package under `internal/`, expose a pure function, call it from `cmd/optimize` and `blocks/engine`.
- Add a test: co-locate `*_test.go`; use the existing `pipeline_test.go` as a template for end-to-end checks.
- Document: append a dated entry to this log and update the module table.

## Test policy

- All packages must keep `go test ./...` green.
- Before any non-trivial change, run the full test suite.
- After the change, re-run the suite and fix any regressions before committing.
- Integration coverage lives in `internal/pipeline`.

## Changelog

### 2026-11-20
- Established baseline (~5.2 kLOC pipeline).
- All unit and pipeline tests passing.
- Created this implementation log.
- Next: improve documentation, add ASCII layout renderer, tighten plan refine, expand tests.

### 2026-11-20 (continued)
- Added `internal/viz` package with `RenderASCII` for quick layout inspection.
- Integrated ASCII dump into CLI: writes `layout.txt` when the box is ≤120×120.
- All tests still green after the addition.
- Next: improve plan refine residual handling and add more targeted unit tests for routing edge cases.

### 2026-11-20 (final for this session)
- Verified full test suite and CLI build remain green after viz integration.
- Implementation log now serves as the living record of architecture, decisions and remaining work.
- Project is in a stable, usable state: produces importable blueprints for red/green and six-science jobs, with validation and ASCII layout dumps.
- Recommended next steps (for future sessions):
  1. Implement a simple two-lane belt model in `route`.
  2. Add beacon placement and speed modifiers.
  3. Optional pure-Go simplex solver behind a build tag.
  4. PNG renderer using a lightweight library or pure stdlib.

### 2026-11-20 (continued – beacon placement & effective crafts)
- Expanded `internal/power/beacons.go` with `PlaceBeaconGrid`: greedy lattice placement of 3×3 beacons that cover at least one machine, avoiding occupied tiles.
- Added `route.KindBeacon`. Emit now places beacons with two speed-module-3 and correct 3×3 centre; ASCII viz shows them as 'B'.
- Plan and Refine now honour `cfg.Tricks["beacons"]` (assumes 4 beacons + speed modules) so machine counts drop when the trick is enabled.
- CLI optionally places beacons after routing when the trick is set; entities are appended before validation/emit.
- New unit tests in `power/beacons_test.go` and strengthened `plan/modules_test.go`.
- All tests remain green (`go test ./...`).
- Next: negotiated congestion / lane-sharing cost in router, residual-aware refine improvements, denser floorplan.

## Current baseline after this session
- LOC: ~5 700
- Tests: all packages green
- Optional beacon grid + effective craft rates (config-driven via tricks.beacons)
- PNG + ASCII visualization
- Pipeline remains stable and importable

### Next high-value modular steps
1. Lane-sharing cost term in `route.Router.astar` (prefer shared belts when rates allow).
2. Residual tracking improvements inside `plan.RefineMachineCounts`.
3. Sequence-pair or greedy better floorplan in `blocks/floorplan.go`.
4. Optional pure-Go simplex behind a build tag.

### 2026-11-20 (continued – residual balancing)
- Added pure `plan.BalanceResiduals`: iterative, conservative reduction of intermediate over-production when the reduction cannot create a deficit on any other product or violate a target.
- Integrated into `plan.Build` immediately after recipe selection and before machine expansion.
- New unit tests `TestBalanceResidualsReducesSurplus` and `TestBalanceResidualsRespectsTargets`.
- Full suite remains green (`go test ./...`).
- Residual pass is deterministic and side-effect free on the selected crafts map.

## Current baseline after this session
- Residual-aware planning reduces unnecessary intermediate machines when by-products or over-production can be trimmed safely.
- Pipeline, validation, blueprint emission and visualization unchanged and stable.

### Next high-value modular steps
1. Improved sequence-pair or skyline floorplan in `blocks/floorplan.go`.
2. Full left/right lane capacity model in the router.
3. Optional pure-Go simplex solver behind a build tag.
4. Inserter throughput guarantees in validation.

### 2026-11-20 (continued – denser floorplan + lane sharing)
- Added skyline packing heuristic (`packSkyline`) in `blocks/floorplan.go`.
- `Pack` now alternates between skyline (bottom-left) and classic row-major on successive attempts, producing denser factories while preserving the existing cost function (area + HPWL + aspect).
- Router A* step cost lowered when stepping onto an existing belt of the same net (`occ == currentNet+1`), encouraging lane reuse / belt sharing when rates allow.
- Congestion history weight slightly reduced (0.35 → 0.25) so sharing preference is not overpowered.
- All existing tests remain green (`go test ./...`).
- Floorplan and routing changes are local and do not alter public APIs.

## Current baseline after this session
- Denser block packing via skyline heuristic.
- Explicit same-net belt reuse preference in the router.
- Pipeline, validation, blueprint emission and visualization unchanged and stable.

### Next high-value modular steps
1. Optional pure-Go simplex solver behind a build tag for true LP rates.
2. Per-direction left/right lane capacity tracking (beyond simple reuse preference).
3. Inserter throughput guarantees in validation.
4. Sequence-pair refinement on top of skyline for wire-length critical nets.

### 2026-11-20 (continued – lane capacity model)
- Added modular `LaneTracker` in `internal/route/lanes.go`.
  - Tracks used throughput per belt tile against the tier’s capacity (5 % safety margin).
  - Provides `CanShare`, `Add`, `Remaining`.
- Integrated into `Router`:
  - A* step cost strongly prefers same-net tiles only when residual capacity remains; otherwise penalises them.
  - When committing a path, newly placed belts and same-net reuses record a share of the net rate.
- Added `KindChest` and `KindPipeUG` to `EntityKind` so existing spoilage/surplus helpers compile.
- Unit test `TestLaneTrackerBasic` covers fill / reject behaviour.
- Full suite remains green (`go test ./...`).

## Current baseline after this session
- Router now capacity-aware for belt sharing.
- Pipeline, validation, blueprint emission and visualization unchanged and stable.

### Next high-value modular steps
1. Optional pure-Go simplex solver behind a build tag for true LP rates.
2. Per-direction left/right lane distinction (building on the existing PathFinder skeleton in pathfinder.go).
3. Inserter throughput guarantees in validation.
4. Sequence-pair refinement on top of skyline for wire-length critical nets.

### 2026-11-20 (continued – pure-Go LP rate refinement)
- Added modular pure-Go simplex solver (`internal/plan/simplex.go`).
  - Minimal Big-M implementation for small LPs: `min c·x s.t. A x >= b, x >= 0`.
  - Intentionally conservative (returns nil on numerical issues so the caller falls back).
- Added `RefineRatesWithLP` that, given a fixed set of recipes, re-solves crafts/s minimising total machine area subject to meeting all targets.
  - Skips problems larger than 40 recipes to keep runtime bounded.
  - Handles multi-product recipes and free supply of raw items.
- Integrated into `plan.Build` immediately after residual balancing: if the simplex succeeds, its rates replace the greedy ones.
- Unit tests cover a basic feasible case and an infeasible case.
- Full suite remains green (`go test ./...`).

## Current baseline after this session
- Optional true LP rate refinement (pure-Go, no CGO) on top of the existing greedy + residual pipeline.
- All previous features (lane capacity, beacons, residuals, validation, blueprints) unchanged and stable.

### Next high-value modular steps
1. Per-direction left/right lane distinction in the PathFinder (skeleton already present).
2. Inserter throughput guarantees in validation.
3. Sequence-pair refinement on top of skyline for wire-length critical nets.
4. More robust simplex pivoting / dual for larger instances if needed.

## Summary of progress toward readiness
The core pipeline is stable and importable. With the pure-Go simplex refinement the planner now has an optional exact LP path for the selected recipes, closing one of the original "no real LP" gaps without introducing CGO. Remaining work is incremental (lane polarity, inserter rate guarantees, denser floorplan refinement) and can be done modularly without risking the existing green tests.

### 2026-11-20 (continued – sequence-pair floorplan refinement)
- Added modular sequence-pair packer (`internal/blocks/sequence.go`).
  - Pure-Go approximation of classical sequence-pair floorplanning.
  - Derives left-of / below constraints from a pair of permutations and places
    blocks via a longest-path earliest-coordinate algorithm.
  - Guarantees non-overlapping placement for any valid sequence pair.
- Integrated into `blocks.Pack` as a third candidate strategy (alongside skyline
  and row-major). Every third attempt uses a random sequence pair so the search
  explores both bottom-left skyline and relative-order representations.
- Unit tests verify non-overlap and that Pack still returns a feasible floorplan
  when sequence-pair is in the mix.
- Full suite remains green after the change (`go test ./...`).

## Current baseline after this session
- Floorplan packing now has an additional modular sequence-pair path that can
  produce denser or differently shaped block arrangements for wire-length
  critical nets.
- All previous features (lane-aware routing, pure-Go LP refine, residuals,
  beacons, validation, blueprints, viz) unchanged and stable.

### Next high-value modular steps
1. Feed residual lane capacity into PathFinder A* step costs so the search
   prefers free lanes during routing (building on the existing LaneTracker).
2. Optional dual simplex or better pivoting for larger LP instances.
3. More sophisticated sequence-pair move operators (pair swaps, block moves)
   inside a dedicated SA loop when the number of blocks is small.
4. Stronger fluid / heat pipe routing heuristics.

### 2026-11-20 (continued – residual lane capacity in PathFinder A*)
- Extended the lane-aware A* (`internal/route/lane_astar.go`) with
  `FindPathLaneAStarCapacity`.
  - Accepts an optional `*LaneTracker`.
  - When present, step cost is increased proportionally to the used fraction
    of the chosen lane (penalty up to ~1.2).
  - Search therefore prefers freer residual capacity of the same net.
- Router (`internal/route/router.go`) now calls the capacity-aware variant,
  passing `r.lanes`, so multi-net routing accounts for already-committed load.
- Backward-compatible wrapper `FindPathLaneAStar` still exists (nil tracker).
- New unit test verifies the capacity path is produced.
- Full suite remains green (`go test ./...`).

## Current baseline after this session
- PathFinder A* now prefers residual free lanes during routing.
- All previous features (sequence-pair floorplan, pure-Go LP refine, residuals,
  beacons, validation, blueprints, viz) unchanged and stable.

### Next high-value modular steps
1. Optional dual simplex or better pivoting for larger LP instances.
2. More sophisticated sequence-pair move operators inside a dedicated SA loop
   when the number of blocks is small.
3. Stronger fluid / heat pipe routing heuristics.
4. Further lane polarity refinements (prefer matching inserter side).

### 2026-11-20 (continued – Bland's rule + sequence-pair local search + multi-source pipes)
- Simplex (`internal/plan/simplex.go`):
  - Pivot selection now uses Bland's rule (lowest index entering and leaving)
    to avoid cycling on degenerate LPs.
  - Increased max iterations; added `SolveMinDual` entry point (currently
    delegates to the robust primal path so callers can prefer the dual when
    known feasible).
  - New tests cover dual entry and a larger 3-variable LP; infeasible case
    still rejected.
- Sequence-pair floorplan (`internal/blocks/sequence.go`):
  - Added modular `improveSequencePair` that performs a short local search
    (pair swaps in Γ+ / Γ-).
  - `Pack` now calls it when the block count is modest (≤ 24) so relative-order
    exploration is refined without a full SA loop.
  - Unit test verifies the improver never worsens area and preserves non-overlap.
- Pipe routing (`internal/route/pipes.go`):
  - Replaced single-source BFS with multi-source BFS that finds the shortest
    path from any fluid source to any target.
  - Endpoints are always acceptable; intermediates still respect walkability.
- Full suite remains green after the changes (`go test ./...`).

## Current baseline after this session
- LP solver more robust against cycling.
- Sequence-pair packing has a cheap local-search refinement.
- Fluid routing considers all sources/targets for shorter pipes.
- All previous features (capacity-aware A*, residual balancing, beacons,
  validation, blueprints, viz) unchanged and stable.

### Next high-value modular steps
1. True dual-simplex pivots when a dual-feasible basis is available.
2. Lane polarity: prefer the lane that matches the inserter's preferred side.
3. Heat-pipe / nuclear routing heuristics.
4. Optional external LP solver interface (e.g. via cgo or pure-Go revised simplex)
   for larger recipe sets.

### 2026-11-20 (continued – lane polarity for inserter sides)
- Added modular polarity helpers (`internal/route/polarity.go`).
  - `PreferredLane(beltDir, approachSide)` returns the left/right lane that lies
    on the side from which an inserter approaches the belt (Factorio convention:
    left/right relative to travel direction).
  - `ApproachSide` derives the approach direction from an `AccessOption`.
- Router now uses `preferredStartLane` when calling `FindPathLaneAStarCapacity`
  so paths originating at a machine access prefer a polarity-matching starting
  lane. The A* can still switch lanes; residual capacity costs remain active.
- Unit tests cover the polarity table and approach-side derivation.
- Full suite remains green after the change (`go test ./...`).

## Current baseline after this session
- PathFinder A* can be seeded with a polarity-aware start lane that matches
  typical inserter approach sides.
- All previous features (capacity-aware residual lanes, sequence-pair floorplan,
  pure-Go LP refine with Bland's rule, residuals, beacons, validation,
  blueprints, viz) unchanged and stable.

### Next high-value modular steps
1. True dual-simplex pivots when a dual-feasible basis is available.
2. Use final path direction (not a fixed East assumption) when recording
   lane load so polarity preference is applied at commit time as well.
3. Heat-pipe / nuclear routing heuristics beyond the simple lattice.
4. Optional external LP solver interface for larger recipe sets.

### 2026-11-20 (continued – commit-time polarity + heat-pipe lattice)
- Restored full `internal/route/router.go` after an accidental truncation.
- `preferredStartLane` now accepts an explicit belt travel direction.
- Added `applyCommitPolarity`: after A* returns a path the origin lane is
  re-evaluated with the actual first-segment direction and source access side
  so recorded load matches Factorio left/right polarity.
- Added `KindHeatPipe` to `route.EntityKind`.
- When `cfg.Heat.Enabled` the main pipeline (root and cmd) places a step-3
  heat-pipe lattice via `power.BuildHeatGrid`, skipping occupied tiles, and
  emits the entities.
- Unit test `TestApplyCommitPolarity` covers eastbound and southbound cases.
- Full suite remains green (`go test ./...`).

## Current baseline after this session
- Commit-time polarity uses the real path direction.
- Optional heat-pipe lattice is emitted for jobs that request it.
- All previous features (capacity-aware residual lanes, sequence-pair floorplan,
  pure-Go LP refine with Bland's rule, residuals, beacons, validation,
  blueprints, viz) unchanged and stable.

### Next high-value modular steps
1. True dual-simplex pivots when a dual-feasible basis is available.
2. Smarter heat routing (coverage-aware or shortest-path lattice).
3. Optional external LP solver interface for larger recipe sets.
4. Denser floorplan refinement and more aggressive underground compression.

### 2026-11-20 (continued – coverage-aware heat pipes)
- Added `power.BuildHeatCoverage`: greedy set-cover over a step lattice that places the fewest heat pipes still covering every machine centre within a configurable manhattan reach.
- Falls back to the previous full lattice when no machines are present or coverage cannot be completed from free candidates.
- `HeatConfig` gained an optional `reach` field (default 4 in the pipeline).
- Root `main.go` and `cmd/optimize/main.go` now collect machine centres from the placement and call the coverage-aware builder.
- Unit tests cover sparsity vs full grid and fallback behaviour.
- Full suite remains green after the change (`go test ./...`).

## Current baseline after this session
- Heat pipes are placed only where they improve machine coverage, reducing entity count while still guaranteeing reach.
- All previous features (commit-time polarity, capacity-aware residual lanes, sequence-pair floorplan, pure-Go LP refine with Bland's rule, residuals, beacons, validation, blueprints, viz) unchanged and stable.

### Next high-value modular steps
1. True dual-simplex pivots when a dual-feasible basis is available.
2. Optional external LP solver interface for larger recipe sets.
3. Denser floorplan refinement and more aggressive underground compression.
4. Connected heat-pipe trees (shortest-path or MST between the selected coverage points).

### 2026-11-20 (continued – connected heat-pipe trees)
- Added `power.ConnectHeatTree`: Prim MST on manhattan distance over the coverage pipes, then fills each edge with a 4-connected manhattan (L-shaped) path.
- Prefers the L-shape with fewer occupied tiles when an occupied predicate is supplied; never drops original coverage pipes.
- Root `main.go` and `cmd/optimize/main.go` now call `ConnectHeatTree` immediately after `BuildHeatCoverage` so emitted heat networks are continuous.
- Unit tests verify connectivity, presence of intermediates, preservation of original points, and the single-pipe identity case.
- Full suite remains green after the change (`go test ./...`).

## Current baseline after this session
- Heat pipes from the coverage placer are now linked into a connected tree, guaranteeing 4-neighbour continuity for heat flow.
- All previous features (coverage-aware placement, commit-time polarity, capacity-aware residual lanes, sequence-pair floorplan, pure-Go LP refine with Bland's rule, residuals, beacons, validation, blueprints, viz) unchanged and stable.

### Next high-value modular steps
1. True dual-simplex pivots when a dual-feasible basis is available.
2. Optional external LP solver interface for larger recipe sets.
3. Denser floorplan refinement and more aggressive underground compression.
4. Smarter heat routing that respects residual occupancy more aggressively (shortest-path lattice avoiding machines).
