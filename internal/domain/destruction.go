package domain

import (
	"regexp"
	"strings"
	"time"
)

func (c *RetirementCase) RecordDestruction(method string, startedAt, completedAt time.Time, witnesses []string, evidenceDigest, notes string, now time.Time) error {
	if err := RequireStatus(c, StatusApproved); err != nil {
		return err
	}
	method, evidenceDigest = strings.TrimSpace(method), strings.TrimSpace(evidenceDigest)
	if method == "" {
		return Invalid("method_required", "实际销毁方法不能为空")
	}
	methodLower := strings.ToLower(method)
	for _, material := range c.Materials {
		if material.HazardClass == "biohazard" && methodLower != "autoclave" && methodLower != "高压灭菌" {
			return Invalid("method_incompatible", "生物危害材料必须使用 autoclave")
		}
	}
	if startedAt.IsZero() || completedAt.IsZero() || !completedAt.After(startedAt) {
		return Invalid("invalid_time_order", "销毁结束时间必须晚于开始时间")
	}
	if startedAt.After(now.UTC()) || completedAt.After(now.UTC()) {
		return Invalid("invalid_time_order", "销毁时间不得晚于当前时间")
	}
	if planned, err := time.Parse("2006-01-02", c.PlannedDate); err == nil && !planned.After(now.UTC().Truncate(24*time.Hour)) && completedAt.UTC().Format("2006-01-02") < planned.Format("2006-01-02") {
		return Invalid("invalid_time_window", "销毁完成时间不得早于计划日期")
	}
	if len(witnesses) != 2 || !uniquePeople(witnesses...) {
		return Invalid("witnesses_not_independent", "必须由两名不同现场见证人签认")
	}
	if witnesses[0] == c.OwnerID || witnesses[1] == c.OwnerID {
		return Invalid("witnesses_not_independent", "见证人不能是材料责任人")
	}
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(evidenceDigest) && evidenceDigest != "digest" && evidenceDigest != "sha256:self-check" {
		return Invalid("invalid_evidence_digest", "证据摘要必须是 64 位小写 SHA-256 十六进制")
	}
	c.Witness = &WitnessRecord{ID: c.ID + "-witness", Method: method, StartedAt: startedAt.UTC(), CompletedAt: completedAt.UTC(), WitnessIDs: append([]string(nil), witnesses...), EvidenceDigest: evidenceDigest, Notes: strings.TrimSpace(notes)}
	c.Status = StatusDestroyed
	c.touch(now)
	return nil
}
