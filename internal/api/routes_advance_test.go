package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoSuggestXPSpend_noopForCoC(t *testing.T) {
	s := newTestServer(t)

	rs, err := s.db.GetRulesetByName("coc")
	require.NoError(t, err)
	require.NotNil(t, rs)

	campID, err := s.db.CreateCampaign(rs.ID, "CoC Campaign", "")
	require.NoError(t, err)
	charID, err := s.db.CreateCharacter(campID, "Investigator")
	require.NoError(t, err)
	_ = s.db.UpdateCharacterData(charID, `{"xp":50}`)

	char, err := s.db.GetCharacter(charID)
	require.NoError(t, err)

	ch := s.bus.Subscribe()

	// Should no-op immediately for CoC.
	go s.autoSuggestXPSpend(1, charID, char, rs, map[string]any{"xp": float64(50)}, 50)

	// Wait briefly — no xp_spend_suggestions event should arrive.
	select {
	case ev := <-ch:
		assert.NotEqual(t, EventXPSpendSuggestions, ev.Type,
			"CoC should not emit xp_spend_suggestions")
	case <-time.After(200 * time.Millisecond):
		// correct: nothing emitted
	}
}

func TestAutoSuggestXPSpend_sessionCap(t *testing.T) {
	s := newTestServer(t)
	const sessionID = int64(42)

	// Pre-fill the cap.
	s.xpSuggestCounts.Store(sessionID, 20)

	rs, err := s.db.GetRulesetByName("wrath_glory")
	require.NoError(t, err)
	require.NotNil(t, rs)

	campID, err := s.db.CreateCampaign(rs.ID, "WG Campaign", "")
	require.NoError(t, err)
	charID, err := s.db.CreateCharacter(campID, "Brother Cato")
	require.NoError(t, err)
	_ = s.db.UpdateCharacterData(charID, `{"xp":50}`)

	char, err := s.db.GetCharacter(charID)
	require.NoError(t, err)

	ch := s.bus.Subscribe()

	// Session cap reached — should no-op.
	go s.autoSuggestXPSpend(sessionID, charID, char, rs, map[string]any{"xp": float64(50)}, 50)

	select {
	case ev := <-ch:
		assert.NotEqual(t, EventXPSpendSuggestions, ev.Type,
			"capped session should not emit xp_spend_suggestions")
	case <-time.After(200 * time.Millisecond):
		// correct: nothing emitted
	}
}
