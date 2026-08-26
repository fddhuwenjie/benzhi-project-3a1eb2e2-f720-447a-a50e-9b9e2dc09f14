package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

func (c *RetirementCase) BuildArchive(eventCount int, chainHead string, now time.Time) (*ArchiveSummary, error) {
	if c.Status != StatusVerified && c.Status != StatusArchived {
		return nil, Invalid("invalid_status", "仅已验证批次可以生成归档摘要")
	}
	summary := &ArchiveSummary{MaterialCount: len(c.Materials), EventCount: eventCount, ChainHead: chainHead, GeneratedAt: now.UTC(), Revision: c.Revision}
	for _, material := range c.Materials {
		summary.MaterialCodes = append(summary.MaterialCodes, material.MaterialCode)
	}
	if c.RiskBaseline != nil {
		for _, item := range c.RiskBaseline.Findings {
			summary.RiskFindingCodes = append(summary.RiskFindingCodes, item.Code)
		}
	}
	for i := len(c.Reviews) - 1; i >= 0; i-- {
		if c.Reviews[i].Approved {
			summary.ApprovedBy = c.Reviews[i].ReviewerID
			break
		}
	}
	if c.Witness != nil {
		summary.Witnesses = append([]string(nil), c.Witness.WitnessIDs...)
	}
	for _, record := range c.Verifications {
		summary.VerificationChecks = append(summary.VerificationChecks, record.CheckName+":"+record.Result)
	}
	digestView := *summary
	digestView.GeneratedAt = time.Time{}
	digestView.PreviewDigest = ""
	digestView.SnapshotDigest = ""
	data, _ := json.Marshal(digestView)
	sum := sha256.Sum256(data)
	summary.PreviewDigest = hex.EncodeToString(sum[:])
	summary.SnapshotDigest = summary.PreviewDigest
	return summary, nil
}

func (c *RetirementCase) ConfirmArchive(summary ArchiveSummary, now time.Time) error {
	if err := RequireStatus(c, StatusVerified); err != nil {
		return err
	}
	c.Archive = &summary
	archived := now.UTC()
	c.ArchivedAt = &archived
	c.Status = StatusArchived
	c.touch(now)
	return nil
}
