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

func TestApplyVtMPredatorType_Bagger(t *testing.T) {
	stats := map[string]any{"blood_sorcery": 0, "obfuscate": 0, "skill_specialties": "", "merits_flaws": ""}
	ApplyVtMPredatorType("Bagger", stats)
	if stats["blood_sorcery"] != 1 {
		t.Errorf("Bagger: expected blood_sorcery=1, got %v", stats["blood_sorcery"])
	}
	if stats["obfuscate"] != 1 {
		t.Errorf("Bagger: expected obfuscate=1, got %v", stats["obfuscate"])
	}
	if !strings.Contains(stats["skill_specialties"].(string), "Streetwise:Black Market") {
		t.Errorf("Bagger: expected Streetwise:Black Market specialty, got %v", stats["skill_specialties"])
	}
}

func TestApplyVtMPredatorType_BloodLeech(t *testing.T) {
	stats := map[string]any{"animalism": 0, "obfuscate": 0, "skill_specialties": "", "merits_flaws": ""}
	ApplyVtMPredatorType("Blood Leech", stats)
	if stats["animalism"] != 1 {
		t.Errorf("Blood Leech: expected animalism=1, got %v", stats["animalism"])
	}
	if stats["obfuscate"] != 1 {
		t.Errorf("Blood Leech: expected obfuscate=1, got %v", stats["obfuscate"])
	}
	if !strings.Contains(stats["skill_specialties"].(string), "Stealth:Stalking") {
		t.Errorf("Blood Leech: expected Stealth:Stalking specialty, got %v", stats["skill_specialties"])
	}
}

func TestApplyVtMPredatorType_Cleaner(t *testing.T) {
	stats := map[string]any{"auspex": 0, "dominate": 0, "skill_specialties": "", "merits_flaws": ""}
	ApplyVtMPredatorType("Cleaner", stats)
	if stats["auspex"] != 1 {
		t.Errorf("Cleaner: expected auspex=1, got %v", stats["auspex"])
	}
	if stats["dominate"] != 1 {
		t.Errorf("Cleaner: expected dominate=1, got %v", stats["dominate"])
	}
	if !strings.Contains(stats["skill_specialties"].(string), "Investigation:Crime Scenes") {
		t.Errorf("Cleaner: expected Investigation:Crime Scenes specialty, got %v", stats["skill_specialties"])
	}
}

func TestApplyVtMPredatorType_Extortionist(t *testing.T) {
	stats := map[string]any{"dominate": 0, "potence": 0, "skill_specialties": "", "merits_flaws": ""}
	ApplyVtMPredatorType("Extortionist", stats)
	if stats["dominate"] != 1 {
		t.Errorf("Extortionist: expected dominate=1, got %v", stats["dominate"])
	}
	if stats["potence"] != 1 {
		t.Errorf("Extortionist: expected potence=1, got %v", stats["potence"])
	}
	if !strings.Contains(stats["skill_specialties"].(string), "Intimidation:Coercion") {
		t.Errorf("Extortionist: expected Intimidation:Coercion specialty, got %v", stats["skill_specialties"])
	}
}
