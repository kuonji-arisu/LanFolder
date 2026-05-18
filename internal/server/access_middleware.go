package server

import (
	"net/http"
	"strings"
)

func (s *Server) accessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.accessRequired() || accessRouteAllowed(r) || s.requestAuthorized(r) {
			next.ServeHTTP(w, r)
			return
		}
		writeErrorCode(w, http.StatusUnauthorized, "access_required", nil)
	})
}

func (s *Server) accessRequired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.AccessApproval
}

func (s *Server) requestAuthorized(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	return err == nil && s.access.Validate(cookie.Value)
}

func (s *Server) notifyAccessRequest() {
	s.mu.RLock()
	fn := s.onAccessRequest
	s.mu.RUnlock()
	if fn != nil {
		fn()
	}
}

func accessRouteAllowed(r *http.Request) bool {
	switch r.URL.Path {
	case "/api/status", "/api/access/status", "/api/access/poll":
		return r.Method == http.MethodGet || r.Method == http.MethodHead
	case "/api/access/request", "/api/access/logout":
		return r.Method == http.MethodPost
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		return false
	}
	return r.Method == http.MethodGet || r.Method == http.MethodHead
}
