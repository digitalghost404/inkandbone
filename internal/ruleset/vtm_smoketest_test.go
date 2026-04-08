package ruleset

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// 1. XPKey and XPLabel
// ---------------------------------------------------------------------------

func TestVtM_XPKey(t *testing.T) {
	if got := XPKey("vtm"); got != "xp" {
		t.Errorf("XPKey(vtm) = %q, want %q", got, "xp")
	}
}

func TestVtM_XPLabel(t *testing.T) {
	if got := XPLabel("vtm"); got != "XP" {
		t.Errorf("XPLabel(vtm) = %q, want %q", got, "XP")
	}
}

// ---------------------------------------------------------------------------
// 2. ValidFields — exhaustive field presence check
// ---------------------------------------------------------------------------

func TestVtM_ValidFields(t *testing.T) {
	fields := ValidFields("vtm")
	set := make(map[string]bool, len(fields))
	for _, f := range fields {
		set[f] = true
	}

	// 9 attributes
	attributes := []string{
		"strength", "dexterity", "stamina",
		"charisma", "manipulation", "composure",
		"intelligence", "wits", "resolve",
	}
	for _, a := range attributes {
		if !set[a] {
			t.Errorf("ValidFields(vtm) missing attribute %q", a)
		}
	}

	// 36 skills  (counts: 9 physical, 9 social, 9 mental = 27 but V5 has specific list)
	skills := []string{
		// Physical
		"athletics", "brawl", "craft", "drive", "firearms", "larceny", "melee",
		"stealth", "survival",
		// Social
		"animal_ken", "etiquette", "insight", "intimidation", "leadership",
		"performance", "persuasion", "streetwise", "subterfuge",
		// Mental
		"academics", "awareness", "finance", "investigation", "medicine",
		"occult", "politics", "technology",
	}
	for _, sk := range skills {
		if !set[sk] {
			t.Errorf("ValidFields(vtm) missing skill %q", sk)
		}
	}

	// 11 disciplines
	disciplines := []string{
		"animalism", "auspex", "blood_sorcery", "celerity", "dominate",
		"fortitude", "obfuscate", "oblivion", "potence", "presence", "protean",
	}
	for _, d := range disciplines {
		if !set[d] {
			t.Errorf("ValidFields(vtm) missing discipline %q", d)
		}
	}

	// blood_potency
	if !set["blood_potency"] {
		t.Error("ValidFields(vtm) missing blood_potency")
	}
}

// ---------------------------------------------------------------------------
// 3. XPCostFor — all field categories
// ---------------------------------------------------------------------------

func TestVtM_XPCostFor_Attribute(t *testing.T) {
	// Attribute: new_dots * 4
	if got := XPCostFor("vtm", "strength", 3, ""); got != 12 {
		t.Errorf("strength to 3: got %d, want 12", got)
	}
	if got := XPCostFor("vtm", "dexterity", 1, ""); got != 4 {
		t.Errorf("dexterity to 1: got %d, want 4", got)
	}
	if got := XPCostFor("vtm", "resolve", 5, ""); got != 20 {
		t.Errorf("resolve to 5: got %d, want 20", got)
	}
}

func TestVtM_XPCostFor_Skill(t *testing.T) {
	// Skill: new_dots * 3
	if got := XPCostFor("vtm", "athletics", 2, ""); got != 6 {
		t.Errorf("athletics to 2: got %d, want 6", got)
	}
	if got := XPCostFor("vtm", "occult", 4, ""); got != 12 {
		t.Errorf("occult to 4: got %d, want 12", got)
	}
}

func TestVtM_XPCostFor_InClanDiscipline_Brujah(t *testing.T) {
	// In-clan discipline for Brujah: potence is in-clan → new_dots * 5
	statsJSON := `{"clan":"Brujah","potence":1}`
	if got := XPCostFor("vtm", "potence", 2, statsJSON); got != 10 {
		t.Errorf("Brujah in-clan potence to 2: got %d, want 10", got)
	}
	// celerity is also in-clan for Brujah
	statsJSON2 := `{"clan":"Brujah","celerity":0}`
	if got := XPCostFor("vtm", "celerity", 1, statsJSON2); got != 5 {
		t.Errorf("Brujah in-clan celerity to 1: got %d, want 5", got)
	}
}

func TestVtM_XPCostFor_OutOfClanDiscipline_Brujah(t *testing.T) {
	// Out-of-clan discipline for Brujah: oblivion → new_dots * 7
	statsJSON := `{"clan":"Brujah","oblivion":1}`
	if got := XPCostFor("vtm", "oblivion", 2, statsJSON); got != 14 {
		t.Errorf("Brujah out-of-clan oblivion to 2: got %d, want 14", got)
	}
}

func TestVtM_XPCostFor_BloodPotency(t *testing.T) {
	// Blood Potency: new_dots * 10
	if got := XPCostFor("vtm", "blood_potency", 2, ""); got != 20 {
		t.Errorf("blood_potency to 2: got %d, want 20", got)
	}
	if got := XPCostFor("vtm", "blood_potency", 1, ""); got != 10 {
		t.Errorf("blood_potency to 1: got %d, want 10", got)
	}
}

