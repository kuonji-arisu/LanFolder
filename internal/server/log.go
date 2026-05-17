package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"lanfolder/i18n"
)

const (
	logActionBrowse           = "log.action.browse"
	logActionDownload         = "common.download"
	logActionUpload           = "common.upload"
	logActionRequestFailed    = "log.action.requestFailed"
	logActionSendMessage      = "log.action.sendMessage"
	logActionClearMessages    = "log.action.clearMessages"
	logActionDelete           = "common.delete"
	logActionMkdir            = "log.action.mkdir"
	logActionRequestNotNeeded = "log.action.requestNotRequired"
	logActionRequestRateLimit = "log.action.requestRateLimited"
	logActionRequestAccess    = "log.action.requestAccess"
	logDetailFilesPrefix      = "file.count:"
	logRootTarget             = "log.target.root"
)

func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metadata := &accessLogMetadata{}
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		req := r.WithContext(context.WithValue(r.Context(), accessLogMetadataKey{}, metadata))
		next.ServeHTTP(rec, req)
		entry, ok := newLogEntry(req, rec.status, metadata)
		if !ok {
			return
		}
		s.localizeLogEntry(&entry)
		s.addLog(entry)
	})
}

func (s *Server) localizeLogEntry(entry *LogEntry) {
	s.mu.RLock()
	language := s.config.Language
	s.mu.RUnlock()

	if strings.HasPrefix(entry.Action, "log.action.") || strings.HasPrefix(entry.Action, "common.") {
		entry.Action = i18n.T(language, entry.Action, nil)
	}
	if entry.Target == logRootTarget {
		entry.Target = i18n.T(language, "common.root", nil)
	}
	if entry.TargetPath == logRootTarget {
		entry.TargetPath = i18n.T(language, "common.root", nil)
	}
	if count, ok := logFilesDetailCount(entry.Detail); ok {
		entry.Detail = i18n.T(language, "file.count", map[string]any{"count": count})
	}
}

func newLogEntry(r *http.Request, status int, metadata *accessLogMetadata) (LogEntry, bool) {
	action, target, targetPath, detail, ok := readableAccessLog(r, status, metadata)
	if !ok {
		return LogEntry{}, false
	}
	return LogEntry{
		Time:       time.Now(),
		Method:     r.Method,
		Path:       r.URL.RequestURI(),
		Remote:     remoteIP(r.RemoteAddr),
		Status:     status,
		Action:     action,
		Target:     target,
		TargetPath: targetPath,
		Detail:     detail,
	}, true
}

func (s *Server) addLog(entry LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append([]LogEntry{entry}, s.logs...)
	if len(s.logs) > s.maxLogs {
		s.logs = s.logs[:s.maxLogs]
	}
}

type accessLogMetadataKey struct{}

type accessLogMetadata struct {
	Action     string
	Target     string
	TargetPath string
	Detail     string
}

func setAccessLog(r *http.Request, action, target, targetPath, detail string) {
	metadata, ok := r.Context().Value(accessLogMetadataKey{}).(*accessLogMetadata)
	if !ok {
		return
	}
	metadata.Action = action
	metadata.Target = target
	metadata.TargetPath = targetPath
	metadata.Detail = detail
}

func logFilesDetailCount(detail string) (int, bool) {
	if !strings.HasPrefix(detail, logDetailFilesPrefix) {
		return 0, false
	}
	value := strings.TrimPrefix(detail, logDetailFilesPrefix)
	count, err := strconv.Atoi(value)
	return count, err == nil
}

func readableAccessLog(r *http.Request, status int, metadata *accessLogMetadata) (string, string, string, string, bool) {
	if metadata.Action != "" {
		return metadata.Action, metadata.Target, metadata.TargetPath, metadata.Detail, true
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/list":
		target, targetPath := logTarget(r.URL.Query().Get("path"))
		return logActionBrowse, target, targetPath, "", true
	case r.Method == http.MethodGet && r.URL.Path == "/api/download":
		target, targetPath := logTarget(r.URL.Query().Get("path"))
		return logActionDownload, target, targetPath, "", true
	case r.Method == http.MethodPost && r.URL.Path == "/api/upload":
		target, targetPath := logTarget(r.URL.Query().Get("path"))
		return logActionUpload, target, targetPath, "", true
	case strings.HasPrefix(r.URL.Path, "/api/access/") && r.URL.Path != "/api/access/request":
		return "", "", "", "", false
	case strings.HasPrefix(r.URL.Path, "/api/") && status >= http.StatusBadRequest:
		target, targetPath := apiLogTarget(r)
		return logActionRequestFailed, target, targetPath, "", true
	default:
		return "", "", "", "", false
	}
}

func logTarget(path string) (string, string) {
	full := logDisplayPath(path)
	if full == logRootTarget {
		return full, full
	}
	index := strings.LastIndex(full, "/")
	if index == -1 {
		return full, full
	}
	return full[index+1:], full
}

func logDisplayPath(path string) string {
	clean := strings.Trim(strings.ReplaceAll(path, "\\", "/"), "/")
	if clean == "" {
		return logRootTarget
	}
	return clean
}

func apiLogTarget(r *http.Request) (string, string) {
	switch r.URL.Path {
	case "/api/list", "/api/download", "/api/upload":
		return logTarget(r.URL.Query().Get("path"))
	default:
		return r.URL.Path, r.URL.Path
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func remoteIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
