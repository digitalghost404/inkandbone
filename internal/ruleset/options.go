package ruleset

// CharacterOptions returns a map of field key → possible values for every
// chooseable (text/select) field in a given ruleset system. Fields whose
// values depend on another field (e.g. l5r family depends on clan) are
// omitted. Returns nil for unknown or fully-numeric systems.
func CharacterOptions(system string) map[string][]string {
	switch system {
	case "dnd5e", "ironsworn":
		return nil // all numeric, nothing to choose

	case "vtm":
		return map[string][]string{
			"clan": {
				"Brujah", "Gangrel", "Malkavian", "Nosferatu", "Toreador",
				"Tremere", "Ventrue", "Lasombra", "Tzimisce", "Assamite",
				"Giovanni", "Ravnos", "Setite",
			},
		}

	case "coc":
		return map[string][]string{
			"occupation": {
				"Antiquarian", "Artist", "Author", "Detective", "Doctor",
				"Engineer", "Journalist", "Military Officer", "Occultist",
				"Parapsychologist", "Police Inspector", "Professor", "Thief",
			},
		}

	case "cyberpunk", "cyberpunk_red":
		return map[string][]string{
			"role": {
				"Rockerboy", "Solo", "Netrunner", "Tech", "Medtech",
				"Media", "Cop", "Corporate", "Fixer", "Nomad",
			},
		}

	case "shadowrun":
		return map[string][]string{
			"metatype": {"Human", "Elf", "Dwarf", "Ork", "Troll"},
		}

	case "wfrp":
		return map[string][]string{
			"species": {"Human", "Halfling", "Dwarf", "High Elf", "Wood Elf"},
			"career": {
				"Apothecary", "Engineer", "Lawyer", "Physician", "Scholar",
				"Wizard", "Agitator", "Artisan", "Beggar", "Investigator",
				"Merchant", "Rat Catcher", "Soldier", "Thief", "Entertainer",
				"Messenger", "Scout",
			},
		}

	case "starwars":
		return map[string][]string{
			"species": {
				"Human", "Twi'lek", "Rodian", "Wookiee", "Bothan",
				"Mon Calamari", "Trandoshan", "Duros", "Zabrak", "Togruta",
			},
			"career": {
				"Bounty Hunter", "Colonist", "Explorer", "Hired Gun",
				"Mystic", "Smuggler", "Technician",
			},
		}

	case "l5r":
		return map[string][]string{
			"clan": {
				"Crab", "Crane", "Dragon", "Lion",
				"Mantis", "Phoenix", "Scorpion", "Unicorn",
			},
		}

	case "theonering":
		return map[string][]string{
			"culture": {
				"Bardings", "Beornings", "Dwarves of Erebor",
				"Elves of Mirkwood", "Hobbits of the Shire",
				"Men of Bree", "Rangers of the North", "Woodmen of Wilderland",
			},
			"calling": {"Scholar", "Slayer", "Treasure Hunter", "Wanderer", "Warden"},
		}

	case "wrath_glory":
		return map[string][]string{
			"archetype": {
				"Adeptus Astartes", "Adeptus Mechanicus", "Astra Militarum",
				"Inquisitorial Agent", "Rogue Trader", "Ministorum Priest",
				"Sanctioned Psyker", "Heretic", "Chaos Space Marine", "Cultist",
			},
			"faction": {
				"Imperium of Man", "Adeptus Mechanicus", "Inquisition",
				"Rogue Traders", "Chaos Undivided", "Nurgle", "Tzeentch",
				"Khorne", "Slaanesh",
			},
		}

	case "blades":
		return map[string][]string{
			"playbook": {
				"Cutter", "Hound", "Leech", "Lurk", "Slide", "Spider", "Whisper",
			},
			"heritage": {
				"Akoros", "The Dagger Isles", "Iruvia",
				"Severos", "Skovlan", "Tycheros",
			},
			"background": {
				"Academic", "Labor", "Law", "Trade", "Military", "Underworld",
			},
			"vice": {
				"Faith", "Gambling", "Leisure", "Obligation",
				"Pleasure", "Stupor", "Weird",
			},
		}

	case "paranoia":
		return map[string][]string{
			"management_style": {
				"Authoritarian", "Bureaucratic", "Paranoid", "Obsequious",
			},
			"power_group": {
				"Armed Forces", "CPU", "HPD&MC", "IntSec",
				"PLC", "R&D", "Tech Services",
			},
			"secret_society": {
				"Anti-Mutant", "Communists", "FCCC-P", "Free Enterprise",
				"Humanists", "Mystics", "Pro Tech", "Romantics",
			},
		}

	default:
		return nil
	}
}
