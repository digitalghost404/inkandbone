package ruleset

import (
	"strings"
	"testing"
)

func TestVtMPredatorTypes_allTenDefined(t *testing.T) {
	types := []string{
		"Alleycat", "Bagger", "Blood Leech", "Cleaner", "Consensualist",
		"Extortionist", "Graverobber", "Osiris", "Sandman", "Siren",
	}
	for _, name := range types {
		if _, ok := vtmPredatorTypes[name]; !ok {
			t.Errorf("predator type %q not defined", name)
		}
	}
}

func TestVtMPredatorTypes_eachHasTwoDisciplines(t *testing.T) {
	for name, pt := range vtmPredatorTypes {
		if len(pt.disciplines) != 2 {
			t.Errorf("%q: expected 2 discipline grants, got %d", name, len(pt.disciplines))
		}
	}
}

func TestApplyVtMPredatorType_sirenGrantsPresenceAndPotence(t *testing.T) {
	stats := map[string]any{
		"presence": 0,
		"potence":  0,
		"skill_specialties": "",
		"merits_flaws":      "",
	}
	ApplyVtMPredatorType("Siren", stats)
	if stats["presence"] != 1 {
		t.Errorf("expected presence=1, got %v", stats["presence"])
	}
	if stats["potence"] != 1 {
		t.Errorf("expected potence=1, got %v", stats["potence"])
	}
	if !strings.Contains(stats["skill_specialties"].(string), "Seduction") {
		t.Errorf("expected Seduction specialty, got %v", stats["skill_specialties"])
	}
}

func TestApplyVtMPredatorType_unknownType_noOp(t *testing.T) {
	stats := map[string]any{"presence": 2}
	ApplyVtMPredatorType("Unknown", stats)
	if stats["presence"] != 2 {
		t.Errorf("unknown type should not modify stats")
	}
}
