package domain

import (
	"math"
	"strings"
	"time"
)

func (c *RetirementCase) AddCount(counterID string, observations []CountObservation, reason string, now time.Time) error {
	return c.AddCountDetailed(counterID, observations, reason, nil, now)
}

func (c *RetirementCase) AddCountDetailed(counterID string, observations []CountObservation, reason string, explanations map[string]string, now time.Time) error {
	if err := RequireStatus(c, StatusDraft); err != nil {
		return err
	}
	counterID = strings.TrimSpace(counterID)
	if counterID == "" || counterID == c.OwnerID {
		return Invalid("counter_not_independent", "清点人员不能为空且不能是材料责任人")
	}
	for _, prior := range c.Counts {
		if prior.CounterID == counterID {
			return Invalid("duplicate_counter", "两次清点必须由不同人员完成")
		}
	}
	if len(observations) != len(c.Materials) {
		return Invalid("incomplete_count", "清点必须覆盖全部材料")
	}
	byCode := map[string]CountObservation{}
	for _, observation := range observations {
		if observation.Quantity < 0 {
			return Invalid("invalid_quantity", "实测数量不能小于 0")
		}
		if _, exists := byCode[observation.MaterialCode]; exists {
			return Invalid("duplicate_observation", "同一材料不能重复清点")
		}
		byCode[observation.MaterialCode] = observation
	}
	different := false
	for _, material := range c.Materials {
		observation, ok := byCode[material.MaterialCode]
		if !ok {
			return Invalid("incomplete_count", "清点缺少材料 "+material.MaterialCode)
		}
		if math.Abs(observation.Quantity-material.DeclaredQuantity) > 0.000001 || observation.PackageCondition != material.PackageCondition {
			different = true
		}
	}
	if len(c.Counts) == 1 {
		first := map[string]CountObservation{}
		for _, observation := range c.Counts[0].Observations {
			first[observation.MaterialCode] = observation
		}
		for _, observation := range observations {
			prior := first[observation.MaterialCode]
			if math.Abs(prior.Quantity-observation.Quantity) > 0.000001 || prior.PackageCondition != observation.PackageCondition {
				different = true
			}
		}
	}
	if len(c.Counts) == 1 {
		first := map[string]CountObservation{}
		for _, observation := range c.Counts[0].Observations {
			first[observation.MaterialCode] = observation
		}
		for _, material := range c.Materials {
			prior, current := first[material.MaterialCode], byCode[material.MaterialCode]
			if math.Abs(prior.Quantity-current.Quantity) > 0.000001 || prior.PackageCondition != current.PackageCondition {
				if strings.TrimSpace(explanations[material.MaterialCode]) == "" {
					return Invalid("unexplained_difference", "存在未解释的材料差异："+material.MaterialCode)
				}
			}
		}
	} else if different && strings.TrimSpace(reason) == "" {
		return Invalid("unexplained_difference", "存在数量或包装差异，必须填写差异说明")
	}
	cleanExplanations := map[string]string{}
	for code, explanation := range explanations {
		if strings.TrimSpace(explanation) != "" {
			cleanExplanations[code] = strings.TrimSpace(explanation)
		}
	}
	c.Counts = append(c.Counts, CountConfirmation{ID: c.ID + "-count-" + itoa(len(c.Counts)+1), CounterID: counterID, Observations: append([]CountObservation(nil), observations...), DifferenceReason: strings.TrimSpace(reason), DifferenceExplanations: cleanExplanations, ConfirmedAt: now.UTC()})
	if len(c.Counts) == 2 {
		c.CountDifferenceDetails = c.DifferenceDetails()
		c.Status = StatusCounted
	}
	c.touch(now)
	return nil
}

func (c *RetirementCase) CountDifferences() []string {
	if len(c.Counts) < 2 {
		return nil
	}
	first := map[string]CountObservation{}
	for _, item := range c.Counts[0].Observations {
		first[item.MaterialCode] = item
	}
	var differences []string
	for _, item := range c.Counts[1].Observations {
		prior := first[item.MaterialCode]
		if math.Abs(prior.Quantity-item.Quantity) > 0.000001 || prior.PackageCondition != item.PackageCondition {
			differences = append(differences, item.MaterialCode)
		}
	}
	return differences
}

func (c *RetirementCase) DifferenceDetails() []CountDifference {
	if len(c.Counts) < 2 {
		return nil
	}
	first, second := map[string]CountObservation{}, map[string]CountObservation{}
	for _, item := range c.Counts[0].Observations {
		first[item.MaterialCode] = item
	}
	for _, item := range c.Counts[1].Observations {
		second[item.MaterialCode] = item
	}
	result := make([]CountDifference, 0)
	for _, material := range c.Materials {
		a, b := first[material.MaterialCode], second[material.MaterialCode]
		if math.Abs(a.Quantity-b.Quantity) <= 0.000001 && a.PackageCondition == b.PackageCondition {
			continue
		}
		explanation := c.Counts[1].DifferenceExplanations[material.MaterialCode]
		if explanation == "" {
			explanation = c.Counts[1].DifferenceReason
		}
		result = append(result, CountDifference{MaterialCode: material.MaterialCode, Unit: material.Unit, DeclaredQuantity: material.DeclaredQuantity, FirstQuantity: a.Quantity, SecondQuantity: b.Quantity, QuantityDelta: b.Quantity - a.Quantity, FirstPackageCondition: a.PackageCondition, SecondPackageCondition: b.PackageCondition, FirstCounterID: c.Counts[0].CounterID, SecondCounterID: c.Counts[1].CounterID, Explanation: explanation})
	}
	return result
}
