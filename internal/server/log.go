package server

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
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
		s.addLog(entry)
	})
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

func newAccessEventLog(action, remote, target, detail string) LogEntry {
	return LogEntry{
		Time:   time.Now(),
		Method: "DESKTOP",
		Path:   "/api/access",
		Remote: remote,
		Status: http.StatusOK,
		Action: action,
		Target: target,
		Detail: detail,
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

func readableAccessLog(r *http.Request, status int, metadata *accessLogMetadata) (string, string, string, string, bool) {
	if metadata.Action != "" {
		return metadata.Action, metadata.Target, metadata.TargetPath, metadata.Detail, true
	}

	switch {
	case r.Method == http.MethodGet && (r.URL.Path == "/" || r.URL.Path == "/index.html" || r.URL.Path == "/web.html"):
		return "打开分享页", "", "", "", true
	case r.Method == http.MethodGet && r.URL.Path == "/api/list":
		target, targetPath := logTarget(r.URL.Query().Get("path"))
		return "浏览", target, targetPath, "", true
	case r.Method == http.MethodGet && r.URL.Path == "/api/download":
		target, targetPath := logTarget(r.URL.Query().Get("path"))
		return "下载", target, targetPath, "", true
	case r.Method == http.MethodPost && r.URL.Path == "/api/upload":
		target, targetPath := logTarget(r.URL.Query().Get("path"))
		return "上传", target, targetPath, "", true
	case strings.HasPrefix(r.URL.Path, "/api/") && status >= http.StatusBadRequest:
		target, targetPath := apiLogTarget(r)
		return "请求失败", target, targetPath, "", true
	default:
		return "", "", "", "", false
	}
}

func logTarget(path string) (string, string) {
	full := logDisplayPath(path)
	if full == "根目录" {
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
		return "根目录"
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
