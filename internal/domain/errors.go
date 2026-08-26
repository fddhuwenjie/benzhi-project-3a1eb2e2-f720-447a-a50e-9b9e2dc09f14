package domain

import "fmt"

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *Error) Error() string { return e.Message }

func Invalid(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func FieldError(fields map[string]string) *Error {
	return &Error{Code: "validation_failed", Message: "提交内容未通过校验", Fields: fields}
}

func RequireStatus(c *RetirementCase, allowed ...Status) error {
	if c.Status == StatusArchived {
		return Invalid("case_archived", "批次已归档，不能再修改")
	}
	for _, status := range allowed {
		if c.Status == status {
			return nil
		}
	}
	return Invalid("invalid_status", fmt.Sprintf("当前状态 %s 不允许执行该操作", c.Status))
}
