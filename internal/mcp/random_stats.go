package mcp

import (
	"math/rand"
	"sort"
)

// rollStats returns a map of randomly generated starting field values for the
// given ruleset system. Fields not present in the map are left empty by the
// caller. An unrecognised system returns an empty map.
func rollStats(system string) map[string]any {
	switch system {
	case "dnd5e":
		return map[string]any{
			"str":               roll4d6DropLowest(),
			"dex":               roll4d6DropLowest(),
			"con":               roll4d6DropLowest(),
			"int":               roll4d6DropLowest(),
			"wis":               roll4d6DropLowest(),
			"cha":               roll4d6DropLowest(),
			"level":             1,
			"hp":                10,
			"ac":                10,
			"proficiency_bonus": 2,
		}
	case "ironsworn":
		return ironswornStats()
	case "vtm":
		return map[string]any{
			"generation": 13,
			"humanity":   7,
			"blood_pool": 10,
			"willpower":  3,
		}
	case "coc":
		pow := rollNd(3, 6) * 5
		siz := (rollNd(2, 6) + 6) * 5
		con := rollNd(3, 6) * 5
		return map[string]any{
			"str":    rollNd(3, 6) * 5,
			"con":    con,
			"siz":    siz,
			"dex":    rollNd(3, 6) * 5,
			"app":    rollNd(3, 6) * 5,
			"pow":    pow,
			"int":    (rollNd(2, 6) + 6) * 5,
			"edu":    (rollNd(2, 6) + 6) * 5,
			"hp":     (con/10 + siz/10 + 1) / 1,
			"sanity": pow,
			"luck":   rollNd(3, 6) * 5,
			"mp":     pow / 10,
			"age":    17 + rollNd(2, 6),
		}
	case "cyberpunk_red":
		emp := rollNd(2, 6) + 2
		return map[string]any{
			"int":      rollNd(2, 6) + 2,
			"ref":      rollNd(2, 6) + 2,
			"cool":     rollNd(2, 6) + 2,
			"tech":     rollNd(2, 6) + 2,
			"lk":       rollNd(2, 6) + 2,
			"att":      rollNd(2, 6) + 2,
			"ma":       rollNd(2, 6) + 2,
			"emp":      emp,
			"body":     rollNd(2, 6) + 2,
			"humanity": emp * 10,
		}
	case "shadowrun":
		return map[string]any{
			"body":      rollNd(1, 6),
			"agility":   rollNd(1, 6),
			"reaction":  rollNd(1, 6),
			"strength":  rollNd(1, 6),
			"willpower": rollNd(1, 6),
			"logic":     rollNd(1, 6),
			"intuition": rollNd(1, 6),
			"charisma":  rollNd(1, 6),
			"edge":      rollNd(1, 4),
			"essence":   6,
		}
	case "wfrp":
		return map[string]any{
			"ws":           rollNd(2, 10) + 20,
			"bs":           rollNd(2, 10) + 20,
			"s":            rollNd(2, 10) + 20,
			"t":            rollNd(2, 10) + 20,
			"ag":           rollNd(2, 10) + 20,
			"i":            rollNd(2, 10) + 20,
			"dex":          rollNd(2, 10) + 20,
			"int":          rollNd(2, 10) + 20,
			"wp":           rollNd(2, 10) + 20,
			"fel":          rollNd(2, 10) + 20,
			"wounds":       10,
			"fate":         2,
			"fortune":      2,
			"resilience":   1,
			"resolve":      1,
			"xp":           0,
			"career_level": 1,
		}
	case "starwars":
		return starWarsStats()
	case "l5r":
		return map[string]any{
			"air":          2,
			"earth":        2,
			"fire":         2,
			"water":        2,
			"void":         1,
			"school_rank":  1,
			"endurance":    8,
			"composure":    8,
			"focus":        3,
			"vigilance":    2,
			"glory":        45,
			"honor":        45,
			"status":       30,
			"xp":           0,
		}
	case "theonering":
		body := rollNd(1, 3) + 1
		heart := rollNd(1, 3) + 1
		wits := rollNd(1, 3) + 1
		return map[string]any{
			"body":             body,
			"heart":            heart,
			"wits":             wits,
			"endurance":        20 + body,
			"endurance_max":    20 + body,
			"hope":             8 + heart,
			"hope_max":         8 + heart,
			"shadow_points":    0,
			"shadow_scars":     0,
			"valour":           1,
			"wisdom":           1,
			"fellowship_score": 18,
		}
	case "wrath_glory":
		agi := rollNd(1, 3) + 3
		tgh := rollNd(1, 3) + 3
		wil := rollNd(1, 3) + 3
		return map[string]any{
			"strength":      rollNd(1, 3) + 3,
			"agility":       agi,
			"toughness":     tgh,
			"intellect":     rollNd(1, 3) + 3,
			"willpower":     wil,
			"fellowship":    rollNd(1, 3) + 3,
			"initiative":    agi,
			"wounds":        10,
			"shock":         8,
			"resilience":    tgh,
			"determination": wil,
			"defence":       3,
			"resolve":       2,
			"conviction":    2,
			"wrath":         0,
			"glory":         0,
			"ruin":          0,
			"xp":            0,
		}
	case "blades":
		return bladesStats()
	case "paranoia":
		return map[string]any{
			"violence":           rollNd(1, 6),
			"treachery":          rollNd(1, 6),
			"happiness":          rollNd(1, 6),
			"straight":           rollNd(1, 6),
			"moxie":              rollNd(1, 6),
			"credits":            rollNd(1, 6)*10 + 20,
			"clone_number":       1,
			"treason_points":     0,
			"security_clearance": "INFRARED",
		}
	default:
		return map[string]any{}
	}
}

