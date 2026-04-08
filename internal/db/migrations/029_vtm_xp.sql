-- 029_vtm_xp.sql: Add xp field to VtM character schema
-- Allows vampires to accumulate and spend Beats/XP for advancement.
UPDATE rulesets SET
  schema_json = json_set(
    schema_json,
    '$.fields',
    json_array(
      'clan','predator_type','sect','generation',
      'hunger','blood_potency','bane_severity','humanity','stains',
      'xp',
      'strength','dexterity','stamina',
      'charisma','manipulation','composure',
      'intelligence','wits','resolve',
      'athletics','brawl','craft','drive','firearms','larceny','melee','stealth','survival',
      'animal_ken','etiquette','insight','intimidation','leadership','performance','persuasion','streetwise','subterfuge',
      'academics','awareness','finance','investigation','medicine','occult','politics','technology',
      'animalism','auspex','blood_sorcery','celerity','dominate','fortitude','obfuscate','oblivion','potence','presence','protean',
      'health_max','health_superficial','health_aggravated',
      'willpower_max','willpower_superficial','willpower_aggravated',
      'skill_specialties','merits_flaws','convictions','touchstones','ambition','desire','notes'
    )
  )
WHERE name = 'vtm';
