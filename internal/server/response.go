package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"lanfolder/internal/share"
)

type apiError struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeErrorMessage(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, apiError{Error: message})
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, share.ErrPermissionDenied):
		writeErrorMessage(w, http.StatusForbidden, "permission denied")
	case errors.Is(err, share.ErrInvalidPath), errors.Is(err, share.ErrPathEscape), errors.Is(err, share.ErrCannotDeleteRoot):
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, share.ErrNotFound):
		writeErrorMessage(w, http.StatusNotFound, "not found")
	default:
		writeErrorMessage(w, http.StatusInternalServerError, err.Error())
	}
}