func TestVtM_XPCostFor_Caitiff_AllDisciplinesOutOfClan(t *testing.T) {
	// Caitiff has no in-clan disciplines — all cost out-of-clan rate (7x)
	statsJSONCaitiff := `{"clan":"Caitiff","potence":1}`
	if got := XPCostFor("vtm", "potence", 2, statsJSONCaitiff); got != 14 {
		t.Errorf("Caitiff potence to 2: got %d, want 14 (out-of-clan)", got)
	}
}

func TestVtM_XPCostFor_CaseInsensitiveClanLookup(t *testing.T) {
	// lowercase "brujah" should give same result as "Brujah"
	lower := `{"clan":"brujah","potence":1}`
	upper := `{"clan":"Brujah","potence":1}`
	costLower := XPCostFor("vtm", "potence", 2, lower)
	costUpper := XPCostFor("vtm", "potence", 2, upper)
	if costLower != costUpper {
		t.Errorf("clan lookup case sensitivity: lower=%d upper=%d, want equal", costLower, costUpper)
	}
	if costLower != 10 {
		t.Errorf("brujah potence to 2: got %d, want 10", costLower)
	}
}

// ---------------------------------------------------------------------------
// 4. CanAffordAny
// ---------------------------------------------------------------------------

func TestVtM_CanAffordAny_SufficientXP(t *testing.T) {
	// VtM cheapest advance is in-clan discipline at 1*5=5 XP
	if !CanAffordAny("vtm", 5, `{"clan":"Brujah"}`) {
		t.Error("expected CanAffordAny=true for vtm with 5 XP")
	}
	if !CanAffordAny("vtm", 100, `{}`) {
		t.Error("expected CanAffordAny=true for vtm with 100 XP")
	}
}

func TestVtM_CanAffordAny_InsufficientXP(t *testing.T) {
	// 0 XP cannot afford anything (cheapest is a skill at rank 1 = 1*3 = 3 XP)
	if CanAffordAny("vtm", 0, `{}`) {
		t.Error("expected CanAffordAny=false for vtm with 0 XP")
	}
	// 2 XP cannot afford even the cheapest skill advance (3 XP)
	if CanAffordAny("vtm", 2, `{}`) {
		t.Error("expected CanAffordAny=false for vtm with 2 XP (min skill cost is 3)")
	}
}

// ---------------------------------------------------------------------------
// 5. CostRulesDescription
// ---------------------------------------------------------------------------

func TestVtM_CostRulesDescription(t *testing.T) {
	desc := CostRulesDescription("vtm")
	if desc == "" {
		t.Error("CostRulesDescription(vtm) returned empty string")
	}
}

// ---------------------------------------------------------------------------
// 6. FieldHints
// ---------------------------------------------------------------------------

func TestVtM_FieldHints(t *testing.T) {
	hints := FieldHints("vtm")
	if hints == "" {
		t.Error("FieldHints(vtm) returned empty string")
	}
	if !strings.Contains(hints, "potence") {
		t.Error("FieldHints(vtm) does not mention 'potence'")
	}
	if !strings.Contains(hints, "blood_potency") {
		t.Error("FieldHints(vtm) does not mention 'blood_potency'")
	}
	if !strings.Contains(hints, "strength") {
		t.Error("FieldHints(vtm) does not mention 'strength'")
	}
}

// ---------------------------------------------------------------------------
// 7. VtMInClanDisciplinesFor
// ---------------------------------------------------------------------------

func TestVtM_InClanDisciplinesFor_Brujah(t *testing.T) {
	discs, ok := VtMInClanDisciplinesFor("Brujah")
	if !ok {
		t.Fatal("VtMInClanDisciplinesFor(Brujah) returned ok=false")
	}
	expected := map[string]bool{"celerity": true, "potence": true, "presence": true}
	for _, d := range discs {
		expected[d] = true
	}
	// Verify all 3 are present
	got := map[string]bool{}
	for _, d := range discs {
		got[d] = true
	}
	for want := range map[string]bool{"celerity": true, "potence": true, "presence": true} {
		if !got[want] {
			t.Errorf("Brujah in-clan disciplines missing %q", want)
		}
	}
}

func TestVtM_InClanDisciplinesFor_CaseInsensitive(t *testing.T) {
	discsLower, okLower := VtMInClanDisciplinesFor("brujah")
	discsMixed, okMixed := VtMInClanDisciplinesFor("Brujah")
	if !okLower || !okMixed {
		t.Error("VtMInClanDisciplinesFor should recognize both 'brujah' and 'Brujah'")
	}
	if len(discsLower) != len(discsMixed) {
		t.Errorf("case mismatch: lower=%d mixed=%d disciplines", len(discsLower), len(discsMixed))
	}
}

