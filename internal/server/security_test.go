package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"lanfolder/internal/share"
)

func TestSecurityAllowsSameOriginWrite(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionManage)
	resp := serveSecurityRequest(handler, http.MethodPost, "http://192.168.1.10:8899/api/mkdir", "192.168.1.20:12345", "http://192.168.1.10:8899", `{"path":"","name":"docs"}`)

	if resp.Code != http.StatusCreated {
		t.Fatalf("same-origin write status = %d, body = %s", resp.Code, resp.Body.String())
	}
}

func TestSecurityRejectsCrossOriginWrite(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionManage)
	resp := serveSecurityRequest(handler, http.MethodPost, "http://192.168.1.10:8899/api/mkdir", "192.168.1.20:12345", "https://evil.example", `{"path":"","name":"docs"}`)

	assertAPIError(t, resp, http.StatusForbidden, "bad_origin")
}

func TestSecurityAllowsWriteWithoutOrigin(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionManage)
	resp := serveSecurityRequest(handler, http.MethodPost, "http://192.168.1.10:8899/api/mkdir", "192.168.1.20:12345", "", `{"path":"","name":"docs"}`)

	if resp.Code != http.StatusCreated {
		t.Fatalf("write without origin status = %d, body = %s", resp.Code, resp.Body.String())
	}
}

func TestSecurityDoesNotOriginCheckSafeMethods(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionReadOnly)
	resp := serveSecurityRequest(handler, http.MethodGet, "http://192.168.1.10:8899/api/status", "192.168.1.20:12345", "https://evil.example", "")

	if resp.Code != http.StatusOK {
		t.Fatalf("safe cross-origin read status = %d, body = %s", resp.Code, resp.Body.String())
	}
}

func TestSecurityRejectsPublicRemoteAddress(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionReadOnly)
	resp := serveSecurityRequest(handler, http.MethodGet, "http://192.168.1.10:8899/api/status", "8.8.8.8:12345", "", "")

	assertAPIError(t, resp, http.StatusForbidden, "network_not_allowed")
}

func TestSecurityAllowsLoopbackRemoteAddress(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionReadOnly)
	resp := serveSecurityRequest(handler, http.MethodGet, "http://127.0.0.1:8899/api/status", "127.0.0.1:12345", "", "")

	if resp.Code != http.StatusOK {
		t.Fatalf("loopback request status = %d, body = %s", resp.Code, resp.Body.String())
	}
}

func TestSecurityAllowsIPv6LoopbackRemoteAddress(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionReadOnly)
	resp := serveSecurityRequest(handler, http.MethodGet, "http://[::1]:8899/api/status", "[::1]:12345", "", "")

	if resp.Code != http.StatusOK {
		t.Fatalf("IPv6 loopback request status = %d, body = %s", resp.Code, resp.Body.String())
	}
}

func TestSecurityRejectsMalformedRemoteAddress(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionReadOnly)
	resp := serveSecurityRequest(handler, http.MethodGet, "http://192.168.1.10:8899/api/status", "not-a-host", "", "")

	assertAPIError(t, resp, http.StatusForbidden, "network_not_allowed")
}

func TestSecurityRejectsBadHostForWrites(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionManage)
	resp := serveSecurityRequest(handler, http.MethodPost, "http://evil.example/api/mkdir", "192.168.1.20:12345", "", `{"path":"","name":"docs"}`)

	assertAPIError(t, resp, http.StatusForbidden, "bad_host")
}

func TestSecurityRejectsBadHostForReads(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionReadOnly)
	resp := serveSecurityRequest(handler, http.MethodGet, "http://evil.example/api/status", "192.168.1.20:12345", "", "")

	assertAPIError(t, resp, http.StatusForbidden, "bad_host")
}

func TestSecurityHeadersAreApplied(t *testing.T) {
	handler := testSecureHandler(t, share.PermissionReadOnly)
	resp := serveSecurityRequest(handler, http.MethodGet, "http://127.0.0.1:8899/api/status", "127.0.0.1:12345", "", "")

	if got := resp.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	if got := resp.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Fatalf("Referrer-Policy = %q", got)
	}
	if got := resp.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options = %q", got)
	}
}

func testSecureHandler(t *testing.T, permission share.Permission) http.Handler {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "web.html"), []byte("<main>LanFolder</main>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "appicon.png"), []byte("png"), 0644); err != nil {
		t.Fatal(err)
	}
	s := New(os.DirFS(root))
	if err := s.manager.Configure(root, permission, false); err != nil {
		t.Fatal(err)
	}
	s.config = Config{Host: "127.0.0.1", Port: 8899, Root: root, Permission: permission}
	mux := http.NewServeMux()
	s.routes(mux)
	return s.secureMiddleware(mux)
}

func serveSecurityRequest(handler http.Handler, method, target, remoteAddr, origin, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.RemoteAddr = remoteAddr
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

func assertAPIError(t *testing.T, resp *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if resp.Code != status {
		t.Fatalf("status = %d, want %d, body = %s", resp.Code, status, resp.Body.String())
	}
	var body apiError
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error != code {
		t.Fatalf("error = %q, want %q", body.Error, code)
	}
}
