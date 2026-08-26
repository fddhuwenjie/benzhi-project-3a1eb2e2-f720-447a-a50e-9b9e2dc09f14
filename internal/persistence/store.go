package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

type Store struct {
	directory    string
	guard        sync.Mutex
	locks        map[string]*sync.Mutex
	requestGuard sync.RWMutex
	requests     map[string]application.IdempotencyRecord
}

func Open(directory string) (*Store, error) {
	if strings.TrimSpace(directory) == "" {
		return nil, errors.New("数据目录不能为空")
	}
	if err := os.MkdirAll(filepath.Join(directory, "cases"), 0700); err != nil {
		return nil, err
	}
	store := &Store{directory: directory, locks: map[string]*sync.Mutex{}, requests: map[string]application.IdempotencyRecord{}}
	if err := store.Recover(context.Background()); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) caseLock(caseID string) *sync.Mutex {
	s.guard.Lock()
	defer s.guard.Unlock()
	lock := s.locks[caseID]
	if lock == nil {
		lock = &sync.Mutex{}
		s.locks[caseID] = lock
	}
	return lock
}

func (s *Store) snapshotPath(caseID string) string {
	return filepath.Join(s.directory, "cases", caseID+".json")
}
func (s *Store) eventPath(caseID string) string {
	return filepath.Join(s.directory, "cases", caseID+".events.jsonl")
}

func (s *Store) Load(ctx context.Context, caseID string) (*domain.RetirementCase, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	lock := s.caseLock(caseID)
	lock.Lock()
	defer lock.Unlock()
	value, err := readSnapshot(s.snapshotPath(caseID))
	if errors.Is(err, os.ErrNotExist) {
		return nil, &application.NotFoundError{CaseID: caseID}
	}
	if err != nil {
		return nil, err
	}
	return domain.Clone(value.Case), nil
}

func (s *Store) Events(ctx context.Context, caseID string) ([]audit.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	lock := s.caseLock(caseID)
	lock.Lock()
	defer lock.Unlock()
	return readEvents(s.eventPath(caseID))
}

func (s *Store) FindRequest(ctx context.Context, caseID, requestID string) (*application.IdempotencyRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if record := s.cachedRequest(requestID); record != nil {
		return record, nil
	}
	if caseID != "" {
		lock := s.caseLock(caseID)
		lock.Lock()
		defer lock.Unlock()
		value, err := readSnapshot(s.snapshotPath(caseID))
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		record := findRequest(value.Requests, requestID)
		s.cacheRequest(record)
		return record, nil
	}
	entries, err := os.ReadDir(filepath.Join(s.directory, "cases"))
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		value, readErr := readSnapshot(filepath.Join(s.directory, "cases", entry.Name()))
		if readErr != nil {
			return nil, readErr
		}
		if record := findRequest(value.Requests, requestID); record != nil {
			s.cacheRequest(record)
			return record, nil
		}
	}
	return nil, nil
}

func (s *Store) cachedRequest(requestID string) *application.IdempotencyRecord {
	s.requestGuard.RLock()
	defer s.requestGuard.RUnlock()
	record, ok := s.requests[requestID]
	if !ok {
		return nil
	}
	copy := record
	return &copy
}

func (s *Store) cacheRequest(record *application.IdempotencyRecord) {
	if record == nil {
		return
	}
	s.requestGuard.Lock()
	s.requests[record.RequestID] = *record
	s.requestGuard.Unlock()
}

func findRequest(records []application.IdempotencyRecord, requestID string) *application.IdempotencyRecord {
	for _, record := range records {
		if record.RequestID == requestID {
			copied := record
			return &copied
		}
	}
	return nil
}
