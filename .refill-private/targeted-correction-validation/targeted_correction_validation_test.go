package targetedcorrectionvalidation

import (
	"testing"
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

func TestTargetedCorrectionRejectsInvalidAggregate(t *testing.T) {
	item := &domain.RetirementCase{
		ID:       "case-1",
		Site:     "原场所",
		Reason:   "退役",
		OwnerID:  "owner",
		PlannedDate: "2030-01-01",
		Status:   domain.StatusCounted,
		Reviews: []domain.ReviewDecision{{
			ID: "op-1", Approved: false, AllowedFields: []string{"site"},
		}},
	}
	if err := item.CorrectTargeted("op-1", map[string]any{"site": "   "}, time.Unix(10, 0)); err == nil {
		t.Fatal("targeted correction accepted an invalid empty site")
	}
}
