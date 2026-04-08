-- 028_vtm_oracle_compulsion.sql: VtM-specific oracle tables and clan Compulsion tables
-- Action and Theme oracles replace the generic ones for VtM sessions.
-- Clan Compulsion tables (10 entries each) fire on Messy Critical results.

-- VtM Action oracle (ruleset-specific, rolls 1-50)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'action', 1, 2, 'Hunt' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 3, 4, 'Feed' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 5, 6, 'Deceive' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 7, 8, 'Dominate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 9, 10, 'Seduce' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 11, 12, 'Betray' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 13, 14, 'Protect' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 15, 16, 'Flee' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 17, 18, 'Investigate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 19, 20, 'Embrace' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 21, 22, 'Manipulate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 23, 24, 'Observe' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 25, 26, 'Infiltrate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 27, 28, 'Negotiate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 29, 30, 'Confront' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 31, 32, 'Diablerize' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 33, 34, 'Summon' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 35, 36, 'Conceal' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 37, 38, 'Reveal' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 39, 40, 'Stalk' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 41, 42, 'Escape' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 43, 44, 'Claim' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 45, 46, 'Surrender' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 47, 48, 'Expose' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 49, 50, 'Endure' FROM rulesets WHERE name = 'vtm';

-- VtM Theme oracle (ruleset-specific, rolls 1-50)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'theme', 1, 2, 'The Beast' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 3, 4, 'The Masquerade' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 5, 6, 'Elysium' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 7, 8, 'Clan Politics' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 9, 10, 'The Prince' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 11, 12, 'A Coterie Rival' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 13, 14, 'Hunger' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 15, 16, 'A Touchstone' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 17, 18, 'A Sire' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 19, 20, 'An Elder' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 21, 22, 'A Masquerade Breach' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 23, 24, 'Blood Bond' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 25, 26, 'A Mortal Witness' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 27, 28, 'The Sheriff' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 29, 30, 'Diablerie' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 31, 32, 'A Hidden Secret' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 33, 34, 'Paranoia' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 35, 36, 'Manipulation' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 37, 38, 'Ambition' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 39, 40, 'Desire' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 41, 42, 'Humanity' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 43, 44, 'The Rack' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 45, 46, 'An Old Enemy' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 47, 48, 'A New Threat' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 49, 50, 'Redemption' FROM rulesets WHERE name = 'vtm';

