package domain

import (
	"strings"
	"time"
)

func (c *RetirementCase) Review(reviewerID string, approved bool, reason string, args ...interface{}) error {
	var allowedFields []string
	var now time.Time
	for _, arg := range args {
		switch v := arg.(type) {
		case []string:
			allowedFields = v
		case time.Time:
			now = v
		}
	}
	if err := RequireStatus(c, StatusPendingReview); err != nil {
		return err
	}
	reviewerID, reason = strings.TrimSpace(reviewerID), strings.TrimSpace(reason)
	if reviewerID == "" || reviewerID == c.OwnerID {
		return Invalid("reviewer_not_independent", "复核人员不能为空且不能是材料责任人")
	}
	if !approved && reason == "" {
		return Invalid("review_reason_required", "退回时必须填写理由")
	}
	if !c.RiskCurrent() {
		return Invalid("risk_expired", "风险评估已过期，请重新执行规则")
	}
	decision := ReviewDecision{ID: c.ID + "-review-" + itoa(len(c.Reviews)+1), ReviewerID: reviewerID, Approved: approved, Reason: reason, AllowedFields: append([]string(nil), allowedFields...), DecidedAt: now.UTC()}
	c.Reviews = append(c.Reviews, decision)
	if approved {
		if c.Risk == nil || c.Risk.HasBlocking() {
			return Invalid("risk_not_clear", "风险阻断项未清除，不能批准")
		}
		baseline := *c.Risk
		baseline.Findings = append([]RiskFinding(nil), c.Risk.Findings...)
		baseline.ProtectiveMeasures = append([]string(nil), c.Risk.ProtectiveMeasures...)
		baseline.SiteConditions = append([]string(nil), c.Risk.SiteConditions...)
		c.RiskBaseline = &baseline
		c.Status = StatusApproved
	} else {
		c.Status = StatusCounted
		c.Risk = nil
	}
	c.touch(now)
	return nil
}

func (c *RetirementCase) CorrectTargeted(opinionID string, patch map[string]any, now time.Time) error {
	if err := RequireStatus(c, StatusCounted); err != nil {
		return err
	}
	if len(c.Reviews) == 0 || c.Reviews[len(c.Reviews)-1].Approved {
		return Invalid("opinion_required", "必须引用最近一次退回意见")
	}
	latest := c.Reviews[len(c.Reviews)-1]
	if opinionID == "" || opinionID != latest.ID {
		return Invalid("opinion_mismatch", "退回意见标识已过期")
	}
	allowed := map[string]bool{}
	for _, field := range latest.AllowedFields {
		allowed[field] = true
	}
	for field := range patch {
		if !allowed[field] {
			return Invalid("field_not_allowed", "该字段不在退回意见允许范围内："+field)
		}
	}
	// 先校验全部补正字段并暂存到局部变量，确认无误后再写入聚合，避免校验失败时留下半成品状态。
	site, reason, plannedDate := c.Site, c.Reason, c.PlannedDate
	for field, value := range patch {
		switch field {
		case "site":
			v, ok := value.(string)
			if !ok {
				return Invalid("validation_failed", "site 必须为文本")
			}
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				return Invalid("validation_failed", "场所不能为空")
			}
			site = trimmed
		case "reason":
			v, ok := value.(string)
			if !ok {
				return Invalid("validation_failed", "reason 必须为文本")
			}
			reason = strings.TrimSpace(v)
		case "planned_date":
			v, ok := value.(string)
			if !ok {
				return Invalid("validation_failed", "planned_date 必须为日期")
			}
			plannedDate = v
		default:
			return Invalid("field_not_allowed", "不支持补正字段："+field)
		}
	}
	c.Site, c.Reason, c.PlannedDate = site, reason, plannedDate
	c.Risk = nil
	c.Status = StatusCounted
	c.touch(now)
	return nil
}
