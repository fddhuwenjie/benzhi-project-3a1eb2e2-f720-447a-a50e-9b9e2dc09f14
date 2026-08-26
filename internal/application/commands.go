package application

import (
	"time"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

type CommandMeta struct {
	ActorID          string `json:"actor_id"`
	Role             string `json:"role"`
	RequestID        string `json:"request_id"`
	ExpectedRevision int64  `json:"expected_revision"`
}

type CreateCommand struct {
	CommandMeta
	Site        string                      `json:"site"`
	Reason      string                      `json:"reason"`
	OwnerID     string                      `json:"owner_id"`
	PlannedDate string                      `json:"planned_date"`
	Materials   []domain.ControlledMaterial `json:"materials"`
}
type CorrectCommand struct {
	CommandMeta
	CaseID      string                      `json:"-"`
	Site        string                      `json:"site"`
	Reason      string                      `json:"reason"`
	PlannedDate string                      `json:"planned_date"`
	Materials   []domain.ControlledMaterial `json:"materials"`
	OpinionID   string                      `json:"opinion_id,omitempty"`
	Patch       map[string]any              `json:"patch,omitempty"`
}
type CountCommand struct {
	CommandMeta
	CaseID                 string                    `json:"-"`
	CounterID              string                    `json:"counter_id"`
	DifferenceReason       string                    `json:"difference_reason"`
	DifferenceExplanations map[string]string         `json:"difference_explanations,omitempty"`
	Observations           []domain.CountObservation `json:"observations"`
}
type RiskCommand struct {
	CommandMeta
	CaseID               string            `json:"-"`
	SiteConditions       []string          `json:"site_conditions"`
	ProtectiveMeasures   []string          `json:"protective_measures"`
	WarningConfirmations map[string]string `json:"warning_confirmations,omitempty"`
}
type ReviewCommand struct {
	CommandMeta
	CaseID        string   `json:"-"`
	ReviewerID    string   `json:"reviewer_id"`
	Reason        string   `json:"reason"`
	Approved      bool     `json:"approved"`
	AllowedFields []string `json:"allowed_fields,omitempty"`
}
type DestructionCommand struct {
	CommandMeta
	CaseID         string    `json:"-"`
	Method         string    `json:"method"`
	EvidenceDigest string    `json:"evidence_digest"`
	Notes          string    `json:"notes"`
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    time.Time `json:"completed_at"`
	WitnessIDs     []string  `json:"witness_ids"`
}
type VerificationCommand struct {
	CommandMeta
	CaseID        string  `json:"-"`
	CheckName     string  `json:"check_name"`
	ReviewerID    string  `json:"reviewer_id"`
	Threshold     float64 `json:"threshold"`
	MeasuredValue float64 `json:"measured_value"`
}
type RemediationCommand struct {
	CommandMeta
	CaseID string `json:"-"`
	Note   string `json:"note"`
}
type ArchiveCommand struct {
	CommandMeta
	CaseID            string `json:"-"`
	PreviewDigest     string `json:"preview_digest,omitempty"`
	ExpectedChainHead string `json:"expected_chain_head,omitempty"`
}

type MutationResult struct {
	Case     *domain.RetirementCase `json:"case"`
	Replayed bool                   `json:"replayed"`
}

type Detail struct {
	Case     *domain.RetirementCase `json:"case"`
	Timeline any                    `json:"timeline"`
	ChainOK  bool                   `json:"chain_ok"`
}

type ListResult struct {
	Items  []*domain.RetirementCase `json:"items"`
	Total  int                      `json:"total"`
	Stats  ListStats                `json:"stats"`
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
}

type ListStats struct {
	ByStatus      map[string]int `json:"by_status"`
	ByHazardClass map[string]int `json:"by_hazard_class"`
	Overdue       int            `json:"overdue"`
}
