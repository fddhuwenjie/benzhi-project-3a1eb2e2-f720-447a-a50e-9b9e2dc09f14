package web

import (
	"net/http"

	"benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14/internal/application"
)

func (s *Server) handleAction(w http.ResponseWriter, r *http.Request, caseID, action string) {
	var result *application.MutationResult
	var err error
	switch action {
	case "correct":
		var command application.CorrectCommand
		if !s.readCommand(w, r, &command) {
			return
		}
		command.CaseID = caseID
		result, err = s.service.Correct(r.Context(), command)
	case "count":
		var command application.CountCommand
		if !s.readCommand(w, r, &command) {
			return
		}
		command.CaseID = caseID
		result, err = s.service.Count(r.Context(), command)
	case "risk":
		var command application.RiskCommand
		if !s.readCommand(w, r, &command) {
			return
		}
		command.CaseID = caseID
		result, err = s.service.AssessRisk(r.Context(), command)
	case "review":
		var command application.ReviewCommand
		if !s.readCommand(w, r, &command) {
			return
		}
		command.CaseID = caseID
		result, err = s.service.Review(r.Context(), command)
	case "destruction":
		var command application.DestructionCommand
		if !s.readCommand(w, r, &command) {
			return
		}
		command.CaseID = caseID
		result, err = s.service.RecordDestruction(r.Context(), command)
	case "verification":
		var command application.VerificationCommand
		if !s.readCommand(w, r, &command) {
			return
		}
		command.CaseID = caseID
		result, err = s.service.Verify(r.Context(), command)
	case "remediation":
		var command application.RemediationCommand
		if !s.readCommand(w, r, &command) {
			return
		}
		command.CaseID = caseID
		result, err = s.service.Remediate(r.Context(), command)
	case "archive":
		var command application.ArchiveCommand
		if !s.readCommand(w, r, &command) {
			return
		}
		command.CaseID = caseID
		result, err = s.service.Archive(r.Context(), command)
	default:
		writeError(w, http.StatusNotFound, "not_found", "操作不存在", nil, 0)
		return
	}
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) readCommand(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := decodeJSON(w, r, target); err != nil {
		writeApplicationError(w, err)
		return false
	}
	return true
}
