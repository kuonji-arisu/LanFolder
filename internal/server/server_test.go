package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"lanfolder/internal/share"
)

func TestStartWhileRunningDoesNotReconfigureManager(t *testing.T) {
	root1 := t.TempDir()
	root2 := t.TempDir()
	staticFS := os.DirFS(root1)

	s := New(staticFS)
	if err := s.Start(Config{
		Host:       "127.0.0.1",
		Port:       freePort(t),
		Root:       root1,
		Permission: share.PermissionReadOnly,
	}); err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.Stop(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	err := s.Start(Config{
		Host:       "127.0.0.1",
		Port:       freePort(t),
		Root:       root2,
		Permission: share.PermissionManage,
	})
	if err == nil || !errors.Is(err, errServerAlreadyRunning) {
		t.Fatalf("expected already running error, got %v", err)
	}

	status := s.manager.Status()
	if filepath.Clean(status.Root) != filepath.Clean(root1) {
		t.Fatalf("manager root changed while server was running: got %q want %q", status.Root, root1)
	}
	if status.Permission != share.PermissionReadOnly {
		t.Fatalf("manager permission changed while server was running: got %q", status.Permission)
	}
}

func TestStartClearsExistingMessages(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, ".lanfolder")
	if err := os.MkdirAll(messageDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(messageDir, "messages.jsonl"), []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	s := New(os.DirFS(root))
	if err := s.Start(Config{
		Host:       "127.0.0.1",
		Port:       freePort(t),
		Root:       root,
		Permission: share.PermissionReadOnly,
	}); err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.Stop(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	if _, err := os.Stat(filepath.Join(messageDir, "messages.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("expected server start to clear messages file, got %v", err)
	}
}

func TestStartDoesNotClearMessagesWhenListenFails(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, ".lanfolder")
	if err := os.MkdirAll(messageDir, 0755); err != nil {
		t.Fatal(err)
	}
	messagePath := filepath.Join(messageDir, "messages.jsonl")
	if err := os.WriteFile(messagePath, []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, portText, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}

	s := New(os.DirFS(root))
	err = s.Start(Config{
		Host:       "127.0.0.1",
		Port:       port,
		Root:       root,
		Permission: share.PermissionReadOnly,
	})
	if err == nil {
		t.Fatal("expected start to fail when port is already in use")
	}
	if _, err := os.Stat(messagePath); err != nil {
		t.Fatalf("messages file should remain when server did not start: %v", err)
	}
}

func TestHTTPListMkdirUploadDeleteFlow(t *testing.T) {
	root := t.TempDir()
	s, ts := testServer(t, root, share.PermissionManage)
	defer ts.Close()

	mkdirBody := bytes.NewBufferString(`{"path":"","name":"docs"}`)
	resp := postJSON(t, ts.URL+"/api/mkdir", mkdirBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("mkdir status = %d", resp.StatusCode)
	}
	resp.Body.Close()

	uploadBody, contentType := multipartBody(t, "files", "note.txt", "hello")
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/upload?path=docs", uploadBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("upload status = %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp, err = http.Get(ts.URL + "/api/list?path=docs")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var listed share.ListResult
	if err := json.NewDecoder(resp.Body).Decode(&listed); err != nil {
		t.Fatal(err)
	}
	if got := len(listed.Entries); got != 1 {
		t.Fatalf("listed entries = %d", got)
	}
	if listed.Entries[0].Name != "note.txt" {
		t.Fatalf("listed entry = %q", listed.Entries[0].Name)
	}

	resp = postJSON(t, ts.URL+"/api/delete", bytes.NewBufferString(`{"path":"docs/note.txt"}`))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete status = %d", resp.StatusCode)
	}
	resp.Body.Close()
	if _, err := os.Stat(filepath.Join(root, "docs", "note.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected uploaded file to move to trash, got %v", err)
	}
	if logs := s.Logs(); len(logs) == 0 {
		t.Fatal("expected access logs")
	}
}

func TestHTTPWriteRequiresPermission(t *testing.T) {
	root := t.TempDir()
	_, ts := testServer(t, root, share.PermissionReadOnly)
	defer ts.Close()

	resp := postJSON(t, ts.URL+"/api/mkdir", bytes.NewBufferString(`{"path":"","name":"docs"}`))
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("mkdir status = %d", resp.StatusCode)
	}
	var body apiError
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if body.Error != "permission_denied" {
		t.Fatalf("permission error body = %#v", body)
	}

	uploadBody, contentType := multipartBody(t, "files", "note.txt", "hello")
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/upload", uploadBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("upload status = %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestHTTPMessagesWorkWithReadOnlyPermission(t *testing.T) {
	root := t.TempDir()
	_, ts := testServer(t, root, share.PermissionReadOnly)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/messages")
	if err != nil {
		t.Fatal(err)
	}
	var empty []share.Message
	if err := json.NewDecoder(resp.Body).Decode(&empty); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || len(empty) != 0 {
		t.Fatalf("initial messages status=%d body=%#v", resp.StatusCode, empty)
	}

	resp = postJSON(t, ts.URL+"/api/messages", bytes.NewBufferString(`{"text":"hello","clientId":"client-1"}`))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("send message status = %d", resp.StatusCode)
	}
	var sent share.Message
	if err := json.NewDecoder(resp.Body).Decode(&sent); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if sent.ClientID != "client-1" || sent.Text != "hello" {
		t.Fatalf("sent message = %#v", sent)
	}
	if _, err := os.Stat(filepath.Join(root, ".lanfolder", "messages.jsonl")); err != nil {
		t.Fatalf("expected messages file to be created: %v", err)
	}

	resp, err = http.Get(ts.URL + "/api/messages")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var messages []share.Message
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || messages[0].ID != sent.ID {
		t.Fatalf("messages = %#v, want sent message", messages)
	}

	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/api/messages", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("clear messages status = %d", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(root, ".lanfolder", "messages.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("expected messages file to be removed: %v", err)
	}
}

func TestHTTPMessageRejectsLargeRequest(t *testing.T) {
	root := t.TempDir()
	_, ts := testServer(t, root, share.PermissionReadOnly)
	defer ts.Close()

	resp := postJSON(t, ts.URL+"/api/messages", bytes.NewBufferString(`{"text":"hello","clientId":"`+strings.Repeat("a", 20<<10)+`"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("large message request status = %d", resp.StatusCode)
	}
}

func TestHTTPRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	_, ts := testServer(t, root, share.PermissionReadOnly)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/list?path=..%2F")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("list traversal status = %d", resp.StatusCode)
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, portText, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	return port
}

func testServer(t *testing.T, root string, permission share.Permission) (*Server, *httptest.Server) {
	t.Helper()
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
	return s, httptest.NewServer(s.logMiddleware(mux))
}

func TestFaviconUsesAppIcon(t *testing.T) {
	root := t.TempDir()
	_, ts := testServer(t, root, share.PermissionReadOnly)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/favicon.ico")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("favicon status = %d", resp.StatusCode)
	}
	if contentType := resp.Header.Get("Content-Type"); contentType != "image/png" {
		t.Fatalf("favicon content type = %q", contentType)
	}
}

func postJSON(t *testing.T, url string, body io.Reader) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func multipartBody(t *testing.T, field, name, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(field, name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}
