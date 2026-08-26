package persistence

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

type snapshot struct {
	Case     *domain.RetirementCase          `json:"case"`
	Requests []application.IdempotencyRecord `json:"requests"`
}

func readSnapshot(path string) (*snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value snapshot
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("读取快照 %s 失败: %w", filepath.Base(path), err)
	}
	if value.Case == nil || value.Case.ID == "" {
		return nil, fmt.Errorf("快照 %s 缺少批次数据", filepath.Base(path))
	}
	return &value, nil
}

func writeSnapshotAtomic(path string, value snapshot) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".snapshot-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	clean := func() { _ = temporary.Close(); _ = os.Remove(temporaryPath) }
	if _, err := temporary.Write(data); err != nil {
		clean()
		return err
	}
	if err := temporary.Sync(); err != nil {
		clean()
		return err
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	dir, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func readEvents(path string) ([]audit.Event, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return []audit.Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var events []audit.Event
	lineNumber := 0
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			lineNumber++
			if line[len(line)-1] != '\n' {
				return nil, fmt.Errorf("审计日志第 %d 行被截断", lineNumber)
			}
			var event audit.Event
			if err := json.Unmarshal(line, &event); err != nil {
				return nil, fmt.Errorf("审计日志第 %d 行无效: %w", lineNumber, err)
			}
			events = append(events, event)
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}
	if err := audit.Verify(events); err != nil {
		return nil, err
	}
	return events, nil
}

func appendEvent(path string, event audit.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	return file.Sync()
}
