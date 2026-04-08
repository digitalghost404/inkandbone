// internal/db/queries_masquerade_test.go
package db

import (
	"testing"
)

func TestGetMasqueradeIntegrity_Default(t *testing.T) {
	db := newTestDB(t)
	campID := seedCampaign(t, db)
	sessID := seedSession(t, db, campID)

	integrity, err := db.GetMasqueradeIntegrity(sessID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integrity != 10 {
		t.Errorf("expected default integrity=10, got %d", integrity)
	}
}

func TestUpdateMasqueradeIntegrity_ClampMin(t *testing.T) {
	db := newTestDB(t)
	campID := seedCampaign(t, db)
	sessID := seedSession(t, db, campID)

	if err := db.UpdateMasqueradeIntegrity(sessID, -5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := db.GetMasqueradeIntegrity(sessID)
	if v != 0 {
		t.Errorf("expected clamp to 0, got %d", v)
	}
}

func TestUpdateMasqueradeIntegrity_ClampMax(t *testing.T) {
	db := newTestDB(t)
	campID := seedCampaign(t, db)
	sessID := seedSession(t, db, campID)

	if err := db.UpdateMasqueradeIntegrity(sessID, 15); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := db.GetMasqueradeIntegrity(sessID)
	if v != 10 {
		t.Errorf("expected clamp to 10, got %d", v)
	}
}
