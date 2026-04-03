package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleSearchRulebook(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	query := optStr(req, "query")
	if query == "" {
		return mcplib.NewToolResultError("query is required"), nil
	}

	rulesetID, ok := optInt64(req, "ruleset_id")
	if !ok {
		// Fall back to active campaign's ruleset
		campIDStr, err := s.db.GetSetting("active_campaign_id")
		if err != nil || campIDStr == "" {
			return mcplib.NewToolResultError("ruleset_id required or set active campaign first"), nil
		}
		var campID int64
		fmt.Sscanf(campIDStr, "%d", &campID)
		camp, err := s.db.GetCampaign(campID)
		if err != nil || camp == nil {
			return mcplib.NewToolResultError("active campaign not found"), nil
		}
		rulesetID = camp.RulesetID
	}

	chunks, err := s.db.SearchRulebookChunks(rulesetID, query)
	if err != nil {
		return mcplib.NewToolResultError("db: " + err.Error()), nil
	}

	b, err := json.Marshal(chunks)
	if err != nil {
		return mcplib.NewToolResultError("marshal: " + err.Error()), nil
	}
	return mcplib.NewToolResultText(string(b)), nil
}
