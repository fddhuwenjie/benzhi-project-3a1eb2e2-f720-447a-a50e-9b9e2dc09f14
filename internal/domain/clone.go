package domain

import "encoding/json"

func Clone(c *RetirementCase) *RetirementCase {
	if c == nil {
		return nil
	}
	data, _ := json.Marshal(c)
	var copied RetirementCase
	_ = json.Unmarshal(data, &copied)
	return &copied
}
