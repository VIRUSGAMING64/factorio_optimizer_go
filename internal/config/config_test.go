package config

import "testing"

func TestLoadRedGreen(t *testing.T) {
	cfg, err := Load("../../data/red_green.json")
	if err != nil {
		cfg, err = Load("data/red_green.json")
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Targets) != 2 {
		t.Fatalf("targets %d", len(cfg.Targets))
	}
	if cfg.Targets[0].ResolvedRate() != 0.5 {
		t.Fatalf("rate %v", cfg.Targets[0].ResolvedRate())
	}
	if cfg.BeltTier != "fast" {
		t.Fatalf("belt %s", cfg.BeltTier)
	}
}

func TestTrickDefaults(t *testing.T) {
	cfg := &RunConfiguration{}
	cfg.ApplyDefaults()
	if !cfg.Trick("I1") {
		t.Fatal("I1 should default on")
	}
	cfg.Tricks = map[string]bool{"I1": false}
	if cfg.Trick("I1") {
		t.Fatal("override off")
	}
}

func TestDefaultsTuning(t *testing.T) {
	d := Defaults()
	if d.SAIterations < 1000 {
		t.Fatal("iters")
	}
}
