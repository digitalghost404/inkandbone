package ruleset

import (
	"strings"
	"testing"
)

func TestRollWrathGloryStats_abilitiesPrePopulated(t *testing.T) {
	// Roll many times to ensure various archetypes get tested
	for i := 0; i < 200; i++ {
		stats := rollWrathGloryStats()
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
