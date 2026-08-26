package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
)

const maxRequestBody = 1 << 20

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	if contentType := r.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		return &application.Error{Code: "invalid_content_type", Message: "Content-Type 必须为 application/json"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return &application.Error{Code: "invalid_json", Message: "JSON 请求体无效: " + err.Error()}
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return &application.Error{Code: "invalid_json", Message: "请求体只能包含一个 JSON 对象"}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeApplicationError(w http.ResponseWriter, err error) {
	var app *application.Error
	if !errors.As(err, &app) {
		writeError(w, http.StatusInternalServerError, "internal_error", "服务处理请求失败", nil, 0)
		return
	}
	status := http.StatusUnprocessableEntity
	switch app.Code {
	case "not_found":
		status = http.StatusNotFound
	case "forbidden":
		status = http.StatusForbidden
	case "revision_conflict", "request_id_reused":
		status = http.StatusConflict
	case "invalid_json", "invalid_content_type", "validation_failed":
		status = http.StatusBadRequest
	}
	writeError(w, status, app.Code, app.Message, app.Fields, app.CurrentRevision)
}

func writeError(w http.ResponseWriter, status int, code, message string, fields map[string]string, revision int64) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"code": code, "message": message, "fields": fields, "current_revision": revision}})
}

func methodNotAllowed(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "请求方法不受支持", nil, 0)
}
