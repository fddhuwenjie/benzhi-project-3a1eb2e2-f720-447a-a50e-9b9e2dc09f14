package stale_create_request_cache_test

import (
	"context"
	"testing"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/persistence"
)

func TestCreateRetryDoesNotReuseStaleNegativeLookup(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	service := application.NewService(store)
	command := application.CreateCommand{
		CommandMeta: application.CommandMeta{
			ActorID:   "safety-admin-1",
			Role:      "safety_admin",
			RequestID: "create-retry-1",
		},
		Site:        "A-101",
		Reason:      "到期退役",
		OwnerID:     "owner-1",
		PlannedDate: "2030-01-02",
		Materials: []domain.ControlledMaterial{{
			MaterialCode:     "MAT-001",
			DisplayName:      "受控试剂",
			HazardClass:      "general",
			DeclaredQuantity: 1,
			Unit:             "L",
			PackageCondition: "intact",
			DisposalMethod:   "neutralize",
		}},
	}

	first, err := service.Create(context.Background(), command)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	second, err := service.Create(context.Background(), command)
	if err != nil {
		t.Fatalf("retry create: %v", err)
	}

	if !second.Replayed || second.Case.ID != first.Case.ID {
		t.Fatalf("create retry bypassed persisted idempotency record: first=%s second=%s replayed=%v", first.Case.ID, second.Case.ID, second.Replayed)
	}
}
