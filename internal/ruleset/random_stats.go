package ruleset

import (
	"math/rand"
	"sort"
)

// randPick returns a random element from the given slice.
func randPick(options []string) string {
	return options[rand.Intn(len(options))]
}

// RollStats returns a map of randomly generated starting field values for the
// given ruleset system. Every schema field is populated — text fields are
// randomly chosen from canonical options. An unrecognised system returns an
// empty map.
func RollStats(system string) map[string]any {
	switch system {
	case "dnd5e":
		return map[string]any{
			"race":              randPick([]string{"Human", "Elf", "Dwarf", "Halfling", "Gnome", "Half-Elf", "Half-Orc", "Tiefling", "Dragonborn"}),
			"class":             randPick([]string{"Barbarian", "Bard", "Cleric", "Druid", "Fighter", "Monk", "Paladin", "Ranger", "Rogue", "Sorcerer", "Warlock", "Wizard"}),
			"background":        randPick([]string{"Acolyte", "Charlatan", "Criminal", "Entertainer", "Folk Hero", "Guild Artisan", "Hermit", "Noble", "Outlander", "Sage", "Sailor", "Soldier", "Urchin"}),
			"alignment":         "True Neutral",
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
			"clan":        randPick([]string{"Brujah", "Gangrel", "Malkavian", "Nosferatu", "Toreador", "Tremere", "Ventrue", "Lasombra", "Tzimisce", "Assamite", "Giovanni", "Ravnos", "Setite"}),
			"generation":  13,
			"humanity":    7,
			"blood_pool":  10,
			"willpower":   3,
			"attributes":  "",
			"abilities":   "",
			"disciplines": "",
			"virtues":     "",
			"backgrounds": "",
			"notes":       "",
		}
	case "coc":
		pow := rollNd(3, 6) * 5
		siz := (rollNd(2, 6) + 6) * 5
		con := rollNd(3, 6) * 5
		return map[string]any{
			"occupation": randPick([]string{"Antiquarian", "Artist", "Author", "Detective", "Doctor", "Engineer", "Journalist", "Military Officer", "Occultist", "Parapsychologist", "Police Inspector", "Professor", "Thief"}),
			"str":        rollNd(3, 6) * 5,
			"con":        con,
			"siz":        siz,
			"dex":        rollNd(3, 6) * 5,
			"app":        rollNd(3, 6) * 5,
			"pow":        pow,
			"int":        (rollNd(2, 6) + 6) * 5,
			"edu":        (rollNd(2, 6) + 6) * 5,
			"hp":         (con/10 + siz/10 + 1) / 1,
			"sanity":     pow,
			"luck":       rollNd(3, 6) * 5,
			"mp":         pow / 10,
			"age":        17 + rollNd(2, 6),
			"skills":     "",
			"inventory":  "",
			"notes":      "",
		}
	case "cyberpunk", "cyberpunk_red":
		emp := rollNd(2, 6) + 2
		return map[string]any{
			"role":        randPick([]string{"Rockerboy", "Solo", "Netrunner", "Tech", "Medtech", "Media", "Cop", "Corporate", "Fixer", "Nomad"}),
			"int":         rollNd(2, 6) + 2,
			"ref":         rollNd(2, 6) + 2,
			"cool":        rollNd(2, 6) + 2,
			"tech":        rollNd(2, 6) + 2,
			"lk":          rollNd(2, 6) + 2,
			"att":         rollNd(2, 6) + 2,
			"ma":          rollNd(2, 6) + 2,
			"emp":         emp,
			"body":        rollNd(2, 6) + 2,
			"humanity":    emp * 10,
			"eurodollars": rollNd(2, 6)*100 + 200,
			"skills":      "",
			"cyberware":   "",
			"gear":        "",
			"notes":       "",
		}
	case "shadowrun":
		bod := rollNd(1, 6)
		agi := rollNd(1, 6)
		rea := rollNd(1, 6)
		str := rollNd(1, 6)
		wil := rollNd(1, 6)
		log := rollNd(1, 6)
		intu := rollNd(1, 6)
		cha := rollNd(1, 6)
		return map[string]any{
			"metatype":       randPick([]string{"Human", "Elf", "Dwarf", "Ork", "Troll"}),
			"archetype":      randPick([]string{"Street Samurai", "Adept", "Decker", "Technomancer", "Rigger", "Mage", "Shaman", "Face", "Infiltrator", "Fixer"}),
			"priority":       "A/B/C/D/E",
			"body":           bod,
			"agility":        agi,
			"reaction":       rea,
			"strength":       str,
			"willpower":      wil,
			"logic":          log,
			"intuition":      intu,
			"charisma":       cha,
			"edge":           rollNd(1, 4),
			"essence":        6,
			"physical_limit": (str*2+bod+rea)/3 + 1,
			"mental_limit":   (log*2+intu+wil)/3 + 1,
			"social_limit":   (cha*2+wil+10/3) / 3,
			"nuyen":          rollNd(2, 6) * 100,
			"karma":          0,
			"reputation":     0,
			"notoriety":      0,
			"notes":          "",
		}
	case "wfrp":
		return map[string]any{
			"species":      randPick([]string{"Human", "Halfling", "Dwarf", "High Elf", "Wood Elf"}),
			"career":       randPick([]string{"Apothecary", "Engineer", "Lawyer", "Physician", "Scholar", "Wizard", "Agitator", "Artisan", "Beggar", "Investigator", "Merchant", "Rat Catcher", "Soldier", "Thief", "Entertainer", "Messenger", "Soldier", "Scout"}),
			"career_level": 1,
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
			"ambitions":    "",
			"notes":        "",
		}
	case "starwars":
		return starWarsStats()

	case "l5r":
		clan := randPick([]string{"Crab", "Crane", "Dragon", "Lion", "Mantis", "Phoenix", "Scorpion", "Unicorn"})
		familyByClan := map[string][]string{
			"Crab":     {"Hida", "Hiruma", "Kaiu", "Kuni", "Toritaka", "Yasuki"},
			"Crane":    {"Asahina", "Daidoji", "Doji", "Kakita"},
			"Dragon":   {"Agasha", "Hitomi", "Kitsuki", "Mirumoto", "Tamori"},
			"Lion":     {"Akodo", "Ikoma", "Kitsu", "Matsu"},
			"Mantis":   {"Kamoto", "Moshi", "Tsuruchi", "Yoritomo"},
			"Phoenix":  {"Agasha", "Asako", "Isawa", "Shiba"},
			"Scorpion": {"Bayushi", "Shosuro", "Soshi", "Yogo"},
			"Unicorn":  {"Horiuchi", "Iuchi", "Moto", "Shinjo", "Utaku"},
		}
		return map[string]any{
			"clan":        clan,
			"family":      randPick(familyByClan[clan]),
			"school":      clan + " School",
			"school_rank": 1,
			"air":         2,
			"earth":       2,
			"fire":        2,
			"water":       2,
			"void":        1,
			"endurance":   8,
			"composure":   8,
			"focus":       3,
			"vigilance":   2,
			"glory":       45,
			"honor":       45,
			"status":      30,
			"xp":          0,
			"notes":       "",
		}
	case "theonering":
		body := rollNd(1, 3) + 1
		heart := rollNd(1, 3) + 1
		wits := rollNd(1, 3) + 1
		return map[string]any{
			"culture":          randPick([]string{"Bardings", "Beornings", "Dwarves of Erebor", "Elves of Mirkwood", "Hobbits of the Shire", "Men of Bree", "Rangers of the North", "Woodmen of Wilderland"}),
			"calling":          randPick([]string{"Scholar", "Slayer", "Treasure Hunter", "Wanderer", "Warden"}),
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
			"standing":         randPick([]string{"Emissary", "Exile", "Honoured", "Renowned", "Strider"}),
			"fellowship_score": 18,
			"notes":            "",
		}
	case "wrath_glory":
		str := rollNd(1, 3) + 3
		agi := rollNd(1, 3) + 3
		tgh := rollNd(1, 3) + 3
		itl := rollNd(1, 3) + 3
		wil := rollNd(1, 3) + 3
		fel := rollNd(1, 3) + 3
		archetype := randPick([]string{
			"Adeptus Astartes", "Adeptus Mechanicus", "Astra Militarum",
			"Inquisitorial Agent", "Rogue Trader", "Ministorum Priest",
			"Sanctioned Psyker", "Adepta Sororitas", "Commissar",
			"Arbitrator", "Astropath", "Navigator", "Ogryn",
			"Ratling", "Voidmaster", "Death Cult Assassin",
			"Heretic", "Chaos Space Marine", "Cultist",
		})
		faction := randPick([]string{
			"Imperium of Man", "Adeptus Mechanicus", "Inquisition",
			"Adepta Sororitas", "Astra Militarum", "Adeptus Astartes",
			"Rogue Traders", "Officio Assassinorum",
			"Chaos Undivided", "Nurgle", "Tzeentch", "Khorne", "Slaanesh",
			"Death Guard", "Thousand Sons", "World Eaters", "Emperor's Children",
		})
		return map[string]any{
			"archetype": archetype,
			"faction":   faction,
			"rank":      rollNd(1, 3),
			"keywords":  faction,

			"strength":  str,
			"agility":   agi,
			"toughness": tgh,
			"intellect": itl,
			"willpower": wil,
			"fellowship": fel,

			// Skills: start at 0; a few randomly bumped to 1 for variety
			"ws":              wrathSkillRoll(),
			"bs":              wrathSkillRoll(),
			"athletics":       wrathSkillRoll(),
			"awareness":       wrathSkillRoll(),
			"cunning":         wrathSkillRoll(),
			"deception":       wrathSkillRoll(),
			"fortitude":       wrathSkillRoll(),
			"insight":         wrathSkillRoll(),
			"intimidation":    wrathSkillRoll(),
			"investigation":   wrathSkillRoll(),
			"leadership":      wrathSkillRoll(),
			"medicae":         wrathSkillRoll(),
			"persuasion":      wrathSkillRoll(),
			"pilot":           wrathSkillRoll(),
			"psychic_mastery": 0,
			"scholar":         wrathSkillRoll(),
			"stealth":         wrathSkillRoll(),
			"survival":        wrathSkillRoll(),
			"tech":            wrathSkillRoll(),

			// Derived combat values
			"initiative":    agi,
			"speed":         agi,
			"defence":       agi,
			"resilience":    tgh,
			"determination": tgh + wil,
			"resolve":       wil,
			"conviction":    wil,

			"wounds":     tgh * 2,
			"shock":      wil + rollNd(1, 3),
			"corruption": 0,

			"wrath":  0,
			"glory":  0,
			"ruin":   0,
			"wealth": 2,
			"xp":     0,

			"talents": "",
			"powers":  "",
			"notes":   "",
		}
	case "blades":
		return bladesStats()

	case "paranoia":
		sector := randPick([]string{"ALF", "BRT", "DEN", "GRN", "MEL", "PLC", "RED", "VLT"})
		return map[string]any{
			"full_name":          "",
			"sector":             sector,
			"security_clearance": "INFRARED",
			"management_style":   randPick([]string{"Authoritarian", "Bureaucratic", "Paranoid", "Obsequious"}),
			"power_group":        randPick([]string{"Armed Forces", "CPU", "HPD&MC", "IntSec", "PLC", "R&D", "Tech Services"}),
			"secret_society":     randPick([]string{"Anti-Mutant", "Communists", "FCCC-P", "Free Enterprise", "Humanists", "Mystics", "Pro Tech", "Romantics"}),
			"violence":           rollNd(1, 6),
			"treachery":          rollNd(1, 6),
			"happiness":          rollNd(1, 6),
			"straight":           rollNd(1, 6),
			"moxie":              rollNd(1, 6),
			"credits":            rollNd(1, 6)*10 + 20,
			"clone_number":       1,
			"treason_points":     0,
			"notes":              "",
		}
	default:
		return map[string]any{}
	}
}

