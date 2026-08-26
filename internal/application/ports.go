package application

import (
	"context"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

type IdempotencyRecord struct {
	RequestID string `json:"request_id"`
	CaseID    string `json:"case_id"`
	Operation string `json:"operation"`
	Revision  int64  `json:"revision"`
}

type ListFilter struct {
	Status      domain.Status
	Site        string
	PlannedFrom string
	PlannedTo   string
	HazardClass string
	Limit       int
	Offset      int
}

type Store interface {
	Load(context.Context, string) (*domain.RetirementCase, error)
	Events(context.Context, string) ([]audit.Event, error)
	FindRequest(context.Context, string, string) (*IdempotencyRecord, error)
	Commit(context.Context, int64, *domain.RetirementCase, audit.Event, IdempotencyRecord) error
	List(context.Context, ListFilter) ([]*domain.RetirementCase, int, error)
}

type StatsProvider interface {
	Stats(context.Context, ListFilter) (ListStats, error)
}

type NotFoundError struct{ CaseID string }

func (e *NotFoundError) Error() string { return "case not found: " + e.CaseID }

type RevisionConflictError struct{ Current int64 }

func (e *RevisionConflictError) Error() string { return "revision conflict" }
