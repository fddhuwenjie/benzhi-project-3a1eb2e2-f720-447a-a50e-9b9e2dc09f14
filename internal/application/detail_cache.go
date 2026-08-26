package application

import (
	"sync"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

type detailCacheEntry struct {
	caseValue *domain.RetirementCase
	timeline  []audit.Projection
	chainOK   bool
}

var sharedDetailCache = struct {
	sync.RWMutex
	entries map[string]detailCacheEntry
}{entries: map[string]detailCacheEntry{}}

func cachedDetail(caseID string, revision int64) (*Detail, bool) {
	sharedDetailCache.RLock()
	entry, ok := sharedDetailCache.entries[caseID]
	sharedDetailCache.RUnlock()
	if !ok || entry.caseValue.Revision != revision {
		return nil, false
	}
	return entry.detail(), true
}

func cacheDetail(caseID string, item *domain.RetirementCase, timeline []audit.Projection, chainOK bool) {
	entry := detailCacheEntry{caseValue: domain.Clone(item), timeline: append([]audit.Projection(nil), timeline...), chainOK: chainOK}
	sharedDetailCache.Lock()
	sharedDetailCache.entries[caseID] = entry
	sharedDetailCache.Unlock()
}

func invalidateDetail(caseID string) {
	sharedDetailCache.Lock()
	delete(sharedDetailCache.entries, caseID)
	sharedDetailCache.Unlock()
}

func (e detailCacheEntry) detail() *Detail {
	return &Detail{Case: domain.Clone(e.caseValue), Timeline: append([]audit.Projection(nil), e.timeline...), ChainOK: e.chainOK}
}
