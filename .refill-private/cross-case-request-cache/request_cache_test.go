package crosscaserequestcache

import (
	"context"
	"testing"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/persistence"
)

func TestRequestCacheKeepsCaseScope(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := application.NewService(store)
	first := createCase(t, service, "create-first", "SITE-A")
	second := createCase(t, service, "create-second", "SITE-B")

	countCase(t, service, first.ID, "shared-count-request", "counter-a")
	result := countCase(t, service, second.ID, "shared-count-request", "counter-b")
	if result.Case.ID != second.ID || result.Replayed {
		t.Fatalf("cross-case request cache replayed case %s for case %s", result.Case.ID, second.ID)
	}
}

func createCase(t *testing.T, service *application.Service, requestID, site string) *domain.RetirementCase {
	t.Helper()
	result, err := service.Create(context.Background(), application.CreateCommand{
		CommandMeta: application.CommandMeta{ActorID: "admin", Role: "safety_admin", RequestID: requestID},
		Site:        site, Reason: "retirement", OwnerID: "owner", PlannedDate: "2099-01-01",
		Materials: []domain.ControlledMaterial{{MaterialCode: "MAT-01", DisplayName: "sample", HazardClass: "general", DeclaredQuantity: 1, Unit: "kg", PackageCondition: "intact", DisposalMethod: "approved"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return result.Case
}

func countCase(t *testing.T, service *application.Service, caseID, requestID, counterID string) *application.MutationResult {
	t.Helper()
	result, err := service.Count(context.Background(), application.CountCommand{
		CommandMeta: application.CommandMeta{ActorID: counterID, Role: "witness", RequestID: requestID, ExpectedRevision: 1},
		CaseID:      caseID, CounterID: counterID,
		Observations: []domain.CountObservation{{MaterialCode: "MAT-01", Quantity: 1, PackageCondition: "intact"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
