package listcachealias

import (
	"context"
	"testing"
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/persistence"
)

func TestRepeatedListDoesNotReuseMutableCases(t *testing.T) {
	store, err := persistence.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2030, 1, 1, 12, 0, 0, 0, time.UTC)
	item, err := domain.NewCase(domain.CreateInput{
		ID:          "case-cache-alias",
		Site:        "受控实验室 A",
		Reason:      "到期退役",
		OwnerID:     "owner-1",
		PlannedDate: "2030-01-02",
		Materials: []domain.ControlledMaterial{{
			MaterialCode: "MAT-001", DisplayName: "标准品", HazardClass: "general",
			DeclaredQuantity: 1, Unit: "瓶", PackageCondition: "intact", DisposalMethod: "neutralize",
		}},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	event, err := audit.NewEvent(item.ID, "case_created", "admin-1", "request-cache-1", item.Revision, now, map[string]string{"site": item.Site}, "")
	if err != nil {
		t.Fatal(err)
	}
	record := application.IdempotencyRecord{RequestID: "request-cache-1", CaseID: item.ID, Operation: "create", Revision: item.Revision}
	if err := store.Commit(context.Background(), 0, item, event, record); err != nil {
		t.Fatal(err)
	}

	filter := application.ListFilter{Limit: 50}
	first, _, err := store.List(context.Background(), filter)
	if err != nil || len(first) != 1 {
		t.Fatalf("first list failed: items=%d err=%v", len(first), err)
	}
	first[0].Site = "被调用方篡改的场所"
	first[0].Materials[0].DisplayName = "被调用方篡改的材料"

	second, _, err := store.List(context.Background(), filter)
	if err != nil || len(second) != 1 {
		t.Fatalf("second list failed: items=%d err=%v", len(second), err)
	}
	if second[0].Site != "受控实验室 A" || second[0].Materials[0].DisplayName != "标准品" {
		t.Fatalf("repeated list exposed caller-mutated cached case: site=%q material=%q", second[0].Site, second[0].Materials[0].DisplayName)
	}
}
