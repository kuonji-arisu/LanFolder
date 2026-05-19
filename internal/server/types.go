package server

import (
	"time"

	"lanfolder/internal/share"
)

type Config struct {
	Host                  string
	Port                  int
	Root                  string
	Permission            share.Permission
	ShowHidden            bool
	AccessApproval        bool
	AccessSessionLifetime share.AccessSessionLifetime
	Language              string
}

type LogEntry struct {
	Time       time.Time `json:"time"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Remote     string    `json:"remote"`
	Status     int       `json:"status"`
	Action     string    `json:"action"`
	Target     string    `json:"target,omitempty"`
	TargetPath string    `json:"targetPath,omitempty"`
	Detail     string    `json:"detail,omitempty"`
	Error      string    `json:"error,omitempty"`
}

type RuntimeStatus struct {
	Running               bool                        `json:"running"`
	Host                  string                      `json:"host"`
	Port                  int                         `json:"port"`
	Root                  string                      `json:"root"`
	Permission            share.Permission            `json:"permission"`
	AccessApproval        bool                        `json:"accessApproval"`
	AccessSessionLifetime share.AccessSessionLifetime `json:"accessSessionLifetime"`
	Language              string                      `json:"language"`
}
