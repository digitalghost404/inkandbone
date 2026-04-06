package ruleset

// wgTalentTable maps purchasable W&G core-rulebook talent names to their XP cost.
// Default cost (for any talent not listed here) is 20 XP.
// Range: 10–60 XP. Archetype starting abilities are deliberately excluded.
var wgTalentTable = map[string]int{
	// ── 10 XP ────────────────────────────────────────────────────────────────
	"Weapon Training (Bolt)":             10,
	"Weapon Training (Chain)":            10,
	"Weapon Training (Flame)":            10,
	"Weapon Training (Las)":              10,
	"Weapon Training (Shock)":            10,
	"Weapon Training (Solid Projectile)": 10,
	"Forbidden Lore (Daemonology)":       10,
	"Forbidden Lore (Heresy)":            10,
	"Forbidden Lore (Mutants)":           10,
	"Forbidden Lore (Psykers)":           10,
	"Forbidden Lore (The Warp)":          10,
	"Forbidden Lore (Xenos)":             10,
	"Resistance (Fear)":                  10,
	"Resistance (Poison)":                10,
	"Resistance (Psychic Powers)":        10,
	"Peer (Adeptus Mechanicus)":          10,
	"Peer (Adeptus Ministorum)":          10,
	"Peer (Aeldari)":                     10,
	"Peer (Astra Militarum)":             10,
	"Peer (Criminal)":                    10,
	"Peer (Inquisition)":                 10,
	"Peer (Orks)":                        10,
	"Peer (Rogue Traders)":               10,
	"Peer (Underworld)":                  10,
	"Disturbing Voice":                   10,
	"Paranoia":                           10,

	// ── 15 XP ────────────────────────────────────────────────────────────────
	"Weapon Training (Melta)":     15,
	"Weapon Training (Plasma)":    15,
	"Weapon Training (Power)":     15,
	"Weapon Training (Launcher)":  15,
	"Weapon Training (Primitive)": 15,
	"Careful Aim":                 15,
	"Catfall":                     15,
	"Lay Low":                     15,
	"Quick Draw":                  15,
	"Rapid Reload":                15,
	"Double Team":                 15,
	"Hatred (Chaos)":              15,
	"Hatred (Greenskins)":         15,
	"Hatred (Heretics)":           15,
	"Hatred (Xenos)":              15,
	"Unarmed Warrior":             15,

	// ── 20 XP ────────────────────────────────────────────────────────────────
	"Acute Senses":                  20,
	"Ambidextrous":                  20,
	"Armour Proficiency (Flak)":     20,
	"Armour Proficiency (Mesh)":     20,
	"Blind Fighting":                20,
	"Blood of Heroes":               20,
	"Brutal Charge":                 20,
	"Bulging Biceps":                20,
	"Cold-Blooded":                  20,
	"Combat Formation":              20,
	"Combat Sense":                  20,
	"Counter Attack":                20,
	"Crushing Blow":                 20,
	"Dead Eye Shot":                 20,
	"Defensive Fighting":            20,
	"Deflect":                       20,
	"Die Hard":                      20,
	"Eye of Vengeance":              20,
	"Feel No Pain":                  20,
	"Frenzy":                        20,
	"Hard Target":                   20,
	"Heavy Hitter":                  20,
	"Hip Shooting":                  20,
	"Iron Jaw":                      20,
	"Iron Will":                     20,
	"Marksman":                      20,
	"Mechadendrite Use (Utility)":   20,
	"Mechadendrite Use (Weapon)":    20,
	"Nimble":                        20,
	"Nowhere to Hide":               20,
	"Overwhelming Blow":             20,
	"Practiced Aim":                 20,
	"Rapid Fire":                    20,
	"Reckless Charge":               20,
	"Shield Wall":                   20,
	"Sidestep":                      20,
	"Skirmisher":                    20,
	"Sprint":                        20,
	"Step Aside":                    20,
	"Sure Strike":                   20,
	"Takedown":                      20,
	"Tech-Use":                      20,
	"Unshakeable Faith":             20,
	"Never Say Die":                 20,
	"Power Behind the Blow":         20,
	"Armour Proficiency (Carapace)": 20,
	"Pack Tactics":                  20,
	"Savant":                        20,
	"Sixth Sense":                   20,
	"Wall of Steel":                 20,
	"Whirlwind of Death":            20,

	// ── 25 XP ────────────────────────────────────────────────────────────────
	"Demolition Expert":          25,
	"Dual Strike":                25,
	"Armour Proficiency (Power)": 25,
	"Mighty Blow":                25,

	// ── 30 XP ────────────────────────────────────────────────────────────────
	"Combat Master":      30,
	"Expert at Violence": 30,
	"Hotshot":            30,
	"Lightning Attack":   30,
	"Master Orator":      30,
	"Storm of Blows":     30,
	"Swift Attack":       30,
	"True Grit":          30,

	// ── 35 XP ────────────────────────────────────────────────────────────────
	"Hatred (Daemons)": 35,

	// ── 40 XP ────────────────────────────────────────────────────────────────
	"Banishment":          40,
	"Exceptional Leader":  40,
	"Preternatural Speed": 40,
	"Rite of Awe":         40,
	"Weapon Mastery":      40,
	"Will Not Die":        40,

	// ── 50 XP ────────────────────────────────────────────────────────────────
	"Warp Conduit": 50,

	// ── 60 XP ────────────────────────────────────────────────────────────────
	"Untouchable Talent": 60,
}

// WGTalentExists returns true if name is a known purchasable W&G talent.
// Archetype starting abilities (Fiery Invective, Psyker, etc.) return false.
func WGTalentExists(name string) bool {
	_, ok := wgTalentTable[name]
	return ok
}

// WGTalentCost returns the XP cost for a named W&G talent.
// Returns 20 (default) for unknown talent names.
func WGTalentCost(name string) int {
	if cost, ok := wgTalentTable[name]; ok {
		return cost
	}
	return 20
}
