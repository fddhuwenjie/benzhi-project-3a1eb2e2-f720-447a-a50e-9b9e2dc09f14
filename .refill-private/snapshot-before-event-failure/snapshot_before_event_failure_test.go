package snapshot_before_event_failure_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/persistence"
)

func TestFailedEventAppendDoesNotPublishSnapshot(t *testing.T) {
	caseID := "case-atomic-failure"
	base := t.TempDir()
	directory := longDataDirectory(t, base, 4090-len(filepath.Join("/cases", caseID+".json")))
	store, err := persistence.Open(directory)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	now := time.Date(2026, 8, 26, 9, 30, 0, 0, time.UTC)
	item, err := domain.NewCase(domain.CreateInput{
		ID: caseID, Site: "L-7", Reason: "retirement", OwnerID: "owner-1", PlannedDate: "2026-09-01",
		Materials: []domain.ControlledMaterial{{MaterialCode: "MAT-7", DisplayName: "sample", HazardClass: "general", DeclaredQuantity: 1, Unit: "vial", PackageCondition: "intact", DisposalMethod: "neutralize"}},
	}, now)
	if err != nil {
		t.Fatalf("create case: %v", err)
	}
	event, err := audit.NewEvent(item.ID, "case_created", "admin-1", "request-atomic-failure", item.Revision, now, map[string]string{"site": item.Site}, "")
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	record := application.IdempotencyRecord{RequestID: "request-atomic-failure", CaseID: item.ID, Operation: "create", Revision: item.Revision}
	if err := store.Commit(context.Background(), 0, item, event, record); err == nil {
		t.Fatal("expected event append failure")
	}
	if leaked, err := store.Load(context.Background(), item.ID); err == nil {
		t.Fatalf("commit failure leaked snapshot without audit event: revision=%d", leaked.Revision)
	}
}

func longDataDirectory(t *testing.T, base string, targetLength int) string {
	t.Helper()
	path := base
	for len(path) < targetLength {
		componentLength := targetLength - len(path) - 1
		if componentLength > 200 {
			componentLength = 200
		}
		path = filepath.Join(path, strings.Repeat("d", componentLength))
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatalf("create long data directory: %v", err)
	}
	return path
}
