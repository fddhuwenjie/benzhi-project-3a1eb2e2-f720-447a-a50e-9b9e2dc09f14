package domain

import (
	"testing"
	"time"
)

func testCase(t *testing.T) *RetirementCase {
	t.Helper()
	item, err := NewCase(CreateInput{ID: "case-test", Site: "site", Reason: "expired", OwnerID: "owner", PlannedDate: "2030-01-01", Materials: []ControlledMaterial{{MaterialCode: "M-1", DisplayName: "material", HazardClass: "general", DeclaredQuantity: 2, Unit: "L", PackageCondition: "intact", DisposalMethod: "neutralize"}}}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func TestWorkflowRejectsDuplicateCounterAndArchivesReadOnly(t *testing.T) {
	item := testCase(t)
	now := time.Now()
	observation := []CountObservation{{MaterialCode: "M-1", Quantity: 2, PackageCondition: "intact"}}
	if err := item.AddCount("counter-1", observation, "", now); err != nil {
		t.Fatal(err)
	}
	if err := item.AddCount("counter-1", observation, "", now); err == nil {
		t.Fatal("expected duplicate counter error")
	}
	if err := item.AddCount("counter-2", observation, "", now); err != nil {
		t.Fatal(err)
	}
	if _, err := item.AssessRisk(nil, nil, now); err != nil {
		t.Fatal(err)
	}
	if err := item.Review("reviewer", true, "", now); err != nil {
		t.Fatal(err)
	}
	if err := item.RecordDestruction("neutralize", now.Add(-2*time.Minute), now.Add(-time.Minute), []string{"w1", "w2"}, "digest", "", now); err != nil {
		t.Fatal(err)
	}
	if err := item.AddVerification("residue", 1, 0, "reviewer", now); err != nil {
		t.Fatal(err)
	}
	summary, err := item.BuildArchive(5, "head", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := item.ConfirmArchive(*summary, now); err != nil {
		t.Fatal(err)
	}
	if err := item.AddRemediation("cannot change", now); err == nil {
		t.Fatal("archived case must be immutable")
	}
}

func TestRiskBlocksFlammableWithoutControls(t *testing.T) {
	item, err := NewCase(CreateInput{ID: "case-risk", Site: "site", Reason: "expired", OwnerID: "owner", PlannedDate: "2030-01-01", Materials: []ControlledMaterial{{MaterialCode: "M-2", DisplayName: "fuel", HazardClass: "flammable", DeclaredQuantity: 1, Unit: "L", PackageCondition: "intact", DisposalMethod: "burn"}}}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	observation := []CountObservation{{MaterialCode: "M-2", Quantity: 1, PackageCondition: "intact"}}
	_ = item.AddCount("counter-1", observation, "", now)
	_ = item.AddCount("counter-2", observation, "", now)
	risk, err := item.AssessRisk(nil, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if !risk.HasBlocking() || item.Status != StatusCounted {
		t.Fatal("flammable controls should block review")
	}
}