// wrathSkillRoll returns a starting skill rating for W&G: 70% chance of 0,
// 25% chance of 1, 5% chance of 2 — representing an untrained recruit with
// a few areas of natural aptitude.
func wrathSkillRoll() int {
	r := rand.Intn(20)
	if r < 1 {
		return 2
	}
	if r < 6 {
		return 1
	}
	return 0
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
		"vows":     "",
		"bonds":    "",
		"assets":   "",
		"notes":    "",
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
	career := randPick([]string{"Bounty Hunter", "Colonist", "Explorer", "Hired Gun", "Mystic", "Smuggler", "Technician"})
	specializationByCareeer := map[string][]string{
		"Bounty Hunter": {"Assassin", "Gadgeteer", "Survivalist"},
		"Colonist":      {"Doctor", "Politico", "Scholar"},
		"Explorer":      {"Fringer", "Scout", "Trader"},
		"Hired Gun":     {"Bodyguard", "Marauder", "Mercenary Soldier"},
		"Mystic":        {"Advisor", "Magician", "Seer"},
		"Smuggler":      {"Charmer", "Gambler", "Pilot"},
		"Technician":    {"Mechanic", "Outlaw Tech", "Slicer"},
	}
	result := map[string]any{
		"species":          randPick([]string{"Human", "Twi'lek", "Rodian", "Wookiee", "Bothan", "Mon Calamari", "Trandoshan", "Duros", "Zabrak", "Togruta"}),
		"career":           career,
		"specialization":   randPick(specializationByCareeer[career]),
		"wounds_current":   0,
		"wounds_threshold": 10 + stats["brawn"],
		"strain_current":   0,
		"strain_threshold": 10 + stats["willpower"],
		"soak":             stats["brawn"],
		"defense_melee":    0,
		"defense_ranged":   0,
		"credits":          500,
		"obligation":       10 + rollNd(1, 10),
		"force_rating":     0,
		"notes":            "",
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
		"playbook":    randPick([]string{"Cutter", "Hound", "Leech", "Lurk", "Slide", "Spider", "Whisper"}),
		"heritage":    randPick([]string{"Akoros", "The Dagger Isles", "Iruvia", "Severos", "Skovlan", "Tycheros"}),
		"background":  randPick([]string{"Academic", "Labor", "Law", "Trade", "Military", "Underworld"}),
		"vice":        randPick([]string{"Faith", "Gambling", "Leisure", "Obligation", "Pleasure", "Stupor", "Weird"}),
		"stress":      0,
		"trauma":      0,
		"coin":        2,
		"stash":       0,
		"load":        3,
		"xp_insight":  0,
		"xp_prowess":  0,
		"xp_resolve":  0,
		"notes":       "",
	}
	for k, v := range vals {
		result[k] = v
	}
	return result
}