// roll4d6DropLowest rolls 4d6 and returns the sum of the top 3 (standard D&D method).
func roll4d6DropLowest() int {
	rolls := make([]int, 4)
	for i := range rolls {
		rolls[i] = rand.Intn(6) + 1
	}
	sort.Ints(rolls)
	return rolls[1] + rolls[2] + rolls[3]
}

// rollNd rolls n dice with the given number of sides and returns the total.
func rollNd(n, sides int) int {
	total := 0
	for i := 0; i < n; i++ {
		total += rand.Intn(sides) + 1
	}
	return total
}

// ironswornStats distributes 7 points among the five Ironsworn attributes (each 1–3).
func ironswornStats() map[string]any {
	attrs := [5]string{"edge", "heart", "iron", "shadow", "wits"}
	vals := [5]int{1, 1, 1, 1, 1} // base total = 5; distribute 2 more
	for remaining := 2; remaining > 0; {
		i := rand.Intn(5)
		if vals[i] < 3 {
			vals[i]++
			remaining--
		}
	}
	result := map[string]any{
		"health":   5,
		"spirit":   5,
		"supply":   5,
		"momentum": 2,
	}
	for i, k := range attrs {
		result[k] = vals[i]
	}
	return result
}

// starWarsStats generates Edge of the Empire characteristics for a typical humanoid.
// All six characteristics start at 2; three randomly chosen are raised to 3.
func starWarsStats() map[string]any {
	keys := []string{"brawn", "agility", "intellect", "cunning", "willpower", "presence"}
	stats := map[string]int{
		"brawn": 2, "agility": 2, "intellect": 2,
		"cunning": 2, "willpower": 2, "presence": 2,
	}
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	for _, k := range keys[:3] {
		stats[k]++
	}
	result := map[string]any{
		"wounds_current":   0,
		"wounds_threshold": 10 + stats["brawn"],
		"strain_current":   0,
		"strain_threshold": 10 + stats["willpower"],
		"soak":             stats["brawn"],
		"defense_melee":    0,
		"defense_ranged":   0,
		"credits":          500,
		"obligation":       10 + rollNd(1, 10),
	}
	for k, v := range stats {
		result[k] = v
	}
	return result
}

// bladesStats distributes 4 points among the 12 Blades in the Dark action ratings (each 0–2).
func bladesStats() map[string]any {
	actions := []string{
		"hunt", "study", "survey", "tinker",
		"finesse", "prowl", "skirmish", "wreck",
		"attune", "command", "consort", "sway",
	}
	vals := make(map[string]int, len(actions))
	for _, a := range actions {
		vals[a] = 0
	}
	for remaining := 4; remaining > 0; {
		a := actions[rand.Intn(len(actions))]
		if vals[a] < 2 {
			vals[a]++
			remaining--
		}
	}
	result := map[string]any{
		"stress":      0,
		"trauma":      0,
		"coin":        2,
		"stash":       0,
		"load":        3,
		"xp_insight":  0,
		"xp_prowess":  0,
		"xp_resolve":  0,
	}
	for k, v := range vals {
		result[k] = v
	}
	return result
}
