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
		setAccessLog(r, "访问请求无需批准", "", "", "")
		writeErrorCode(w, http.StatusBadRequest, "access_not_required", nil)
		return
	}
	ip := clientIP(r)
	req, created, err := s.access.CreateRequest(ip, r.UserAgent())
	if err != nil {
		if errors.Is(err, share.ErrAccessRequestLimited) {
			setAccessLog(r, "访问请求过于频繁", ip, ip, "")
			writeErrorCode(w, http.StatusTooManyRequests, "access_request_limited", nil)
			return
		}
		writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
		return
	}
	if created {
		setAccessLog(r, "请求访问", req.Code, req.IP, "")
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
		setAccessLog(r, "访问轮询无效", "", "", "")
		writeErrorCode(w, http.StatusBadRequest, "invalid_request", nil)
		return
	}
	result, token, err := s.access.Poll(id)
	if err != nil {
		writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
		return
	}
	if result.State == share.AccessPollExpired && !s.invalidPolls.Allow(clientIP(r)) {
		setAccessLog(r, "访问轮询过于频繁", "", "", "")
		writeErrorCode(w, http.StatusTooManyRequests, "access_request_limited", nil)
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
	if result.State != share.AccessPollPending {
		setAccessLog(r, "访问轮询", string(result.State), "", "")
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAccessLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		s.access.Revoke(cookie.Value)
	}
	setAccessLog(r, "退出访问", "", "", "")
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
	req, err := s.access.Approve(id)
	if err != nil {
		return err
	}
	s.addLog(newAccessEventLog("批准访问", req.IP, req.Code, ""))
	return nil
}

func (s *Server) DenyAccessRequest(id string) error {
	req, err := s.access.Deny(id)
	if err != nil {
		return err
	}
	s.addLog(newAccessEventLog("拒绝访问", req.IP, req.Code, ""))
	return nil
}

func (s *Server) RevokeAccessSession(id string) bool {
	session, ok := s.access.RevokeSession(id)
	if ok {
		s.addLog(newAccessEventLog("撤销授权", session.IP, session.IP, ""))
	}
	return ok
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
