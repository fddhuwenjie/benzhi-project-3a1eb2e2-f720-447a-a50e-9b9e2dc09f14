package persistence

import (
	"context"
	"errors"
	"os"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

func (s *Store) Commit(ctx context.Context, expected int64, item *domain.RetirementCase, event audit.Event, record application.IdempotencyRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	lock := s.caseLock(item.ID)
	lock.Lock()
	defer lock.Unlock()
	path := s.snapshotPath(item.ID)
	current, err := readSnapshot(path)
	if errors.Is(err, os.ErrNotExist) {
		if expected != 0 {
			return &application.RevisionConflictError{Current: 0}
		}
		current = &snapshot{}
	} else if err != nil {
		return err
	}
	if current.Case != nil && current.Case.Revision != expected {
		return &application.RevisionConflictError{Current: current.Case.Revision}
	}
	if prior := findRequest(current.Requests, record.RequestID); prior != nil {
		return nil
	}
	events := []audit.Event{}
	if current.Case != nil {
		events, err = readEvents(s.eventPath(item.ID))
		if err != nil {
			return err
		}
	}
	if event.PreviousHash != audit.Head(events) {
		return &application.RevisionConflictError{Current: expected}
	}
	if event.Revision != item.Revision || item.Revision != expected+1 {
		return errors.New("提交修订号与聚合不一致")
	}
	requests := append(append([]application.IdempotencyRecord(nil), current.Requests...), record)
	if err := writeSnapshotAtomic(path, snapshot{Case: domain.Clone(item), Requests: requests}); err != nil {
		return err
	}
	if err := appendEvent(s.eventPath(item.ID), event); err != nil {
		return err
	}
	return nil
}