func TestVtM_InClanDisciplinesFor_Caitiff(t *testing.T) {
	discs, ok := VtMInClanDisciplinesFor("Caitiff")
	if !ok {
		t.Fatal("VtMInClanDisciplinesFor(Caitiff) returned ok=false — Caitiff should be a recognized clan")
	}
	if len(discs) != 0 {
		t.Errorf("Caitiff should have 0 in-clan disciplines, got %d: %v", len(discs), discs)
	}
}

func TestVtM_InClanDisciplinesFor_Gangrel(t *testing.T) {
	discs, ok := VtMInClanDisciplinesFor("gangrel")
	if !ok {
		t.Fatal("gangrel not recognized")
	}
	got := map[string]bool{}
	for _, d := range discs {
		got[d] = true
	}
	for _, want := range []string{"animalism", "fortitude", "protean"} {
		if !got[want] {
			t.Errorf("Gangrel missing in-clan discipline %q", want)
		}
	}
}

func TestVtM_InClanDisciplinesFor_UnknownClan(t *testing.T) {
	_, ok := VtMInClanDisciplinesFor("Unknown Clan XYZ")
	if ok {
		t.Error("VtMInClanDisciplinesFor should return ok=false for unknown clan")
	}
}

// ---------------------------------------------------------------------------
// 8. ApplyVtMPredatorType
// ---------------------------------------------------------------------------

func TestVtM_ApplyPredatorType_Alleycat(t *testing.T) {
	stats := map[string]any{
		"celerity": float64(0),
		"potence":  float64(0),
	}
	ApplyVtMPredatorType("Alleycat", stats)

	if v, ok := stats["celerity"].(int); !ok || v != 1 {
		t.Errorf("celerity after Alleycat: got %v, want 1", stats["celerity"])
	}
	if v, ok := stats["potence"].(int); !ok || v != 1 {
		t.Errorf("potence after Alleycat: got %v, want 1", stats["potence"])
	}
	specialty, _ := stats["skill_specialties"].(string)
	if !strings.Contains(specialty, "Athletics") {
		t.Errorf("Alleycat specialty should mention Athletics, got %q", specialty)
	}
}

func TestVtM_ApplyPredatorType_StacksOnExistingDiscipline(t *testing.T) {
	// If character already has celerity 2, Alleycat bumps it to 3
	stats := map[string]any{
		"celerity": float64(2),
		"potence":  float64(1),
	}
	ApplyVtMPredatorType("Alleycat", stats)
	if v, ok := stats["celerity"].(int); !ok || v != 3 {
		t.Errorf("celerity after Alleycat: got %v, want 3", stats["celerity"])
	}
}

func TestVtM_ApplyPredatorType_UnknownNoOp(t *testing.T) {
	stats := map[string]any{"celerity": float64(1)}
	ApplyVtMPredatorType("NoSuchType", stats)
	if v, _ := stats["celerity"].(float64); v != 1 {
		t.Errorf("unknown predator type should be no-op, celerity changed to %v", stats["celerity"])
	}
}

func TestVtM_ApplyPredatorType_SpecialtyAppended(t *testing.T) {
	// If character already has specialties, Alleycat should append
	stats := map[string]any{
		"celerity":          float64(0),
		"potence":           float64(0),
		"skill_specialties": "Occult:Vampiric Lore",
	}
	ApplyVtMPredatorType("Alleycat", stats)
	specialty, _ := stats["skill_specialties"].(string)
	if !strings.Contains(specialty, "Occult:Vampiric Lore") {
		t.Errorf("Alleycat should not overwrite existing specialties, got %q", specialty)
	}
	if !strings.Contains(specialty, "Athletics:Brawling") {
		t.Errorf("Alleycat should add Athletics:Brawling specialty, got %q", specialty)
	}
}

func TestVtM_ApplyPredatorType_MeritsFlawsSet(t *testing.T) {
	stats := map[string]any{"auspex": float64(0), "dominate": float64(0)}
	ApplyVtMPredatorType("Cleaner", stats)
	mf, _ := stats["merits_flaws"].(string)
	if mf == "" {
		t.Error("Cleaner should set merits_flaws")
	}
}

// ---------------------------------------------------------------------------
// 9. All recognized clans have in-clan disciplines
// ---------------------------------------------------------------------------

func TestVtM_AllClansDefined(t *testing.T) {
	clans := []string{
		"brujah", "gangrel", "malkavian", "nosferatu", "toreador",
		"tremere", "ventrue", "lasombra", "tzimisce", "assamite",
		"giovanni", "ravnos", "setite", "caitiff",
	}
	for _, clan := range clans {
		discs, ok := VtMInClanDisciplinesFor(clan)
		if !ok {
			t.Errorf("clan %q not recognized by VtMInClanDisciplinesFor", clan)
			continue
		}
		if clan != "caitiff" && len(discs) != 3 {
			t.Errorf("clan %q: expected 3 in-clan disciplines, got %d: %v", clan, len(discs), discs)
		}
	}
}
