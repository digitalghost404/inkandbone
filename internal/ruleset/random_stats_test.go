package ruleset

import (
	"strings"
	"testing"
)

func TestRollWrathGloryStats_abilitiesPrePopulated(t *testing.T) {
	// Roll many times to ensure various archetypes get tested
	for i := 0; i < 200; i++ {
		stats := rollWrathGloryStats("")
		archetype, _ := stats["archetype"].(string)
		if archetype == "" {
			t.Fatal("archetype field is empty")
		}
		def, ok := wgArchetypes[archetype]
		if !ok {
			t.Fatalf("archetype %q not found in wgArchetypes", archetype)
		}
		talents, _ := stats["talents"].(string)
		for _, ability := range def.abilities {
			if !strings.Contains(talents, ability) {
				t.Errorf("archetype %q: expected %q in talents field %q",
					archetype, ability, talents)
			}
		}
	}
}

func TestWGArchetypeAbilities_knownArchetypes(t *testing.T) {
	tests := []struct {
		archetype string
		abilities []string
	}{
		{"Imperial Guardsman", []string{"Look Out, Sir!"}},
		{"Sanctioned Psyker", []string{"Psyker", "Unlock Disciplines"}},
		{"Tactical Space Marine", []string{"Tactical Versatility"}},
		{"Chaos Space Marine", []string{"Tactical Versatility", "Corruption"}},
		{"Tech-Priest", []string{"Rite of Repair"}},
		{"Heretek", []string{"Rite of Repair", "Corruption"}},
		{"Inquisitor", []string{"Unchecked Authority"}},
	}
	for _, tt := range tests {
		def, ok := wgArchetypes[tt.archetype]
		if !ok {
			t.Errorf("archetype %q not found", tt.archetype)
			continue
		}
		abilitySet := map[string]bool{}
		for _, a := range def.abilities {
			abilitySet[a] = true
		}
		for _, want := range tt.abilities {
			if !abilitySet[want] {
				t.Errorf("archetype %q missing ability %q (got %v)", tt.archetype, want, def.abilities)
			}
		}
	}
}

func TestRollVtMStats_V5Fields(t *testing.T) {
	for i := 0; i < 10; i++ {
		stats := RollStats("vtm", "")
		for _, key := range []string{
			"hunger", "blood_potency", "bane_severity", "humanity", "stains",
			"strength", "dexterity", "stamina",
			"charisma", "manipulation", "composure",
			"intelligence", "wits", "resolve",
		} {
			if _, ok := stats[key]; !ok {
				t.Errorf("missing field %q", key)
			}
		}
		if stats["hunger"] != 1 {
			t.Errorf("hunger should be 1, got %v", stats["hunger"])
		}
		if stats["humanity"] != 7 {
			t.Errorf("humanity should be 7, got %v", stats["humanity"])
		}
		stamina, _ := stats["stamina"].(int)
		healthMax, _ := stats["health_max"].(int)
		if healthMax != stamina+3 {
			t.Errorf("health_max should be stamina+3=%d, got %d", stamina+3, healthMax)
		}
		composure, _ := stats["composure"].(int)
		resolve, _ := stats["resolve"].(int)
		willpowerMax, _ := stats["willpower_max"].(int)
		if willpowerMax != composure+resolve {
			t.Errorf("willpower_max should be composure+resolve=%d, got %d", composure+resolve, willpowerMax)
		}
	}
}
