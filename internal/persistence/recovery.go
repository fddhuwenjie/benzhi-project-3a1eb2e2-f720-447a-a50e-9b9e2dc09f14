package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (s *Store) Recover(ctx context.Context) error {
	entries, err := os.ReadDir(filepath.Join(s.directory, "cases"))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		value, err := readSnapshot(filepath.Join(s.directory, "cases", entry.Name()))
		if err != nil {
			return err
		}
		events, err := readEvents(s.eventPath(value.Case.ID))
		if err != nil {
			return fmt.Errorf("恢复批次 %s 失败: %w", value.Case.ID, err)
		}
		if len(events) == 0 {
			return fmt.Errorf("恢复批次 %s 失败: 缺少审计事件", value.Case.ID)
		}
		if events[len(events)-1].Revision != value.Case.Revision {
			return fmt.Errorf("恢复批次 %s 失败: 快照修订号与事件链不一致", value.Case.ID)
		}
	}
	return nil
}
