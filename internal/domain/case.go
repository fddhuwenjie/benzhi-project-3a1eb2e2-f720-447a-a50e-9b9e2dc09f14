package domain

import (
	"strings"
	"time"
)

func NewCase(in CreateInput, now time.Time) (*RetirementCase, error) {
	if err := ValidateCreate(in); err != nil {
		return nil, err
	}
	c := &RetirementCase{
		ID: strings.TrimSpace(in.ID), Site: strings.TrimSpace(in.Site), Reason: strings.TrimSpace(in.Reason),
		OwnerID: strings.TrimSpace(in.OwnerID), PlannedDate: in.PlannedDate, Status: StatusDraft,
		Revision: 1, CreatedAt: now.UTC(), UpdatedAt: now.UTC(), Materials: append([]ControlledMaterial(nil), in.Materials...),
		Counts: []CountConfirmation{}, Reviews: []ReviewDecision{}, Verifications: []VerificationRecord{}, RemediationNotes: []string{},
	}
	for i := range c.Materials {
		c.Materials[i].ID = c.ID + "-material-" + itoa(i+1)
		c.Materials[i].MaterialCode = strings.TrimSpace(c.Materials[i].MaterialCode)
	}
	return c, nil
}

func (c *RetirementCase) touch(now time.Time) { c.Revision++; c.UpdatedAt = now.UTC() }

func (c *RetirementCase) Correct(site, reason, plannedDate string, materials []ControlledMaterial, now time.Time) error {
	if err := RequireStatus(c, StatusDraft, StatusCounted); err != nil {
		return err
	}
	input := CreateInput{ID: c.ID, Site: site, Reason: reason, OwnerID: c.OwnerID, PlannedDate: plannedDate, Materials: materials}
	if err := ValidateCreate(input); err != nil {
		return err
	}
	c.Site, c.Reason, c.PlannedDate = strings.TrimSpace(site), strings.TrimSpace(reason), plannedDate
	c.Materials = append([]ControlledMaterial(nil), materials...)
	for i := range c.Materials {
		c.Materials[i].ID = c.ID + "-material-" + itoa(i+1)
	}
	c.Counts = nil
	c.CountDifferenceDetails = nil
	c.Risk = nil
	c.Status = StatusDraft
	c.touch(now)
	return nil
}
