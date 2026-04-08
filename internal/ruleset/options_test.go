package ruleset

import "testing"

func TestVtMOptions_V5Clans(t *testing.T) {
	opts := CharacterOptions("vtm")
	clans, ok := opts["clan"]
	if !ok {
		t.Fatal("vtm options missing clan key")
	}
	want := []string{"Brujah", "Gangrel", "Malkavian", "Nosferatu", "Toreador", "Tremere", "Ventrue", "Caitiff", "Thin-Blooded"}
	if len(clans) != len(want) {
		t.Fatalf("expected %d clans, got %d: %v", len(want), len(clans), clans)
	}
	clanSet := map[string]bool{}
	for _, c := range clans {
		clanSet[c] = true
	}
	for _, w := range want {
		if !clanSet[w] {
			t.Errorf("missing clan %q", w)
		}
	}
}

func TestVtMOptions_PredatorType(t *testing.T) {
	opts := CharacterOptions("vtm")
	types, ok := opts["predator_type"]
	if !ok {
		t.Fatal("vtm options missing predator_type key")
	}
	if len(types) != 10 {
		t.Fatalf("expected 10 predator types, got %d", len(types))
	}
}

func TestVtMOptions_Sect(t *testing.T) {
	opts := CharacterOptions("vtm")
	if _, ok := opts["sect"]; !ok {
		t.Fatal("vtm options missing sect key")
	}
}
