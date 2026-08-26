package domain

import "time"

type Status string

const (
	StatusDraft         Status = "draft"
	StatusCounted       Status = "counted"
	StatusPendingReview Status = "pending_review"
	StatusApproved      Status = "approved"
	StatusDestroyed     Status = "destroyed"
	StatusRemediation   Status = "remediation"
	StatusVerified      Status = "verified"
	StatusArchived      Status = "archived"
)

type ControlledMaterial struct {
	ID               string  `json:"id"`
	MaterialCode     string  `json:"material_code"`
	DisplayName      string  `json:"display_name"`
	HazardClass      string  `json:"hazard_class"`
	DeclaredQuantity float64 `json:"declared_quantity"`
	Unit             string  `json:"unit"`
	PackageCondition string  `json:"package_condition"`
	DisposalMethod   string  `json:"disposal_method"`
}

type CountObservation struct {
	MaterialCode     string  `json:"material_code"`
	Quantity         float64 `json:"quantity"`
	PackageCondition string  `json:"package_condition"`
}

type CountConfirmation struct {
	ID                     string             `json:"id"`
	CounterID              string             `json:"counter_id"`
	Observations           []CountObservation `json:"observations"`
	DifferenceReason       string             `json:"difference_reason,omitempty"`
	DifferenceExplanations map[string]string  `json:"difference_explanations,omitempty"`
	ConfirmedAt            time.Time          `json:"confirmed_at"`
}

type CountDifference struct {
	MaterialCode           string  `json:"material_code"`
	Unit                   string  `json:"unit"`
	DeclaredQuantity       float64 `json:"declared_quantity"`
	FirstQuantity          float64 `json:"first_quantity"`
	SecondQuantity         float64 `json:"second_quantity"`
	QuantityDelta          float64 `json:"quantity_delta"`
	FirstPackageCondition  string  `json:"first_package_condition"`
	SecondPackageCondition string  `json:"second_package_condition"`
	FirstCounterID         string  `json:"first_counter_id"`
	SecondCounterID        string  `json:"second_counter_id"`
	Explanation            string  `json:"explanation"`
}

type RiskFinding struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type RiskAssessment struct {
	RuleVersion          string            `json:"rule_version"`
	InputDigest          string            `json:"input_digest"`
	Revision             int64             `json:"revision"`
	Findings             []RiskFinding     `json:"findings"`
	WarningConfirmations map[string]string `json:"warning_confirmations,omitempty"`
	ProtectiveMeasures   []string          `json:"protective_measures"`
	SiteConditions       []string          `json:"site_conditions"`
	EvaluatedAt          time.Time         `json:"evaluated_at"`
}

type ReviewDecision struct {
	ID            string    `json:"id"`
	ReviewerID    string    `json:"reviewer_id"`
	Approved      bool      `json:"approved"`
	Reason        string    `json:"reason,omitempty"`
	AllowedFields []string  `json:"allowed_fields,omitempty"`
	DecidedAt     time.Time `json:"decided_at"`
}

type WitnessRecord struct {
	ID             string    `json:"id"`
	Method         string    `json:"method"`
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    time.Time `json:"completed_at"`
	WitnessIDs     []string  `json:"witness_ids"`
	EvidenceDigest string    `json:"evidence_digest"`
	Notes          string    `json:"notes,omitempty"`
}

type VerificationRecord struct {
	ID              string    `json:"id"`
	CheckName       string    `json:"check_name"`
	Threshold       float64   `json:"threshold"`
	MeasuredValue   float64   `json:"measured_value"`
	Result          string    `json:"result"`
	ReviewerID      string    `json:"reviewer_id"`
	RemediationNote string    `json:"remediation_note,omitempty"`
	VerifiedAt      time.Time `json:"verified_at"`
	RemediationID   string    `json:"remediation_id,omitempty"`
}

type ArchiveSummary struct {
	MaterialCount      int       `json:"material_count"`
	MaterialCodes      []string  `json:"material_codes"`
	RiskFindingCodes   []string  `json:"risk_finding_codes"`
	ApprovedBy         string    `json:"approved_by"`
	Witnesses          []string  `json:"witnesses"`
	VerificationChecks []string  `json:"verification_checks"`
	EventCount         int       `json:"event_count"`
	ChainHead          string    `json:"chain_head"`
	GeneratedAt        time.Time `json:"generated_at"`
	Revision           int64     `json:"revision"`
	PreviewDigest      string    `json:"preview_digest"`
	SnapshotDigest     string    `json:"snapshot_digest"`
}

type RetirementCase struct {
	ID                     string               `json:"id"`
	Site                   string               `json:"site"`
	Reason                 string               `json:"reason"`
	OwnerID                string               `json:"owner_id"`
	PlannedDate            string               `json:"planned_date"`
	Status                 Status               `json:"status"`
	Revision               int64                `json:"revision"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
	ArchivedAt             *time.Time           `json:"archived_at,omitempty"`
	Materials              []ControlledMaterial `json:"materials"`
	Counts                 []CountConfirmation  `json:"counts"`
	CountDifferenceDetails []CountDifference    `json:"count_differences,omitempty"`
	Risk                   *RiskAssessment      `json:"risk,omitempty"`
	RiskBaseline           *RiskAssessment      `json:"risk_baseline,omitempty"`
	Reviews                []ReviewDecision     `json:"reviews"`
	Witness                *WitnessRecord       `json:"witness,omitempty"`
	Verifications          []VerificationRecord `json:"verifications"`
	RemediationNotes       []string             `json:"remediation_notes"`
	Archive                *ArchiveSummary      `json:"archive,omitempty"`
}

type CreateInput struct {
	ID          string               `json:"id"`
	Site        string               `json:"site"`
	Reason      string               `json:"reason"`
	OwnerID     string               `json:"owner_id"`
	PlannedDate string               `json:"planned_date"`
	Materials   []ControlledMaterial `json:"materials"`
}
