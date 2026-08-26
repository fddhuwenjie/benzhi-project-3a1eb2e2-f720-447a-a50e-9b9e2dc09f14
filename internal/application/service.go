package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service { return &Service{store: store, now: time.Now} }

func (s *Service) Create(ctx context.Context, command CreateCommand) (*MutationResult, error) {
	if err := validateMeta(command.CommandMeta, "safety_admin"); err != nil {
		return nil, err
	}
	caseID := newID("case")
	if existing, err := s.store.FindRequest(ctx, "", command.RequestID); err == nil && existing != nil {
		if existing.Operation != "create" {
			return nil, &Error{Code: "request_id_reused", Message: "request_id 已用于其他操作"}
		}
		return s.replayed(ctx, existing)
	}
	now := s.now().UTC()
	item, err := domain.NewCase(domain.CreateInput{ID: caseID, Site: command.Site, Reason: command.Reason, OwnerID: command.OwnerID, PlannedDate: command.PlannedDate, Materials: command.Materials}, now)
	if err != nil {
		return nil, MapError(err)
	}
	event, err := audit.NewEvent(item.ID, "case_created", command.ActorID, command.RequestID, item.Revision, now, command, "")
	if err != nil {
		return nil, err
	}
	record := IdempotencyRecord{RequestID: command.RequestID, CaseID: item.ID, Operation: "create", Revision: item.Revision}
	if err := s.store.Commit(ctx, 0, item, event, record); err != nil {
		return nil, s.mapStoreError(err)
	}
	return &MutationResult{Case: domain.Clone(item)}, nil
}

func (s *Service) mutate(ctx context.Context, caseID, operation, eventType string, meta CommandMeta, roles []string, payload any, mutate func(*domain.RetirementCase, time.Time) error) (*MutationResult, error) {
	if err := validateMeta(meta, roles...); err != nil {
		return nil, err
	}
	if strings.TrimSpace(caseID) == "" {
		return nil, &Error{Code: "validation_failed", Message: "批次标识不能为空"}
	}
	if record, err := s.store.FindRequest(ctx, caseID, meta.RequestID); err == nil && record != nil {
		if record.Operation != operation {
			return nil, &Error{Code: "request_id_reused", Message: "request_id 已用于其他操作"}
		}
		return s.replayed(ctx, record)
	}
	item, err := s.store.Load(ctx, caseID)
	if err != nil {
		return nil, s.mapStoreError(err)
	}
	if meta.ExpectedRevision != item.Revision {
		return nil, Conflict(item.Revision)
	}
	events, err := s.store.Events(ctx, caseID)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	if err := mutate(item, now); err != nil {
		return nil, MapError(err)
	}
	event, err := audit.NewEvent(caseID, eventType, meta.ActorID, meta.RequestID, item.Revision, now, payload, audit.Head(events))
	if err != nil {
		return nil, err
	}
	record := IdempotencyRecord{RequestID: meta.RequestID, CaseID: caseID, Operation: operation, Revision: item.Revision}
	if err := s.store.Commit(ctx, meta.ExpectedRevision, item, event, record); err != nil {
		return nil, s.mapStoreError(err)
	}
	invalidateDetail(caseID)
	return &MutationResult{Case: domain.Clone(item)}, nil
}

func (s *Service) replayed(ctx context.Context, record *IdempotencyRecord) (*MutationResult, error) {
	item, err := s.store.Load(ctx, record.CaseID)
	if err != nil {
		return nil, s.mapStoreError(err)
	}
	return &MutationResult{Case: item, Replayed: true}, nil
}

func (s *Service) mapStoreError(err error) error {
	var missing *NotFoundError
	if errors.As(err, &missing) {
		return NotFound()
	}
	var conflict *RevisionConflictError
	if errors.As(err, &conflict) {
		return Conflict(conflict.Current)
	}
	return err
}

func validateMeta(meta CommandMeta, roles ...string) error {
	fields := map[string]string{}
	if strings.TrimSpace(meta.ActorID) == "" {
		fields["actor_id"] = "操作者不能为空"
	}
	if strings.TrimSpace(meta.RequestID) == "" || len(meta.RequestID) > 128 {
		fields["request_id"] = "request_id 不能为空且不能超过 128 字符"
	}
	allowed := false
	for _, role := range roles {
		if meta.Role == role {
			allowed = true
		}
	}
	if !allowed {
		return &Error{Code: "forbidden", Message: "当前角色无权执行该操作"}
	}
	if len(fields) > 0 {
		return &Error{Code: "validation_failed", Message: "请求元数据无效", Fields: fields}
	}
	return nil
}

func newID(prefix string) string {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err != nil {
		return prefix + "-" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return prefix + "-" + hex.EncodeToString(buffer)
}
