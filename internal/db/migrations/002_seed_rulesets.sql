INSERT OR IGNORE INTO rulesets (name, schema_json, version) VALUES
  ('dnd5e',     '{"system":"dnd5e","fields":["race","class","level","hp","ac","str","dex","con","int","wis","cha","proficiency_bonus","skills","inventory","spells","features"]}', '5e'),
  ('ironsworn', '{"system":"ironsworn","fields":["edge","heart","iron","shadow","wits","health","spirit","supply","momentum","vows","bonds","assets","notes"]}', '1.0'),
  ('vtm',       '{"system":"vtm","fields":["clan","generation","humanity","blood_pool","willpower","attributes","abilities","disciplines","virtues","backgrounds","notes"]}', 'V20'),
  ('coc',       '{"system":"coc","fields":["occupation","age","hp","sanity","luck","mp","str","con","siz","dex","app","int","pow","edu","skills","inventory","notes"]}', '7e'),
  ('cyberpunk', '{"system":"cyberpunk","fields":["role","int","ref","cool","tech","lk","att","ma","emp","body","humanity","eurodollars","skills","cyberware","gear","notes"]}', 'Red');
