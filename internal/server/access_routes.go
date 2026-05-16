package server

import (
	"errors"
	"net"
	"net/http"

	"lanfolder/internal/share"
)

const sessionCookieName = "lf_session"

func (s *Server) handleAccessStatus(w http.ResponseWriter, r *http.Request) {
	required := s.accessRequired()
	authorized := !required || s.requestAuthorized(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"required":   required,
		"authorized": authorized,
	})
}

func (s *Server) handleAccessRequest(w http.ResponseWriter, r *http.Request) {
	if !s.accessRequired() {
		writeErrorCode(w, http.StatusBadRequest, "access_not_required", nil)
		return
	}
	req, created, err := s.access.CreateRequest(clientIP(r), r.UserAgent())
	if err != nil {
		if errors.Is(err, share.ErrAccessRequestLimited) {
			writeErrorCode(w, http.StatusTooManyRequests, "access_request_limited", nil)
			return
		}
		writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
		return
	}
	if created {
		s.notifyAccessRequest()
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, map[string]any{
		"id":        req.ID,
		"code":      req.Code,
		"expiresAt": req.ExpiresAt,
	})
}

func (s *Server) handleAccessPoll(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErrorCode(w, http.StatusBadRequest, "invalid_request", nil)
		return
	}
	result, token, err := s.access.Poll(id)
	if err != nil {
		writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
		return
	}
	if result.State == share.AccessPollApproved && token != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAccessLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		s.access.Revoke(cookie.Value)
	}
	clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) PendingAccessRequests() []share.AccessRequest {
	return s.access.Pending()
}

func (s *Server) AccessSessions() []share.AccessSession {
	return s.access.Sessions()
}

func (s *Server) ApproveAccessRequest(id string) error {
	return s.access.Approve(id)
}

func (s *Server) DenyAccessRequest(id string) error {
	return s.access.Deny(id)
}

func (s *Server) RevokeAccessSession(id string) bool {
	return s.access.RevokeSession(id)
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
