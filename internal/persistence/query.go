package persistence

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

func (s *Store) List(ctx context.Context, filter application.ListFilter) ([]*domain.RetirementCase, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	s.queryMu.RLock()
	cached, ok := s.queries[filter]
	s.queryMu.RUnlock()
	if ok {
		results := make([]*domain.RetirementCase, len(cached.items))
		for i, item := range cached.items {
			results[i] = domain.Clone(item)
		}
		return results, cached.total, nil
	}
	entries, err := os.ReadDir(filepath.Join(s.directory, "cases"))
	if err != nil {
		return nil, 0, err
	}
	items := make([]*domain.RetirementCase, 0)
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		value, err := readSnapshot(filepath.Join(s.directory, "cases", entry.Name()))
		if err != nil {
			return nil, 0, err
		}
		if filter.Status != "" && value.Case.Status != filter.Status {
			continue
		}
		if filter.Site != "" && value.Case.Site != filter.Site {
			continue
		}
		if filter.PlannedFrom != "" && value.Case.PlannedDate < filter.PlannedFrom {
			continue
		}
		if filter.PlannedTo != "" && value.Case.PlannedDate > filter.PlannedTo {
			continue
		}
		if filter.HazardClass != "" {
			found := false
			for _, m := range value.Case.Materials {
				if m.HazardClass == filter.HazardClass {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		items = append(items, domain.Clone(value.Case))
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	total := len(items)
	start := filter.Offset
	if start > total {
		start = total
	}
	end := start + filter.Limit
	if end > total {
		end = total
	}
	result := items[start:end]
	cloned := make([]*domain.RetirementCase, len(result))
	for i, item := range result {
		cloned[i] = domain.Clone(item)
	}
	s.queryMu.Lock()
	s.queries[filter] = cachedList{items: cloned, total: total}
	s.queryMu.Unlock()
	return result, total, nil
}

func (s *Store) Stats(ctx context.Context, filter application.ListFilter) (application.ListStats, error) {
	entries, err := os.ReadDir(filepath.Join(s.directory, "cases"))
	if err != nil {
		return application.ListStats{}, err
	}
	stats := application.ListStats{ByStatus: map[string]int{}, ByHazardClass: map[string]int{}}
	today := time.Now().UTC().Format("2006-01-02")
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		value, readErr := readSnapshot(filepath.Join(s.directory, "cases", entry.Name()))
		if readErr != nil {
			return stats, readErr
		}
		c := value.Case
		if filter.Status != "" && c.Status != filter.Status || filter.Site != "" && c.Site != filter.Site || filter.PlannedFrom != "" && c.PlannedDate < filter.PlannedFrom || filter.PlannedTo != "" && c.PlannedDate > filter.PlannedTo {
			continue
		}
		if filter.HazardClass != "" {
			found := false
			for _, m := range c.Materials {
				if m.HazardClass == filter.HazardClass {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		stats.ByStatus[string(c.Status)]++
		if c.Status != domain.StatusArchived && c.PlannedDate < today {
			stats.Overdue++
		}
		for _, m := range c.Materials {
			stats.ByHazardClass[m.HazardClass]++
		}
	}
	return stats, nil
}
