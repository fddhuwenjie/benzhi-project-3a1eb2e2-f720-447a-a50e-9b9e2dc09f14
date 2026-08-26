package application

import (
	"context"
	"strings"
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

func (s *Service) Get(ctx context.Context, caseID string) (*Detail, error) {
	item, err := s.store.Load(ctx, caseID)
	if err != nil {
		return nil, s.mapStoreError(err)
	}
	events, err := s.store.Events(ctx, caseID)
	if err != nil {
		return nil, err
	}
	chainOK := audit.Verify(events) == nil
	return &Detail{Case: item, Timeline: audit.Project(events), ChainOK: chainOK}, nil
}

func (s *Service) List(ctx context.Context, filter ListFilter) (*ListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 || filter.Offset < 0 {
		return nil, &Error{Code: "validation_failed", Message: "分页参数无效", Fields: map[string]string{"limit": "limit 必须为 1-100", "offset": "offset 不能为负数"}}
	}
	if filter.Status != "" {
		valid := map[domain.Status]bool{domain.StatusDraft: true, domain.StatusCounted: true, domain.StatusPendingReview: true, domain.StatusApproved: true, domain.StatusDestroyed: true, domain.StatusRemediation: true, domain.StatusVerified: true, domain.StatusArchived: true}
		if !valid[filter.Status] {
			return nil, &Error{Code: "validation_failed", Message: "状态无效", Fields: map[string]string{"status": "未知状态"}}
		}
	}
	if filter.PlannedFrom != "" {
		if _, err := time.Parse("2006-01-02", filter.PlannedFrom); err != nil {
			return nil, &Error{Code: "validation_failed", Message: "日期范围无效", Fields: map[string]string{"planned_date_from": "日期格式必须为 YYYY-MM-DD"}}
		}
	}
	if filter.PlannedTo != "" {
		if _, err := time.Parse("2006-01-02", filter.PlannedTo); err != nil {
			return nil, &Error{Code: "validation_failed", Message: "日期范围无效", Fields: map[string]string{"planned_date_to": "日期格式必须为 YYYY-MM-DD"}}
		}
	}
	if filter.PlannedFrom != "" && filter.PlannedTo != "" && filter.PlannedFrom > filter.PlannedTo {
		return nil, &Error{Code: "validation_failed", Message: "日期范围无效", Fields: map[string]string{"planned_date": "起始日期不能晚于结束日期"}}
	}
	if strings.TrimSpace(filter.Site) == "" {
		filter.Site = ""
	}
	items, total, err := s.store.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats := ListStats{ByStatus: map[string]int{}, ByHazardClass: map[string]int{}}
	if provider, ok := s.store.(StatsProvider); ok {
		if computed, statsErr := provider.Stats(ctx, filter); statsErr == nil {
			stats = computed
		}
	}
	if len(stats.ByStatus) == 0 && len(stats.ByHazardClass) == 0 {
		for _, item := range items {
			stats.ByStatus[string(item.Status)]++
			for _, material := range item.Materials {
				stats.ByHazardClass[material.HazardClass]++
			}
			if item.Status != domain.StatusArchived && item.PlannedDate < s.now().UTC().Format("2006-01-02") {
				stats.Overdue++
			}
		}
	}
	return &ListResult{Items: items, Total: total, Stats: stats, Limit: filter.Limit, Offset: filter.Offset}, nil
}

func (s *Service) ArchivePreview(ctx context.Context, caseID string) (*domain.ArchiveSummary, error) {
	item, err := s.store.Load(ctx, caseID)
	if err != nil {
		return nil, s.mapStoreError(err)
	}
	events, err := s.store.Events(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if err := audit.Verify(events); err != nil {
		return nil, err
	}
	return item.BuildArchive(len(events), audit.Head(events), s.now())
}

func (s *Service) ArchiveExport(ctx context.Context, caseID string) (map[string]any, error) {
	item, err := s.store.Load(ctx, caseID)
	if err != nil {
		return nil, s.mapStoreError(err)
	}
	if item.Status != domain.StatusArchived || item.Archive == nil {
		return nil, &Error{Code: "invalid_status", Message: "仅已归档批次可以导出"}
	}
	events, err := s.store.Events(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if err := audit.Verify(events); err != nil {
		return nil, &Error{Code: "audit_chain_corrupt", Message: "审计链校验失败"}
	}
	return map[string]any{"case": item, "archive": item.Archive, "chain_head": audit.Head(events), "snapshot_digest": item.Archive.SnapshotDigest}, nil
}
