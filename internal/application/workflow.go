package application

import (
	"context"
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/audit"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

func (s *Service) Correct(ctx context.Context, command CorrectCommand) (*MutationResult, error) {
	return s.mutate(ctx, command.CaseID, "correct", "case_corrected", command.CommandMeta, []string{"safety_admin"}, command, func(c *domain.RetirementCase, now time.Time) error {
		if command.OpinionID != "" || command.Patch != nil {
			return c.CorrectTargeted(command.OpinionID, command.Patch, now)
		}
		return c.Correct(command.Site, command.Reason, command.PlannedDate, command.Materials, now)
	})
}

func (s *Service) Count(ctx context.Context, command CountCommand) (*MutationResult, error) {
	return s.mutate(ctx, command.CaseID, "count", "count_confirmed", command.CommandMeta, []string{"material_owner", "witness"}, command, func(c *domain.RetirementCase, now time.Time) error {
		return c.AddCountDetailed(command.CounterID, command.Observations, command.DifferenceReason, command.DifferenceExplanations, now)
	})
}

func (s *Service) AssessRisk(ctx context.Context, command RiskCommand) (*MutationResult, error) {
	return s.mutate(ctx, command.CaseID, "risk", "risk_assessed", command.CommandMeta, []string{"safety_admin"}, command, func(c *domain.RetirementCase, now time.Time) error {
		_, err := c.AssessRiskWithConfirmations(command.SiteConditions, command.ProtectiveMeasures, command.WarningConfirmations, now)
		return err
	})
}

func (s *Service) Review(ctx context.Context, command ReviewCommand) (*MutationResult, error) {
	if command.ReviewerID == "" {
		command.ReviewerID = command.ActorID
	}
	event := "review_returned"
	if command.Approved {
		event = "review_approved"
	}
	return s.mutate(ctx, command.CaseID, "review", event, command.CommandMeta, []string{"compliance_reviewer"}, command, func(c *domain.RetirementCase, now time.Time) error {
		return c.Review(command.ReviewerID, command.Approved, command.Reason, command.AllowedFields, now)
	})
}

func (s *Service) RecordDestruction(ctx context.Context, command DestructionCommand) (*MutationResult, error) {
	return s.mutate(ctx, command.CaseID, "destruction", "destruction_witnessed", command.CommandMeta, []string{"witness"}, command, func(c *domain.RetirementCase, now time.Time) error {
		return c.RecordDestruction(command.Method, command.StartedAt, command.CompletedAt, command.WitnessIDs, command.EvidenceDigest, command.Notes, now)
	})
}

func (s *Service) Verify(ctx context.Context, command VerificationCommand) (*MutationResult, error) {
	if command.ReviewerID == "" {
		command.ReviewerID = command.ActorID
	}
	return s.mutate(ctx, command.CaseID, "verification", "residue_verified", command.CommandMeta, []string{"compliance_reviewer"}, command, func(c *domain.RetirementCase, now time.Time) error {
		return c.AddVerification(command.CheckName, command.Threshold, command.MeasuredValue, command.ReviewerID, now)
	})
}

func (s *Service) Remediate(ctx context.Context, command RemediationCommand) (*MutationResult, error) {
	return s.mutate(ctx, command.CaseID, "remediation", "remediation_added", command.CommandMeta, []string{"safety_admin"}, command, func(c *domain.RetirementCase, now time.Time) error {
		return c.AddRemediation(command.Note, now)
	})
}

func (s *Service) Archive(ctx context.Context, command ArchiveCommand) (*MutationResult, error) {
	return s.mutate(ctx, command.CaseID, "archive", "case_archived", command.CommandMeta, []string{"compliance_reviewer"}, command, func(c *domain.RetirementCase, now time.Time) error {
		events, err := s.store.Events(ctx, c.ID)
		if err != nil {
			return err
		}
		summary, err := c.BuildArchive(len(events), audit.Head(events), now)
		if err != nil {
			return err
		}
		if command.PreviewDigest == "" {
			return domain.Invalid("preview_required", "必须提交归档预览摘要")
		}
		if command.PreviewDigest != summary.PreviewDigest {
			return domain.Invalid("preview_stale", "归档预览已变化，请刷新后重试")
		}
		if command.ExpectedChainHead != "" && command.ExpectedChainHead != audit.Head(events) {
			return domain.Invalid("preview_stale", "审计链已变化，请刷新后重试")
		}
		return c.ConfirmArchive(*summary, now)
	})
}
