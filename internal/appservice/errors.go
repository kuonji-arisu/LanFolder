package appservice

import (
	"encoding/json"
	"errors"

	"lanfolder/internal/desktop"
)

var (
	errInvalidPort              = errors.New("invalid_port")
	errSharedDirRequired        = errors.New("shared_dir_required")
	errShareNotRunning          = errors.New("share_not_running")
	errAccessApprovalRequired   = errors.New("access_approval_required")
	errAccessRequestUnavailable = errors.New("access_request_unavailable")
)

type commandError struct {
	Code   string         `json:"error"`
	Params map[string]any `json:"params,omitempty"`
}

func newCommandError(err error, params map[string]any) error {
	return commandError{Code: err.Error(), Params: params}
}

func (e commandError) Error() string {
	return e.Code
}

func MarshalCommandError(err error) []byte {
	payload := commandErrorPayload(err)
	if payload == nil {
		return nil
	}
	data, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil
	}
	return data
}

func commandErrorPayload(err error) *desktop.ErrorPayload {
	var commandErr commandError
	if !errors.As(err, &commandErr) {
		return nil
	}
	return &desktop.ErrorPayload{Error: commandErr.Code, Params: commandErr.Params}
}
