package server

import (
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
)

func (s *Server) secureMiddleware(next http.Handler) http.Handler {
	return securityHeaders(privateNetworkGuard(hostGuard(writeGuard(next))))
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func privateNetworkGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addr, ok := remoteAddr(r.RemoteAddr)
		if !ok || !remoteIPAllowed(addr) {
			writeErrorCode(w, http.StatusForbidden, "network_not_allowed", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func hostGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, ok := requestHost(r.Host)
		if !ok || !hostAllowed(host) {
			writeErrorCode(w, http.StatusForbidden, "bad_host", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSafeMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		if origin != "" && !originMatchesHost(origin, r.Host) {
			writeErrorCode(w, http.StatusForbidden, "bad_origin", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func requestHost(hostport string) (string, bool) {
	if hostport == "" {
		return "", false
	}

	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	} else if strings.HasPrefix(hostport, "[") && strings.Contains(hostport, "]") {
		return "", false
	} else if strings.Count(hostport, ":") > 0 {
		return "", false
	}

	host = strings.Trim(strings.ToLower(host), "[]")
	if host == "" {
		return "", false
	}
	return host, true
}

func hostAllowed(host string) bool {
	if host == "localhost" {
		return true
	}
	addr, err := netip.ParseAddr(host)
	return err == nil && remoteIPAllowed(addr)
}

func originMatchesHost(origin, host string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return strings.EqualFold(u.Host, host)
}

func remoteAddr(remote string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		return netip.Addr{}, false
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

func remoteIPAllowed(addr netip.Addr) bool {
	return addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast()
}
