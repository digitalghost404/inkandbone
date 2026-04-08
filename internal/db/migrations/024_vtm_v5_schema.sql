-- 024_vtm_v5_schema.sql: Update VtM ruleset to V5 schema and fields
-- Updates the vtm ruleset row to V5 schema. Other rulesets untouched.
UPDATE rulesets SET
  schema_json = '{"system":"vtm","fields":[
    "clan","predator_type","sect","generation",
    "hunger","blood_potency","bane_severity","humanity","stains",
    "strength","dexterity","stamina",
    "charisma","manipulation","composure",
    "intelligence","wits","resolve",
    "athletics","brawl","craft","drive","firearms","larceny","melee","stealth","survival",
    "animal_ken","etiquette","insight","intimidation","leadership","performance","persuasion","streetwise","subterfuge",
    "academics","awareness","finance","investigation","medicine","occult","politics","technology",
    "animalism","auspex","blood_sorcery","celerity","dominate","fortitude","obfuscate","oblivion","potence","presence","protean",
    "health_max","health_superficial","health_aggravated",
    "willpower_max","willpower_superficial","willpower_aggravated",
    "skill_specialties","merits_flaws","convictions","touchstones","ambition","desire","notes"
  ]}',
  version = 'V5'
WHERE name = 'vtm';
