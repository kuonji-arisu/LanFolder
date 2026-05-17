package server

import (
	"errors"
	"net"
	"net/http"
	"time"

	"lanfolder/internal/share"
)

const sessionCookieName = "lf_session"
const requestCookieName = "lf_request"
const requestCookiePath = "/api/access"

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
	requestToken := ""
	if cookie, err := r.Cookie(requestCookieName); err == nil {
		requestToken = cookie.Value
	}
	if requestToken == "" {
		var err error
		requestToken, err = share.NewAccessRequestToken()
		if err != nil {
			writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
			return
		}
	}
	req, created, err := s.access.CreateRequest(requestToken, ip, r.UserAgent())
	if err != nil {
		if errors.Is(err, share.ErrAccessRequestLimited) {
			clearAccessRequestCookie(w)
			setAccessLog(r, "访问请求过于频繁", ip, ip, "")
			writeErrorCode(w, http.StatusTooManyRequests, "access_request_limited", nil)
			return
		}
		writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
		return
	}
	setAccessRequestCookie(w, requestToken, req.ExpiresAt)
	if created {
		setAccessLog(r, "请求访问", req.IP, req.IP, "")
		s.notifyAccessRequest()
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, map[string]any{
		"expiresAt": req.ExpiresAt,
	})
}

func (s *Server) handleAccessPoll(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(requestCookieName)
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusOK, share.AccessPollResult{State: share.AccessPollExpired})
		return
	}
	result, token, err := s.access.Poll(cookie.Value)
	if err != nil {
		writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
		return
	}
	if result.State == share.AccessPollExpired && !s.invalidPolls.Allow(clientIP(r)) {
		clearAccessRequestCookie(w)
		setAccessLog(r, "访问轮询过于频繁", "", "", "")
		writeErrorCode(w, http.StatusTooManyRequests, "access_request_limited", nil)
		return
	}
	if result.State == share.AccessPollApproved && token != "" {
		clearAccessRequestCookie(w)
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	} else if result.State == share.AccessPollDenied || result.State == share.AccessPollExpired {
		clearAccessRequestCookie(w)
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
	s.addLog(newAccessEventLog("批准访问", req.IP, req.IP, ""))
	return nil
}

func (s *Server) DenyAccessRequest(id string) error {
	req, err := s.access.Deny(id)
	if err != nil {
		return err
	}
	s.addLog(newAccessEventLog("拒绝访问", req.IP, req.IP, ""))
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

func setAccessRequestCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	maxAge, ok := accessRequestCookieMaxAge(expiresAt)
	if !ok {
		clearAccessRequestCookie(w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     requestCookieName,
		Value:    token,
		Path:     requestCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func accessRequestCookieMaxAge(expiresAt time.Time) (int, bool) {
	remaining := time.Until(expiresAt)
	if remaining <= 0 {
		return 0, false
	}
	seconds := int(remaining.Seconds())
	if seconds <= 0 {
		return 1, true
	}
	return seconds, true
}

func clearAccessRequestCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     requestCookieName,
		Value:    "",
		Path:     requestCookiePath,
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
