package share

import "errors"

var (
	ErrInvalidRoot      = errors.New("shared root is not available")
	ErrInvalidPath      = errors.New("invalid path")
	ErrPathEscape       = errors.New("path escapes shared root")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNotFound         = errors.New("not found")
	ErrCannotDeleteRoot = errors.New("cannot delete shared root")
)
