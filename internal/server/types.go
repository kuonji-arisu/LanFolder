package server

import (
	"time"

	"lanfolder/internal/share"
)

type Config struct {
	Host       string
	Port       int
	Root       string
	Permission share.Permission
	ShowHidden bool
}

type LogEntry struct {
	Time   time.Time `json:"time"`
	Method string    `json:"method"`
	Path   string    `json:"path"`
	Remote string    `json:"remote"`
	Status int       `json:"status"`
	Error  string    `json:"error,omitempty"`
}

type RuntimeStatus struct {
	Running    bool             `json:"running"`
	Host       string           `json:"host"`
	Port       int              `json:"port"`
	Root       string           `json:"root"`
	Permission share.Permission `json:"permission"`
}
