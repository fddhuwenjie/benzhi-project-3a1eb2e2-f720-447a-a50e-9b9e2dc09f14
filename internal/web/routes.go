package web

import (
	"net/http"
	"strconv"
	"strings"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/domain"
)

func (s *Server) HandleCases(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		result, err := s.service.List(r.Context(), application.ListFilter{Status: domain.Status(r.URL.Query().Get("status")), Site: r.URL.Query().Get("site"), PlannedFrom: r.URL.Query().Get("planned_date_from"), PlannedTo: r.URL.Query().Get("planned_date_to"), HazardClass: r.URL.Query().Get("hazard_class"), Limit: limit, Offset: offset})
		if err != nil {
			writeApplicationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	case http.MethodPost:
		var command application.CreateCommand
		if err := decodeJSON(w, r, &command); err != nil {
			writeApplicationError(w, err)
			return
		}
		result, err := s.service.Create(r.Context(), command)
		if err != nil {
			writeApplicationError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, result)
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) HandleCase(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/cases/"), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not_found", "批次不存在", nil, 0)
		return
	}
	caseID := parts[0]
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		result, err := s.service.Get(r.Context(), caseID)
		if err != nil {
			writeApplicationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}
	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, "not_found", "接口不存在", nil, 0)
		return
	}
	if parts[1] == "archive-preview" {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		result, err := s.service.ArchivePreview(r.Context(), caseID)
		if err != nil {
			writeApplicationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}
	if parts[1] == "archive-export" {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		result, err := s.service.ArchiveExport(r.Context(), caseID)
		if err != nil {
			writeApplicationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	s.handleAction(w, r, caseID, parts[1])
}
