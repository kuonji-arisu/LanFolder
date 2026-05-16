package server

import (
	"bytes"
	"io/fs"
	"net/http"
	"strings"
	"time"
)

func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("GET /api/access/status", s.handleAccessStatus)
	mux.HandleFunc("POST /api/access/request", s.handleAccessRequest)
	mux.HandleFunc("GET /api/access/poll", s.handleAccessPoll)
	mux.HandleFunc("POST /api/access/logout", s.handleAccessLogout)
	mux.HandleFunc("GET /api/list", s.handleList)
	mux.HandleFunc("GET /api/messages", s.handleMessages)
	mux.HandleFunc("GET /api/download", s.handleDownload)
	mux.HandleFunc("POST /api/messages", s.handleSendMessage)
	mux.HandleFunc("DELETE /api/messages", s.handleClearMessages)
	mux.HandleFunc("POST /api/upload", s.handleUpload)
	mux.HandleFunc("POST /api/delete", s.handleDelete)
	mux.HandleFunc("POST /api/mkdir", s.handleMkdir)
	mux.HandleFunc("/", s.handleStatic)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	cleanPath := strings.TrimPrefix(r.URL.Path, "/")
	if cleanPath == "" {
		cleanPath = "web.html"
	}
	if cleanPath == "index.html" {
		cleanPath = "web.html"
	}
	if cleanPath == "favicon.ico" {
		cleanPath = "appicon.png"
	}
	if strings.Contains(cleanPath, "..") {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(s.staticFS, cleanPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, cleanPath, time.Time{}, bytes.NewReader(data))
}
