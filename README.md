# Factorio Optimizer (Go)

Go port of [SeezoCode/factorio_optimizer_v2](https://github.com/SeezoCode/factorio_optimizer_v2) — Factorio 2.0 layout optimizer:

**plan → place (SA + HPWL + bins + DI) → route (lane A* + multipass) → validate → emit blueprint**

## Requirements

- Go 1.22+
- `data/recipes.json` from the original project:

```bash
curl -L -o data/recipes.json \
  https://raw.githubusercontent.com/SeezoCode/factorio_optimizer_v2/main/data/recipes.json
```

## Run

```bash
go test ./...
go run ./cmd/optimize --config=data/red_green.json --sa-iters=30000 --attempts=2
go run ./cmd/optimize --config=data/config.json --sa-iters=8000
```

## Status (~9k LOC, 116+ tests)

Implemented: greedy plan + rebalance, SA placement (HPWL, congestion bins, DI credit, clusters),
lane-aware PathFinder, multipass/negotiate routing, factory-of-blocks + column templates,
validation (rates, layout detail, inserters, shape), blueprint emit, poles/beacons/heat, viz.

Still lighter than upstream: full Templates.kt, PlaceModel fluid stubs, real LP solver, UI.

## License

Port for learning/interoperability; respect upstream license for algorithms and data.
