package create_lookup_error_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/persistence"
)

func TestCreatePropagatesLookupCorruption(t *testing.T) {
	directory := t.TempDir()
	store, err := persistence.Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	corruptPath := filepath.Join(directory, "cases", "zzz-corrupt.json")
	if err := os.WriteFile(corruptPath, []byte("{broken snapshot"), 0600); err != nil {
		t.Fatal(err)
	}

	service := application.NewService(store)
	_, err = service.Create(context.Background(), application.CreateCommand{
		CommandMeta: application.CommandMeta{ActorID: "admin", Role: "safety_admin", RequestID: "create-corrupt-lookup"},
		Site:        "实验室 A",
		Reason:      "到期退役",
		OwnerID:     "owner-1",
		PlannedDate: "2030-01-01",
		Materials: []domain.ControlledMaterial{{
			MaterialCode: "MAT-1", DisplayName: "受控材料", HazardClass: "general",
			DeclaredQuantity: 1, Unit: "瓶", PackageCondition: "intact", DisposalMethod: "neutralize",
		}},
	})
	if err == nil {
		t.Fatalf("expected create to reject corrupted request lookup")
	}
	if got := err.Error(); got == "" {
		t.Fatalf("expected a diagnostic persistence error, got %v", err)
	}
}
