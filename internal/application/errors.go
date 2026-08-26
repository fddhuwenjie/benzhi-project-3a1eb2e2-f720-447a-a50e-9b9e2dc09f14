package application

import (
	"errors"
	"fmt"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

type Error struct {
	Code            string            `json:"code"`
	Message         string            `json:"message"`
	Fields          map[string]string `json:"fields,omitempty"`
	CurrentRevision int64             `json:"current_revision,omitempty"`
}

func (e *Error) Error() string { return e.Message }

func NotFound() *Error { return &Error{Code: "not_found", Message: "未找到指定批次"} }

func Conflict(revision int64) *Error {
	return &Error{Code: "revision_conflict", Message: "批次已被其他操作更新，请刷新后重试", CurrentRevision: revision}
}

func MapError(err error) error {
	if err == nil {
		return nil
	}
	var app *Error
	if errors.As(err, &app) {
		return app
	}
	var business *domain.Error
	if errors.As(err, &business) {
		return &Error{Code: business.Code, Message: business.Message, Fields: business.Fields}
	}
	return fmt.Errorf("application operation failed: %w", err)
}
