package ruleset

import "testing"

func TestWGTalentExists(t *testing.T) {
	// Known purchasable talents must exist
	for _, name := range []string{"Iron Will", "Unshakeable Faith", "True Grit", "Rapid Reload"} {
		if !WGTalentExists(name) {
			t.Errorf("WGTalentExists(%q) = false, want true", name)
		}
	}
	// Starting archetype abilities are NOT purchasable
	for _, name := range []string{"Fiery Invective", "Loyal Compassion", "Psyker", "Tactical Versatility"} {
		if WGTalentExists(name) {
			t.Errorf("WGTalentExists(%q) = true, want false (archetype ability)", name)
		}
	}
	// Nonexistent talent
	if WGTalentExists("Made Up Talent XYZ") {
		t.Error("WGTalentExists nonexistent = true, want false")
	}
}

func TestWGTalentCost(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"Iron Will", 20},
		{"True Grit", 30},
		{"Warp Conduit", 50},
		{"Untouchable Talent", 60},
		{"Weapon Training (Las)", 10},
		{"Unknown Talent", 20}, // default
	}
	for _, tt := range tests {
		if got := WGTalentCost(tt.name); got != tt.want {
			t.Errorf("WGTalentCost(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}
