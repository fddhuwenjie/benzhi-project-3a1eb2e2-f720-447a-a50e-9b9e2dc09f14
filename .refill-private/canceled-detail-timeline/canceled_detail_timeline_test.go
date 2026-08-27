package canceled_detail_timeline_test

import (
	"context"
	"errors"
	"testing"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

type cancelAfterLoadStore struct {
	item   *domain.RetirementCase
	cancel context.CancelFunc
}

func (s *cancelAfterLoadStore) Load(context.Context, string) (*domain.RetirementCase, error) {
	s.cancel()
	return domain.Clone(s.item), nil
}

func (s *cancelAfterLoadStore) Events(ctx context.Context, _ string) ([]audit.Event, error) {
	return nil, ctx.Err()
}

func (*cancelAfterLoadStore) FindRequest(context.Context, string, string) (*application.IdempotencyRecord, error) {
	panic("unexpected FindRequest call")
}

func (*cancelAfterLoadStore) Commit(context.Context, int64, *domain.RetirementCase, audit.Event, application.IdempotencyRecord) error {
	panic("unexpected Commit call")
}

func (*cancelAfterLoadStore) List(context.Context, application.ListFilter) ([]*domain.RetirementCase, int, error) {
	panic("unexpected List call")
}

func canceledService(item *domain.RetirementCase) (*application.Service, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	return application.NewService(&cancelAfterLoadStore{item: item, cancel: cancel}), ctx
}

func TestCanceledQueriesPreserveContextError(t *testing.T) {
	t.Run("detail", func(t *testing.T) {
		service, ctx := canceledService(&domain.RetirementCase{ID: "case-context", Status: domain.StatusDraft})
		_, err := service.Get(ctx, "case-context")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Get error = %v, want context.Canceled", err)
		}
	})

	t.Run("archive preview", func(t *testing.T) {
		service, ctx := canceledService(&domain.RetirementCase{ID: "case-context", Status: domain.StatusVerified})
		_, err := service.ArchivePreview(ctx, "case-context")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("ArchivePreview error = %v, want context.Canceled", err)
		}
	})

	t.Run("archive export", func(t *testing.T) {
		service, ctx := canceledService(&domain.RetirementCase{ID: "case-context", Status: domain.StatusArchived, Archive: &domain.ArchiveSummary{}})
		_, err := service.ArchiveExport(ctx, "case-context")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("ArchiveExport error = %v, want context.Canceled", err)
		}
	})
}
