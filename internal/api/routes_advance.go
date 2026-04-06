package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"

	ruleset "github.com/digitalghost404/inkandbone/internal/ruleset"
)

func (s *Server) handleAdvanceCharacter(w http.ResponseWriter, r *http.Request) {
	charID, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid character id", http.StatusBadRequest)
		return
	}

	var body struct {
		Field    string `json:"field"`
		NewValue int    `json:"new_value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Field == "" {
		http.Error(w, "field and new_value required", http.StatusBadRequest)
		return
	}

	// Load character.
	char, err := s.db.GetCharacter(charID)
	if err != nil || char == nil {
		http.NotFound(w, r)
		return
	}

	// Load campaign -> ruleset.
	camp, err := s.db.GetCampaign(char.CampaignID)
	if err != nil || camp == nil {
		http.Error(w, "campaign not found", http.StatusInternalServerError)
		return
	}
	rs, err := s.db.GetRuleset(camp.RulesetID)
	if err != nil || rs == nil {
		http.Error(w, "ruleset not found", http.StatusInternalServerError)
		return
	}
	system := rs.Name

	// Parse current stats.
	var stats map[string]any
	if char.DataJSON != "" && char.DataJSON != "{}" {
		if err := json.Unmarshal([]byte(char.DataJSON), &stats); err != nil {
			http.Error(w, "invalid stats JSON", http.StatusInternalServerError)
			return
		}
	}
	if stats == nil {
		stats = map[string]any{}
	}

	xpKey := ruleset.XPKey(system)
	currentXP := 0
	if v, ok := stats[xpKey].(float64); ok {
		currentXP = int(v)
	}

	field := body.Field
	newVal := body.NewValue

	// ── Validate ────────────────────────────────────────────────────────────

	if strings.HasPrefix(field, "talent:") {
		if system != "wrath_glory" {
			http.Error(w, "talent advances only supported for wrath_glory", http.StatusBadRequest)
			return
		}
		talentName := strings.TrimPrefix(field, "talent:")

		if !ruleset.WGTalentExists(talentName) {
			http.Error(w, "unknown or non-purchasable talent", http.StatusBadRequest)
			return
		}

		ownedStr, _ := stats["talents"].(string)
		archetypeName, _ := stats["archetype"].(string)
		if isWGTalentOwned(ownedStr, archetypeName, talentName) {
			http.Error(w, "talent already owned", http.StatusBadRequest)
			return
		}

		cost := ruleset.WGTalentCost(talentName)
		if currentXP < cost {
			http.Error(w, "not enough XP", http.StatusBadRequest)
			return
		}

	} else if system == "dnd5e" && field == "level" {
		currentLevel := 1
		if v, ok := stats["level"].(float64); ok {
			currentLevel = int(v)
		}
		if newVal != currentLevel+1 {
			http.Error(w, "new_value must be current level + 1", http.StatusBadRequest)
			return
		}
		dnd5eThresholds := []int{
			0, 300, 900, 2700, 6500, 14000, 23000, 34000,
			48000, 64000, 85000, 100000, 120000, 140000,
			165000, 195000, 225000, 265000, 305000, 355000,
		}
		if currentLevel >= len(dnd5eThresholds)-1 {
			http.Error(w, "already max level", http.StatusBadRequest)
			return
		}
		if currentXP < dnd5eThresholds[currentLevel] {
			http.Error(w, "not enough XP to level up", http.StatusBadRequest)
			return
		}

	} else {
		currentVal := 0
		if v, ok := stats[field].(float64); ok {
			currentVal = int(v)
		}
		if newVal != currentVal+1 {
			http.Error(w, "new_value must be current value + 1", http.StatusBadRequest)
			return
		}
		if system == "blades" {
			cost := ruleset.XPCostFor(system, field, newVal, "")
			if currentXP < cost {
				http.Error(w, fmt.Sprintf("not enough XP (need %d)", cost), http.StatusBadRequest)
				return
			}
		} else if system != "dnd5e" {
			cost := ruleset.XPCostFor(system, field, newVal, char.DataJSON)
			if currentXP < cost {
				http.Error(w, "not enough XP", http.StatusBadRequest)
				return
			}
		}
	}

	// ── Apply ────────────────────────────────────────────────────────────────

	switch {
	case strings.HasPrefix(field, "talent:"):
		talentName := strings.TrimPrefix(field, "talent:")
		cost := ruleset.WGTalentCost(talentName)
		stats[xpKey] = float64(currentXP - cost)
		ownedStr, _ := stats["talents"].(string)
		if ownedStr == "" {
			stats["talents"] = talentName
		} else {
			stats["talents"] = ownedStr + "|" + talentName
		}

	case system == "blades":
		stats[xpKey] = float64(0)
		currentVal := 0
		if v, ok := stats[field].(float64); ok {
			currentVal = int(v)
		}
		stats[field] = float64(currentVal + 1)

	case system == "dnd5e" && field == "level":
		currentLevel := 1
		if v, ok := stats["level"].(float64); ok {
			currentLevel = int(v)
		}
		newLevel := currentLevel + 1
		stats["level"] = float64(newLevel)
		currentHP := 0
		if v, ok := stats["hp"].(float64); ok {
			currentHP = int(v)
		}
		stats["hp"] = float64(currentHP + 5)
		stats["proficiency_bonus"] = float64(math.Floor(float64(newLevel-1)/4) + 2)

	default:
		cost := ruleset.XPCostFor(system, field, newVal, char.DataJSON)
		stats[xpKey] = float64(currentXP - cost)
		stats[field] = float64(newVal)

		if system == "wrath_glory" {
			ruleset.WGRecalcDerived(stats, field)
		}
	}

	// Persist.
	updated, err := json.Marshal(stats)
	if err != nil {
		http.Error(w, "marshal error", http.StatusInternalServerError)
		return
	}
	if err := s.db.UpdateCharacterData(charID, string(updated)); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	s.bus.Publish(Event{Type: EventCharacterUpdated, Payload: map[string]any{
		"id":       charID,
		"data_json": string(updated),
	}})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data_json": string(updated)}) //nolint:errcheck
}

// isWGTalentOwned returns true if talentName is in the pipe-delimited ownedStr
// OR is an archetype starting ability for archetypeName.
func isWGTalentOwned(ownedStr, archetypeName, talentName string) bool {
	for _, t := range strings.Split(ownedStr, "|") {
		if strings.TrimSpace(t) == talentName {
			return true
		}
	}
	if def, ok := ruleset.WGArchetypeDefFor(archetypeName); ok {
		for _, a := range def.Abilities() {
			if a == talentName {
				return true
			}
		}
	}
	return false
}
