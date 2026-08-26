package domain

import (
	"math"
	"sort"
	"strings"
	"time"
)

func (c *RetirementCase) AddVerification(checkName string, threshold, measured float64, reviewerID string, now time.Time) error {
	if err := RequireStatus(c, StatusDestroyed, StatusRemediation); err != nil {
		return err
	}
	checkName, reviewerID = strings.TrimSpace(checkName), strings.TrimSpace(reviewerID)
	if checkName == "" {
		return Invalid("check_name_required", "验证项目不能为空")
	}
	if math.IsNaN(threshold) || math.IsInf(threshold, 0) || math.IsNaN(measured) || math.IsInf(measured, 0) || threshold < 0 || measured < 0 {
		return Invalid("invalid_measurement", "阈值与实测结果不能小于 0")
	}
	if reviewerID == "" || reviewerID == c.OwnerID {
		return Invalid("verifier_not_independent", "复核人不能为空且不能是材料责任人")
	}
	if c.Status == StatusRemediation && len(c.RemediationNotes) == 0 {
		return Invalid("remediation_required", "复验前必须追加补救记录")
	}
	if c.Status == StatusRemediation {
		failed := ""
		for i := len(c.Verifications) - 1; i >= 0; i-- {
			if c.Verifications[i].Result == "failed" {
				failed = c.Verifications[i].CheckName
				break
			}
		}
		if failed != checkName {
			return Invalid("remediation_scope", "复验只能针对最近一次失败项目")
		}
	}
	required := c.RequiredVerificationChecks()
	for _, prior := range c.Verifications {
		if prior.CheckName == checkName && prior.Result == "passed" && c.Status != StatusRemediation {
			return Invalid("duplicate_verification", "同一轮不能重复提交已通过项目")
		}
	}
	result := "passed"
	if measured > threshold {
		result = "failed"
	}
	record := VerificationRecord{ID: c.ID + "-verification-" + itoa(len(c.Verifications)+1), CheckName: checkName, Threshold: threshold, MeasuredValue: measured, Result: result, ReviewerID: reviewerID, VerifiedAt: now.UTC()}
	if c.Status == StatusRemediation {
		record.RemediationNote = c.RemediationNotes[len(c.RemediationNotes)-1]
	}
	c.Verifications = append(c.Verifications, record)
	if result == "failed" {
		c.Status = StatusRemediation
	} else {
		passed := map[string]bool{}
		for _, item := range c.Verifications {
			if item.Result == "passed" {
				passed[item.CheckName] = true
			}
		}
		passed[checkName] = true
		complete := true
		for _, name := range required {
			if !passed[name] {
				complete = false
				break
			}
		}
		if complete {
			c.Status = StatusVerified
		} else {
			c.Status = StatusDestroyed
		}
	}
	c.touch(now)
	return nil
}

func (c *RetirementCase) RequiredVerificationChecks() []string {
	set := map[string]bool{}
	hasSpecial := false
	for _, m := range c.Materials {
		switch m.HazardClass {
		case "toxic":
			hasSpecial = true
			set["toxic_residue"] = true
		case "corrosive":
			hasSpecial = true
			set["corrosive_residue"] = true
		case "biohazard":
			hasSpecial = true
			set["biological_inactivation"] = true
		}
	}
	if !hasSpecial {
		return nil
	}
	result := make([]string, 0, len(set))
	for name := range set {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func (c *RetirementCase) AddRemediation(note string, now time.Time) error {
	if err := RequireStatus(c, StatusRemediation); err != nil {
		return err
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return Invalid("remediation_note_required", "补救记录不能为空")
	}
	c.RemediationNotes = append(c.RemediationNotes, note)
	c.touch(now)
	return nil
}
