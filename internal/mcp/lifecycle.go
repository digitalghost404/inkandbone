package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/digitalghost404/inkandbone/internal/api"
	"github.com/digitalghost404/inkandbone/internal/db"
	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleCreateCampaign(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	rulesetName, ok := reqStr(req, "ruleset")
	if !ok {
		return mcplib.NewToolResultError("ruleset is required"), nil
	}
	name, ok := reqStr(req, "name")
	if !ok {
		return mcplib.NewToolResultError("name is required"), nil
	}
	description := optStr(req, "description")

	rs, err := s.db.GetRulesetByName(rulesetName)
	if err != nil {
		return mcplib.NewToolResultError("db error: " + err.Error()), nil
	}
	if rs == nil {
		return mcplib.NewToolResultError(fmt.Sprintf("unknown ruleset %q — valid: dnd5e, ironsworn, vtm, coc, cyberpunk", rulesetName)), nil
	}

	campID, err := s.db.CreateCampaign(rs.ID, name, description)
	if err != nil {
		return mcplib.NewToolResultError("create campaign: " + err.Error()), nil
	}
	if err := s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)); err != nil {
		return mcplib.NewToolResultError("set active campaign: " + err.Error()), nil
	}

	s.bus.Publish(api.Event{Type: api.EventCampaignCreated, Payload: map[string]any{"campaign_id": campID, "name": name}})
	return mcplib.NewToolResultText(fmt.Sprintf("campaign %d created and activated: %s (%s)", campID, name, rulesetName)), nil
}

func (s *Server) handleListCampaigns(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	campaigns, err := s.db.ListCampaigns()
	if err != nil {
		return mcplib.NewToolResultError("db error: " + err.Error()), nil
	}
	if campaigns == nil {
		campaigns = []db.Campaign{}
	}
	b, err := json.Marshal(campaigns)
	if err != nil {
		return mcplib.NewToolResultError("marshal error: " + err.Error()), nil
	}
	return mcplib.NewToolResultText(string(b)), nil
}

func (s *Server) handleCreateCharacter(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	name, ok := reqStr(req, "name")
	if !ok {
		return mcplib.NewToolResultError("name is required"), nil
	}

	var campID int64
	if id, ok := optInt64(req, "campaign_id"); ok && id > 0 {
		campID = id
	} else {
		var err error
		campID, err = s.activeCampaignID()
		if err != nil {
			return mcplib.NewToolResultError(err.Error()), nil
		}
	}

	charID, err := s.db.CreateCharacter(campID, name)
	if err != nil {
		return mcplib.NewToolResultError("create character: " + err.Error()), nil
	}
	if err := s.db.SetSetting("active_character_id", strconv.FormatInt(charID, 10)); err != nil {
		return mcplib.NewToolResultError("set active character: " + err.Error()), nil
	}

	msg := fmt.Sprintf("character %d created and activated: %s", charID, name)

	if optBool(req, "random_stats") {
		camp, err := s.db.GetCampaign(campID)
		if err != nil || camp == nil {
			return mcplib.NewToolResultError("could not load campaign for random_stats"), nil
		}
		rs, err := s.db.GetRuleset(camp.RulesetID)
		if err != nil || rs == nil {
			return mcplib.NewToolResultError("could not load ruleset for random_stats"), nil
		}
		var schema struct {
			System string `json:"system"`
		}
		if err := json.Unmarshal([]byte(rs.SchemaJSON), &schema); err != nil {
			return mcplib.NewToolResultError("invalid ruleset schema: " + err.Error()), nil
		}
		stats := rollStats(schema.System)
		if len(stats) > 0 {
			dataJSON, err := json.Marshal(stats)
			if err != nil {
				return mcplib.NewToolResultError("marshal stats: " + err.Error()), nil
			}
			if err := s.db.UpdateCharacterData(charID, string(dataJSON)); err != nil {
				return mcplib.NewToolResultError("save stats: " + err.Error()), nil
			}
			msg += fmt.Sprintf(" (stats rolled for %s)", schema.System)
		}
	}

	s.bus.Publish(api.Event{Type: api.EventCharacterCreated, Payload: map[string]any{"character_id": charID, "name": name}})
	return mcplib.NewToolResultText(msg), nil
}

func (s *Server) handleListCharacters(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	var campID int64
	if id, ok := optInt64(req, "campaign_id"); ok && id > 0 {
		campID = id
	} else {
		var err error
		campID, err = s.activeCampaignID()
		if err != nil {
			return mcplib.NewToolResultError(err.Error()), nil
		}
	}

	characters, err := s.db.ListCharacters(campID)
	if err != nil {
		return mcplib.NewToolResultError("db error: " + err.Error()), nil
	}
	if characters == nil {
		characters = []db.Character{}
	}
	b, err := json.Marshal(characters)
	if err != nil {
		return mcplib.NewToolResultError("marshal error: " + err.Error()), nil
	}
	return mcplib.NewToolResultText(string(b)), nil
}

func (s *Server) handleListSessions(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	var campID int64
	if id, ok := optInt64(req, "campaign_id"); ok && id > 0 {
		campID = id
	} else {
		var err error
		campID, err = s.activeCampaignID()
		if err != nil {
			return mcplib.NewToolResultError(err.Error()), nil
		}
	}

	sessions, err := s.db.ListSessions(campID)
	if err != nil {
		return mcplib.NewToolResultError("db error: " + err.Error()), nil
	}
	if sessions == nil {
		sessions = []db.Session{}
	}
	b, err := json.Marshal(sessions)
	if err != nil {
		return mcplib.NewToolResultError("marshal error: " + err.Error()), nil
	}
	return mcplib.NewToolResultText(string(b)), nil
}
