package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"lanfolder/internal/share"
)

const (
	maxUploadBytes       int64 = 512 << 20
	multipartMemoryBytes int64 = 32 << 20
)

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := s.Status()
	writeJSON(w, http.StatusOK, map[string]any{
		"running":     true,
		"port":        status.Port,
		"permission":  status.Permission,
		"permissions": share.PermissionOptions(),
	})
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	result, err := s.manager.List(r.URL.Query().Get("path"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	file, entry, err := s.manager.OpenForDownload(r.URL.Query().Get("path"))
	if err != nil {
		writeError(w, err)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", entry.Name))
	http.ServeContent(w, r, entry.Name, entry.ModTime, file)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(multipartMemoryBytes); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeErrorCode(w, http.StatusRequestEntityTooLarge, "file_too_large", map[string]any{"maxBytes": maxUploadBytes})
			return
		}
		writeErrorCode(w, http.StatusBadRequest, "invalid_request", nil)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		if one, _, err := r.FormFile("file"); err == nil {
			_ = one.Close()
			files = r.MultipartForm.File["file"]
		}
	}
	if len(files) == 0 {
		writeErrorCode(w, http.StatusBadRequest, "no_files_uploaded", nil)
		return
	}
	created := make([]share.Entry, 0, len(files))
	for _, header := range files {
		entry, err := s.manager.SaveUpload(r.URL.Query().Get("path"), header)
		if err != nil {
			writeError(w, err)
			return
		}
		created = append(created, entry)
	}
	writeJSON(w, http.StatusCreated, map[string]any{"entries": created})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "invalid_request", nil)
		return
	}
	if err := s.manager.Delete(body.Path); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleMkdir(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "invalid_request", nil)
		return
	}
	entry, err := s.manager.Mkdir(body.Path, body.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}
