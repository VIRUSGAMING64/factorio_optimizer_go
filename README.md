# Factorio Optimizer (Go)

Port de [SeezoCode/factorio_optimizer_v2](https://github.com/SeezoCode/factorio_optimizer_v2) a Go.

Pipeline: **plan → place → route → validate → blueprint**

## Requisitos

- Go 1.22+

## Uso rápido (tras clonar)

```bash
git clone https://github.com/VIRUSGAMING64/factorio_optimizer_go.git
cd factorio_optimizer_go

# Demo red + green science (defaults)
go run .

# Con flags
go run . -config=data/red_green.json -sa-iters=30000 -attempts=2
go run . -config=data/config.json -sa-iters=20000

go test ./...
go build -o optimize .
./optimize -config=data/red_green.json
```

También: `make run` · `make test` · `make build`

### Flags

| Flag | Default | Descripción |
|------|---------|-------------|
| `-config` | `data/red_green.json` | Job (targets, máquinas) |
| `-data` | `data/recipes.json` | Base de datos Factorio |
| `-tuning` | `data/tuning.json` | Knobs SA/route |
| `-sa-iters` | `30000` | Iteraciones annealing |
| `-attempts` | `1` | Reintentos place+route |
| `-out` | `solves/<timestamp>-sa` | Carpeta de salida |

Salida: `blueprint.txt` (importable), `validation.txt`, `plan_summary.txt`.

## Estructura

```
main.go              ← go run .
cmd/optimize/        ← go run ./cmd/optimize
data/recipes.json    ← incluido en el repo
data/*.json          ← configs de ejemplo
internal/...
```

## Estado

~9k LOC, 116+ tests. PathFinder con lanes, SA con HPWL/bins/DI, factory-of-blocks, validación rates/layout/inserters/shape.
