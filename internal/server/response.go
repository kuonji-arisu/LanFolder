package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"lanfolder/internal/share"
)

const maxJSONBodyBytes int64 = 16 << 10

type apiError struct {
	Error  string         `json:"error"`
	Params map[string]any `json:"params,omitempty"`
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSONDecodeError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		writeErrorCode(w, http.StatusRequestEntityTooLarge, "request_too_large", map[string]any{"maxBytes": maxJSONBodyBytes})
		return
	}
	writeErrorCode(w, http.StatusBadRequest, "invalid_request", nil)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeErrorCode(w http.ResponseWriter, status int, code string, params map[string]any) {
	writeJSON(w, status, apiError{Error: code, Params: params})
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, share.ErrPermissionDenied):
		writeErrorCode(w, http.StatusForbidden, "permission_denied", nil)
	case errors.Is(err, share.ErrCannotDeleteRoot):
		writeErrorCode(w, http.StatusBadRequest, "cannot_delete_root", nil)
	case errors.Is(err, share.ErrInvalidPath), errors.Is(err, share.ErrPathEscape):
		writeErrorCode(w, http.StatusBadRequest, "invalid_path", nil)
	case errors.Is(err, share.ErrInvalidMessage):
		writeErrorCode(w, http.StatusBadRequest, "invalid_message", nil)
	case errors.Is(err, share.ErrInvalidFilename):
		writeErrorCode(w, http.StatusBadRequest, "invalid_filename", map[string]any{"maxBytes": share.MaxFilenameBytes})
	case errors.Is(err, share.ErrNotFound):
		writeErrorCode(w, http.StatusNotFound, "not_found", nil)
	default:
		writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
	}
}
