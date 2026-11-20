package emit

import (
	"strings"
	"testing"
)

func TestBuildInserterRateKit(t *testing.T) {
	k, err := BuildInserterRateKit("fast-inserter")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(k.Blueprint, "0") {
		t.Fatal("bp")
	}
	if len(k.Entities) != 3 {
		t.Fatal(len(k.Entities))
	}
}

func TestBuildBeltThroughputKit(t *testing.T) {
	k, err := BuildBeltThroughputKit("fast-transport-belt", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(k.Entities) < 8 {
		t.Fatal("ents")
	}
}

func TestBuildUndergroundGapKit(t *testing.T) {
	k, err := BuildUndergroundGapKit("fast-underground-belt", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(k.Entities) < 4 {
		t.Fatal("ug pairs")
	}
}
