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

// errUnsafeCaseID is returned for case identifiers that could escape the cases
// directory (for example by containing ".." or path separators). Callers map
// this to their not-found semantics so no external file is ever accessed.
var errUnsafeCaseID = errors.New("invalid case identifier")

// safeCaseID reports whether caseID is a plain filename component that cannot
// traverse out of the cases directory. It rejects empty values, path
// separators (both "/" and the OS separator), and any component equal to or
// containing "." or ".." before any file is accessed.
func safeCaseID(caseID string) error {
	if caseID == "" {
		return errUnsafeCaseID
	}
	if strings.ContainsAny(caseID, `/\`) {
		return errUnsafeCaseID
	}
	for _, part := range strings.Split(caseID, string(filepath.Separator)) {
		if part == "." || part == ".." {
			return errUnsafeCaseID
		}
	}
	if caseID == "." || caseID == ".." {
		return errUnsafeCaseID
	}
	return nil
}

type Store struct {
	directory string
	guard     sync.Mutex
	locks     map[string]*sync.Mutex
}

func Open(directory string) (*Store, error) {
	if strings.TrimSpace(directory) == "" {
		return nil, errors.New("数据目录不能为空")
	}
	if err := os.MkdirAll(filepath.Join(directory, "cases"), 0700); err != nil {
		return nil, err
	}
	store := &Store{directory: directory, locks: map[string]*sync.Mutex{}}
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
	if err := safeCaseID(caseID); err != nil {
		return nil, &application.NotFoundError{CaseID: caseID}
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
	if err := safeCaseID(caseID); err != nil {
		return []audit.Event{}, nil
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
	if caseID != "" {
		if err := safeCaseID(caseID); err != nil {
			return nil, nil
		}
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
		return findRequest(value.Requests, requestID), nil
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
			return record, nil
		}
	}
	return nil, nil
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
