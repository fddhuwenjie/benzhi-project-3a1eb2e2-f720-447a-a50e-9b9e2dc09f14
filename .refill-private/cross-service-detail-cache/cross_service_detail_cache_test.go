package crossservicedetailcache_test

import (
	"context"
	"testing"
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

type isolatedStore struct {
	item   *domain.RetirementCase
	events []audit.Event
}

func (s *isolatedStore) Load(context.Context, string) (*domain.RetirementCase, error) {
	return domain.Clone(s.item), nil
}

func (s *isolatedStore) Events(context.Context, string) ([]audit.Event, error) {
	return append([]audit.Event(nil), s.events...), nil
}

func (s *isolatedStore) FindRequest(context.Context, string, string) (*application.IdempotencyRecord, error) {
	return nil, nil
}

func (s *isolatedStore) Commit(context.Context, int64, *domain.RetirementCase, audit.Event, application.IdempotencyRecord) error {
	return nil
}

func (s *isolatedStore) List(context.Context, application.ListFilter) ([]*domain.RetirementCase, int, error) {
	return nil, 0, nil
}

func TestServicesDoNotShareDetailCacheAcrossStores(t *testing.T) {
	const caseID = "case-shared-name"
	now := time.Unix(1_700_000_000, 0).UTC()
	newStore := func(site, requestID string) *isolatedStore {
		item := &domain.RetirementCase{ID: caseID, Site: site, Status: domain.StatusDraft, Revision: 1, CreatedAt: now, UpdatedAt: now}
		event, err := audit.NewEvent(caseID, "case_created", "admin", requestID, 1, now, map[string]string{"site": site}, "")
		if err != nil {
			t.Fatalf("build event: %v", err)
		}
		return &isolatedStore{item: item, events: []audit.Event{event}}
	}

	first := application.NewService(newStore("一号数据目录", "request-first"))
	firstDetail, err := first.Get(context.Background(), caseID)
	if err != nil {
		t.Fatalf("query first service: %v", err)
	}
	if firstDetail.Case.Site != "一号数据目录" {
		t.Fatalf("unexpected first site: %q", firstDetail.Case.Site)
	}

	second := application.NewService(newStore("二号数据目录", "request-second"))
	secondDetail, err := second.Get(context.Background(), caseID)
	if err != nil {
		t.Fatalf("query second service: %v", err)
	}
	if secondDetail.Case.Site != "二号数据目录" {
		t.Fatalf("second service reused another store's cached detail: got site %q", secondDetail.Case.Site)
	}
}
