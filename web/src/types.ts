export interface Campaign {
  id: number
  ruleset_id: number
  name: string
  description: string
  active: boolean
  created_at: string
}

export interface Character {
  id: number
  campaign_id: number
  name: string
  data_json: string
  portrait_path: string
  created_at: string
}

export interface Session {
  id: number
  campaign_id: number
  title: string
  date: string
  summary: string
  notes: string
  created_at: string
}

export interface Message {
  id: number
  session_id: number
  role: string
  content: string
  created_at: string
  whisper?: boolean
}

export interface SessionNPC {
  id: number
  session_id: number
  name: string
  note: string
  created_at: string
}

export interface CombatEncounter {
  id: number
  session_id: number
  name: string
  active: boolean
  active_turn_index: number
  created_at: string
}

export interface Combatant {
  id: number
  encounter_id: number
  character_id: number | null
  name: string
  initiative: number
  hp_current: number
  hp_max: number
  conditions_json: string
  is_player: boolean
}

export interface CombatSnapshot {
  encounter: CombatEncounter
  combatants: Combatant[]
}

export interface GameContext {
  campaign: Campaign | null
  character: Character | null
  session: Session | null
  recent_messages: Message[]
  active_combat: CombatSnapshot | null
}

export interface WorldNote {
  id: number
  campaign_id: number
  title: string
  content: string
  category: string
  tags_json: string
  personality_json: string
  created_at: string
}

export interface DiceRoll {
  id: number
  session_id: number
  expression: string
  result: number
  breakdown_json: string
  created_at: string
}

export interface TimelineEntry {
  type: 'message' | 'dice_roll' | 'world_note_event' | 'combat_event'
  timestamp: string
  data: Record<string, unknown>
}

export interface Objective {
  id: number
  campaign_id: number
  title: string
  description: string
  status: 'active' | 'completed' | 'failed'
  parent_id: number | null
  created_at: string
}

export interface Item {
  id: number
  character_id: number
  name: string
  description: string
  quantity: number
  equipped: boolean
  created_at: string
}

export interface XPEntry {
  id: number
  session_id: number
  note: string
  amount: number | null
  created_at: string
}
