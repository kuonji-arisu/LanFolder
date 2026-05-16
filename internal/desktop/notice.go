package desktop

import "time"

type NoticeLevel string

const (
	NoticeInfo    NoticeLevel = "info"
	NoticeSuccess NoticeLevel = "success"
	NoticeWarning NoticeLevel = "warning"
	NoticeError   NoticeLevel = "error"
)

type NoticeSource string

const (
	NoticeSourceCommand NoticeSource = "command"
	NoticeSourceStartup NoticeSource = "startup"
	NoticeSourceSystem  NoticeSource = "system"
)

type ErrorPayload struct {
	Error  string         `json:"error"`
	Params map[string]any `json:"params,omitempty"`
}

type Notice struct {
	ID        string        `json:"id"`
	Level     NoticeLevel   `json:"level"`
	Source    NoticeSource  `json:"source"`
	Error     *ErrorPayload `json:"error,omitempty"`
	Message   string        `json:"message,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
}
