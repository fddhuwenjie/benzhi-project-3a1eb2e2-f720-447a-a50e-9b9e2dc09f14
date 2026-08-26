package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

const RiskRuleVersion = "risk-rules-v1"

func (c *RetirementCase) AssessRisk(siteConditions, protectiveMeasures []string, now time.Time) (*RiskAssessment, error) {
	return c.AssessRiskWithConfirmations(siteConditions, protectiveMeasures, nil, now)
}

func (c *RetirementCase) AssessRiskWithConfirmations(siteConditions, protectiveMeasures []string, confirmations map[string]string, now time.Time) (*RiskAssessment, error) {
	if err := RequireStatus(c, StatusCounted); err != nil {
		return nil, err
	}
	conditions, measures := stringSet(siteConditions), stringSet(protectiveMeasures)
	assessment := &RiskAssessment{RuleVersion: RiskRuleVersion, SiteConditions: append([]string(nil), siteConditions...), ProtectiveMeasures: append([]string(nil), protectiveMeasures...), WarningConfirmations: map[string]string{}, EvaluatedAt: now.UTC(), Revision: c.Revision}
	for _, material := range c.Materials {
		switch material.HazardClass {
		case "flammable":
			if !conditions["ventilated"] {
				assessment.Findings = append(assessment.Findings, finding("VENTILATION_REQUIRED", "block", "易燃材料处置场所必须具备通风条件"))
			}
			if !measures["fire_extinguisher"] {
				assessment.Findings = append(assessment.Findings, finding("FIRE_CONTROL_MISSING", "block", "易燃材料处置必须配置灭火设施"))
			}
		case "corrosive":
			if !measures["face_shield"] || !measures["chemical_gloves"] {
				assessment.Findings = append(assessment.Findings, finding("CORROSIVE_PPE_MISSING", "block", "腐蚀性材料需要面屏和耐化学手套"))
			}
		case "toxic":
			if !conditions["fume_hood"] {
				assessment.Findings = append(assessment.Findings, finding("FUME_HOOD_REQUIRED", "block", "有毒材料必须在通风柜条件下处置"))
			}
		case "reactive":
			if !conditions["isolated"] {
				assessment.Findings = append(assessment.Findings, finding("ISOLATION_REQUIRED", "block", "反应性材料需要隔离作业区"))
			}
		case "biohazard":
			if material.DisposalMethod != "autoclave" {
				assessment.Findings = append(assessment.Findings, finding("BIO_METHOD_WARNING", "warning", "生物危害材料建议使用高压灭菌"))
			}
		}
		if material.PackageCondition == "leaking" && !measures["spill_kit"] {
			assessment.Findings = append(assessment.Findings, finding("SPILL_KIT_REQUIRED", "block", "泄漏包装必须配备泄漏处置包"))
		}
		if material.PackageCondition == "damaged" {
			assessment.Findings = append(assessment.Findings, finding("DAMAGED_PACKAGE", "warning", "包装受损，转运时应采用二次容器"))
		}
	}
	for code, text := range confirmations {
		if strings.TrimSpace(text) != "" {
			assessment.WarningConfirmations[code] = strings.TrimSpace(text)
		}
	}
	assessment.InputDigest = riskDigest(c, conditions, measures)
	c.Risk = assessment
	if !assessment.HasBlocking() && assessment.WarningsConfirmed() {
		c.Status = StatusPendingReview
	}
	c.touch(now)
	return assessment, nil
}

func (r RiskAssessment) WarningsConfirmed() bool {
	for _, finding := range r.Findings {
		if finding.Severity == "warning" && strings.TrimSpace(r.WarningConfirmations[finding.Code]) == "" {
			return false
		}
	}
	return true
}

func (c *RetirementCase) RiskCurrent() bool {
	if c.Risk == nil {
		return false
	}
	return c.Risk.InputDigest == riskDigest(c, stringSet(c.Risk.SiteConditions), stringSet(c.Risk.ProtectiveMeasures)) && c.Risk.RuleVersion == RiskRuleVersion
}

func riskDigest(c *RetirementCase, conditions, measures map[string]bool) string {
	cond, meas := make([]string, 0, len(conditions)), make([]string, 0, len(measures))
	for k := range conditions {
		if k != "" {
			cond = append(cond, k)
		}
	}
	for k := range measures {
		if k != "" {
			meas = append(meas, k)
		}
	}
	sort.Strings(cond)
	sort.Strings(meas)
	type materialInput struct {
		Code, Hazard, Package, Method string
		Quantity                      float64
	}
	materials := make([]materialInput, 0, len(c.Materials))
	for _, m := range c.Materials {
		materials = append(materials, materialInput{m.MaterialCode, m.HazardClass, m.PackageCondition, m.DisposalMethod, m.DeclaredQuantity})
	}
	payload, _ := json.Marshal(struct {
		Rule                 string
		Materials            []materialInput
		Conditions, Measures []string
		Counts               int64
	}{RiskRuleVersion, materials, cond, meas, int64(len(c.Counts))})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func (r RiskAssessment) HasBlocking() bool {
	for _, item := range r.Findings {
		if item.Severity == "block" {
			return true
		}
	}
	return false
}

func finding(code, severity, message string) RiskFinding {
	return RiskFinding{Code: code, Severity: severity, Message: message}
}

func stringSet(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[strings.TrimSpace(value)] = true
	}
	return result
}