-- Brujah Compulsion: Rebellion (table_name = 'compulsion_brujah', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_brujah', 1, 1, 'You openly contradict the most powerful person in the room, loudly and publicly.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 2, 2, 'You refuse a direct order, even from an ally, on principle.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 3, 3, 'You destroy something that symbolizes authority — a badge, a door, a throne.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 4, 4, 'You side with whoever is being oppressed, regardless of the facts.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 5, 5, 'You start a fight with the most dominant figure present.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 6, 6, 'You loudly enumerate every injustice you have witnessed tonight.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 7, 7, 'You demand an explanation for every rule you are expected to follow.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 8, 8, 'You refuse to be the first to back down from any confrontation.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 9, 9, 'You champion a stranger as an act of defiance against their oppressor.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 10, 10, 'You announce that the current power structure is corrupt and must fall.' FROM rulesets WHERE name = 'vtm';

-- Gangrel Compulsion: Feral Impulse (table_name = 'compulsion_gangrel', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_gangrel', 1, 1, 'You drop to all fours and sniff the ground, tracking a scent only you can sense.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 2, 2, 'You circle the room slowly, marking territory in your mind.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 3, 3, 'You growl audibly at anyone who steps too close.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 4, 4, 'You crouch rather than sit. Standing feels wrong.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 5, 5, 'You find the nearest exit and position yourself near it, unable to relax otherwise.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 6, 6, 'You snap your teeth at the nearest mortal who speaks to you.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 7, 7, 'You eat something raw — meat, vermin, whatever is available.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 8, 8, 'You refuse to enter any building — the outdoors is the only safe place.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 9, 9, 'You track a target across the room on instinct before catching yourself.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 10, 10, 'You produce a low, rumbling territorial growl whenever a stranger approaches.' FROM rulesets WHERE name = 'vtm';

-- Malkavian Compulsion: Delusion (table_name = 'compulsion_malkavian', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_malkavian', 1, 1, 'You are convinced someone in the room is not who they claim to be — a spy, a demon, or an impostor.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 2, 2, 'You believe an inanimate object in the room is speaking to you and must be obeyed.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 3, 3, 'You are certain tonight is a night you have lived before — an exact repetition.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 4, 4, 'You believe you are being watched by someone invisible and act accordingly.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 5, 5, 'You are convinced that one specific person is the key to preventing a catastrophe.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 6, 6, 'You believe the numbers in the room have deep significance and must be counted.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 7, 7, 'You are certain the current location is about to be destroyed and urge everyone to leave.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 8, 8, 'You become convinced you have already been betrayed tonight by a trusted ally.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 9, 9, 'You believe that if you speak above a whisper, something terrible will happen.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 10, 10, 'You are certain that the correct course of action is the exact opposite of what seems logical.' FROM rulesets WHERE name = 'vtm';

-- Nosferatu Compulsion: Cryptophilia (table_name = 'compulsion_nosferatu', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_nosferatu', 1, 1, 'You must learn where the most valuable person in the room sleeps during the day.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 2, 2, 'You must discover what the most powerful person present is hiding.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 3, 3, 'You must obtain a confession of some kind before the scene ends.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 4, 4, 'You must find out who in the room is lying about their identity.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 5, 5, 'You must learn what secret deal was recently struck between two parties present.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 6, 6, 'You must acquire physical proof of wrongdoing by someone present.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 7, 7, 'You must discover what weakness the nearest Elder possesses.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 8, 8, 'You must find out who sent someone here tonight and why.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 9, 9, 'You must learn the true name of a mortal who has interacted with the Kindred recently.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 10, 10, 'You must discover what illicit transaction is occurring or has recently occurred nearby.' FROM rulesets WHERE name = 'vtm';

-- Toreador Compulsion: Obsession (table_name = 'compulsion_toreador', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_toreador', 1, 1, 'A piece of music playing nearby transfixes you. You cannot willingly leave until it ends.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 2, 2, 'A mortal in the room possesses an almost supernatural physical perfection. You cannot stop watching them.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 3, 3, 'The architecture or decor of this location captivates you. You must examine every detail.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 4, 4, 'Someone''s voice is so beautiful that you cannot act while they are speaking.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 5, 5, 'A work of art — painting, sculpture, or photograph — demands your complete attention.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 6, 6, 'The way someone moves across the room is so elegant you must follow and observe.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 7, 7, 'A tragic story being told nearby is so affecting you cannot act until it concludes.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 8, 8, 'The interplay of light and shadow in this location arrests your attention completely.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 9, 9, 'Someone''s grief or joy is so raw and genuine that you are unable to look away.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 10, 10, 'A scent — perfume, blood, old paper — triggers a powerful aesthetic memory. You are lost in it.' FROM rulesets WHERE name = 'vtm';

-- Tremere Compulsion: Perfectionism (table_name = 'compulsion_tremere', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_tremere', 1, 1, 'Your last spoken statement contained an imprecision. You must correct it, in detail, immediately.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 2, 2, 'An action you recently took was suboptimal. You must attempt it again, properly.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 3, 3, 'Someone nearby is doing something incorrectly. You cannot proceed until you have corrected them.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 4, 4, 'The plan as stated has a flaw. You refuse to proceed until it is revised to your satisfaction.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 5, 5, 'You must restate your position using precisely the correct terminology, not approximations.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 6, 6, 'A tool or object is not in its correct place. You must correct this before anything else.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 7, 7, 'Your appearance is imperfect. You spend time correcting it even if the situation is urgent.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 8, 8, 'The outcome was acceptable but not excellent. You must explain how it could have been better.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 9, 9, 'An agreement is missing crucial specifics. You refuse to act on it until every term is defined.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 10, 10, 'A ritual or formula was performed incorrectly by someone nearby. You must perform it again, correctly.' FROM rulesets WHERE name = 'vtm';

-- Ventrue Compulsion: Arrogance (table_name = 'compulsion_ventrue', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_ventrue', 1, 1, 'You refuse to accept assistance from anyone who has not proven their worth to you.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 2, 2, 'You insist on leading, even in a domain that is clearly not yours.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 3, 3, 'You publicly dismiss the opinion of whoever speaks last, regardless of its merit.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 4, 4, 'You demand to be addressed by your full title before cooperating with anyone.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 5, 5, 'You will not share a resource or advantage with someone of lower station, even if it costs you.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 6, 6, 'You correct someone''s etiquette publicly, in detail, even if it is deeply inconvenient.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 7, 7, 'You take credit for a group success, attributing it to your leadership.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 8, 8, 'You refuse to negotiate as an equal with anyone you consider beneath you.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 9, 9, 'You delegate a task to a subordinate rather than perform it yourself, even urgently.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 10, 10, 'You make a unilateral decision affecting the group without consulting anyone.' FROM rulesets WHERE name = 'vtm';
