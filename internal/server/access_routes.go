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
		"language":   s.Status().Language,
	})
}

func (s *Server) handleAccessRequest(w http.ResponseWriter, r *http.Request) {
	if !s.accessRequired() {
		setAccessLog(r, logActionRequestNotNeeded, "", "", "")
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
			setAccessLog(r, logActionRequestRateLimit, ip, ip, "")
			writeErrorCode(w, http.StatusTooManyRequests, "access_request_limited", nil)
			return
		}
		writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
		return
	}
	setAccessRequestCookie(w, requestToken, req.ExpiresAt)
	if created {
		setAccessLog(r, logActionRequestAccess, req.IP, req.IP, "")
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
	result, token, session, err := s.access.Poll(cookie.Value)
	if err != nil {
		writeErrorCode(w, http.StatusInternalServerError, "server_error", nil)
		return
	}
	if result.State == share.AccessPollExpired && !s.invalidPolls.Allow(clientIP(r)) {
		clearAccessRequestCookie(w)
		writeErrorCode(w, http.StatusTooManyRequests, "access_request_limited", nil)
		return
	}
	if result.State == share.AccessPollApproved && token != "" {
		clearAccessRequestCookie(w)
		setSessionCookie(w, token, session)
	} else if result.State == share.AccessPollDenied || result.State == share.AccessPollExpired {
		clearAccessRequestCookie(w)
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
	_, err := s.access.Approve(id)
	return err
}

func (s *Server) DenyAccessRequest(id string) error {
	_, err := s.access.Deny(id)
	return err
}

func (s *Server) RevokeAccessSession(id string) bool {
	_, ok := s.access.RevokeSession(id)
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

func setSessionCookie(w http.ResponseWriter, token string, session share.AccessSession) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	if session.ExpiresAt != nil {
		maxAge, ok := sessionCookieMaxAge(*session.ExpiresAt)
		if !ok {
			clearSessionCookie(w)
			return
		}
		cookie.MaxAge = maxAge
	}
	http.SetCookie(w, cookie)
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
	return cookieMaxAge(expiresAt)
}

func sessionCookieMaxAge(expiresAt time.Time) (int, bool) {
	return cookieMaxAge(expiresAt)
}

func cookieMaxAge(expiresAt time.Time) (int, bool) {
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
