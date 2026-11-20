// Command optimize is the entry point of the Factorio layout optimizer (Go port).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/blocks"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/config"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/data"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/emit"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/place"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/plan"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/power"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/route"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/search"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/validate"
	"github.com/VIRUSGAMING64/factorio_optimizer_go/internal/viz"
)

func main() {
	configPath := flag.String("config", "data/example_config.json", "job configuration JSON")
	tuningPath := flag.String("tuning", "data/tuning.json", "tuning knobs JSON")
	dataPath := flag.String("data", "data/recipes.json", "Factorio data (recipes.json)")
	outDir := flag.String("out", "", "output directory (default: solves/<timestamp>)")
	printTuning := flag.Bool("print-tuning", false, "print tuning and exit")
	saIters := flag.Int("sa-iters", 30000, "simulated annealing iterations")
	attempts := flag.Int("attempts", 1, "place+route trials to try")
	flag.Parse()

	fmt.Println("Factorio Optimizer (Go port)")
	fmt.Println("============================")

	tuning, err := config.LoadTuning(*tuningPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: tuning: %v (using defaults)\n", err)
		tuning = config.Defaults()
	}
	if *printTuning {
		fmt.Printf("%+v\n", tuning)
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	if cfg.Data != "" {
		*dataPath = cfg.Data
	}

	root, err := data.LoadRootData(*dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "data: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded game data %s – %d recipes, %d items, %d entities\n",
		root.GameVersion, len(root.Recipes), len(root.Items), len(root.Entities))

	if *outDir == "" {
		stamp := time.Now().Format("2006-01-02_15-04-05")
		*outDir = filepath.Join("solves", stamp+"-sa")
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	// ---- Stage 1: Plan ----
	fmt.Println("\nStage 1: building production plan…")
	t0 := time.Now()
	p, err := plan.Build(root, cfg, tuning)
	if err != nil {
		fmt.Fprintf(os.Stderr, "plan: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  → %s  (%.2fs)\n", p.Summary(), time.Since(t0).Seconds())
	writePlanSummary(*outDir, cfg, p)

	areaFactor := tuning.AreaFactor
	if areaFactor < 1.1 {
		areaFactor = 1.4
	}
	iters := *saIters
	if tuning.SAIterations > 0 && *saIters == 30000 {
		iters = tuning.SAIterations
	}
	beltTier := cfg.BeltTier
	if beltTier == "" {
		beltTier = "fast"
	}

	maxM := 20
	if cfg.Blocks != nil && cfg.Blocks.MaxMachines > 0 {
		maxM = cfg.Blocks.MaxMachines
	}

	var pl place.Placement
	var rt route.Result

	if len(p.Machines) > maxM {
		// ---- Factory of blocks ----
		fmt.Printf("\nStage 2–3: factory of blocks (%d machines > max %d)…\n", len(p.Machines), maxM)
		t1 := time.Now()
		fr, err := blocks.BuildFactory(p, cfg, tuning, maxM, iters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "factory: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  → %s  (%.2fs)\n", fr.Message, time.Since(t1).Seconds())
		pl = place.Placement{W: fr.Floorplan.W, H: fr.Floorplan.H}
		rt = route.Result{Success: true, Entities: fr.Entities, Message: fr.Message}
		for _, e := range fr.Entities {
			switch e.Kind {
			case route.KindBelt:
				rt.Belts++
			case route.KindUndergroundIn, route.KindUndergroundOut:
				rt.Undergrounds++
			case route.KindInserter:
				rt.Inserters++
			}
		}
		_ = os.WriteFile(filepath.Join(*outDir, "factory.txt"), []byte(fr.Message+"\n"), 0o644)
	} else if *attempts > 1 {
		fmt.Printf("\nStage 2–3: %d place+route trials…\n", *attempts)
		t1 := time.Now()
		trial, err := search.RunTrials(p, search.Options{
			Attempts: *attempts, SAIters: iters, AreaFactor: areaFactor,
			BeltTier: beltTier, Tuning: tuning,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "search: %v\n", err)
			os.Exit(1)
		}
		pl, rt = trial.Placement, trial.Route
		fmt.Printf("  → best score=%.1f box %d×%d  %s  (%.2fs)\n",
			trial.Score, pl.W, pl.H, rt.Message, time.Since(t1).Seconds())
		writePlacement(*outDir, p, pl)
	} else {
		fmt.Println("\nStage 2: placing machines (simulated annealing)…")
		t1 := time.Now()
		pl, err = place.Place(p, areaFactor, iters, tuning.SAInitialTemp, tuning.SACoolingRate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "place: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  → box %d×%d, cost=%.1f  (%.2fs)\n", pl.W, pl.H, pl.Cost, time.Since(t1).Seconds())
		writePlacement(*outDir, p, pl)

		fmt.Println("\nStage 3: routing belts and inserters…")
		t2 := time.Now()
		rt = route.RouteMultiPass(p, pl, beltTier, 3)
		fmt.Printf("  → %s  (%.2fs)\n", rt.Message, time.Since(t2).Seconds())
	}

	// Optional beacon placement after routing (when enabled via tricks.beacons)
	if cfg.Tricks != nil && cfg.Tricks["beacons"] && len(rt.Entities) > 0 && pl.W > 0 {
		occupied := func(x, y int) bool {
			if x < 0 || y < 0 || x >= pl.W || y >= pl.H {
				return true
			}
			for _, e := range rt.Entities {
				if (e.Kind == route.KindMachine || e.Kind == route.KindPole || e.Kind == route.KindBeacon || e.Kind == route.KindBelt) && e.X == x && e.Y == y {
					return true
				}
			}
			return false
		}
		rects := make([]power.MachineRect, 0, len(pl.Machines))
		for _, pm := range pl.Machines {
			mi := p.Machines[pm.ID]
			rects = append(rects, power.MachineRect{X: pm.X, Y: pm.Y, W: mi.W, H: mi.H})
		}
		bps := power.PlaceBeaconGrid(pl.W, pl.H, rects, occupied, nil)
		for _, b := range bps {
			rt.Entities = append(rt.Entities, route.Entity{
				Name: data.Beacon,
				Kind: route.KindBeacon,
				X:    b.X,
				Y:    b.Y,
				Item: "speed-module-3",
			})
		}
		if len(bps) > 0 {
			fmt.Printf("  → placed %d beacons\n", len(bps))
		}
	}

	// ---- Stage 4: Validate ----
	fmt.Println("\nStage 4: validating…")
	rep := validate.Check(p, pl, rt)
	fmt.Print("  ", rep.String())
	rates := validate.CheckPlanRates(p, cfg, root)
	fmt.Print("  ", rates.String())
	layout := validate.CheckLayout(rt.Entities)
	fmt.Print("  ", layout.String())
	lrates := validate.CheckLayoutRates(p, rt.Entities, beltTier)
	fmt.Print("  ", lrates.String())
	detail := validate.CheckLayoutDetail(rt.Entities, pl.W, pl.H)
	fmt.Print("  ", detail.String())
	ins := validate.CheckInserters(rt.Entities, false)
	fmt.Print("  ", ins.String())
	valText := rep.String() + "\n" + rates.String() + "\n" + layout.String() + "\n" + lrates.String() + "\n" + detail.String() + "\n" + ins.String()
	_ = os.WriteFile(filepath.Join(*outDir, "validate.txt"), []byte(valText), 0o644)
	// ASCII layout for quick inspection
	if pl.W > 0 && pl.H > 0 && pl.W <= 120 && pl.H <= 120 {
		ascii := viz.RenderASCII(pl.W, pl.H, rt.Entities)
		_ = os.WriteFile(filepath.Join(*outDir, "layout.txt"), []byte(ascii), 0o644)
		fmt.Println("  → layout.txt written")
	}

	// ---- Stage 5: Emit blueprint ----
	fmt.Println("\nStage 5: emitting blueprint…")
	bp, err := emit.FromRoute(rt.Entities, "Go optimizer – "+filepath.Base(*outDir))
	if err != nil {
		fmt.Fprintf(os.Stderr, "emit: %v\n", err)
	} else {
		bpPath := filepath.Join(*outDir, "blueprint.txt")
		if err := os.WriteFile(bpPath, []byte(bp), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write blueprint: %v\n", err)
		} else {
			fmt.Printf("  → %s (%d bytes)\n", bpPath, len(bp))
			if js, err := emit.DecodeToJSON(bp); err == nil {
				_ = os.WriteFile(filepath.Join(*outDir, "blueprint.json"), []byte(js), 0o644)
			}
		}
	}

	fmt.Println("\nDone.")
}

func writePlanSummary(outDir string, cfg *config.RunConfiguration, p *plan.Plan) {
	path := filepath.Join(outDir, "plan_summary.txt")
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write plan: %v\n", err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "Plan summary – %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "Targets:\n")
	for _, t := range cfg.Targets {
		fmt.Fprintf(f, "  %s @ %.4f/s\n", t.ResolvedItem(), t.ResolvedRate())
	}
	fmt.Fprintf(f, "\nMachines (%d):\n", len(p.Machines))
	for _, m := range p.Machines {
		fmt.Fprintf(f, "  [%d] %s  recipe=%s  %dx%d  crafts=%.3f/s\n",
			m.ID, m.Entity, m.Recipe, m.W, m.H, m.Crafts)
	}
	fmt.Fprintf(f, "\nPorts:\n")
	for _, port := range p.Ports {
		dir := "OUT"
		if port.IsInput {
			dir = "IN "
		}
		fmt.Fprintf(f, "  %s  %-28s  %.4f/s  side=%s\n", dir, port.Item, port.Rate, port.Side)
	}
	fmt.Printf("  plan → %s\n", path)
}

func writePlacement(outDir string, p *plan.Plan, pl place.Placement) {
	path := filepath.Join(outDir, "placement.txt")
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write placement: %v\n", err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "Placement – box %d×%d  cost=%.2f\n\n", pl.W, pl.H, pl.Cost)
	for _, m := range pl.Machines {
		mi := p.Machines[m.ID]
		fmt.Fprintf(f, "  [%d] %s (%s) at (%d,%d)  %dx%d\n",
			m.ID, mi.Entity, mi.Recipe, m.X, m.Y, mi.W, mi.H)
	}
	fmt.Fprintf(f, "\nASCII preview:\n%s", pl.ASCII(p))
	fmt.Printf("  placement → %s\n", path)
}
