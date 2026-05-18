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
	"net/url"
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

func TestStopClearsMessages(t *testing.T) {
	root := t.TempDir()
	messagePath := filepath.Join(root, ".lanfolder", "messages.jsonl")

	s := New(os.DirFS(root))
	if err := s.Start(Config{
		Host:       "127.0.0.1",
		Port:       freePort(t),
		Root:       root,
		Permission: share.PermissionReadOnly,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.manager.SendMessage("client-1", "hello"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(messagePath); err != nil {
		t.Fatalf("expected messages file before stop: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Stop(ctx); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(messagePath); !os.IsNotExist(err) {
		t.Fatalf("expected server stop to clear messages file, got %v", err)
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
	trashEntries, err := os.ReadDir(filepath.Join(root, ".lanfolder", "trash"))
	if err != nil {
		t.Fatal(err)
	}
	if len(trashEntries) != 1 {
		t.Fatalf("trash entries = %d, want 1", len(trashEntries))
	}
	if logs := s.Logs(); len(logs) == 0 {
		t.Fatal("expected access logs")
	}
}

func TestAccessLogsSkipPageStaticAssetsAndStatusPolling(t *testing.T) {
	root := t.TempDir()
	s, ts := testServer(t, root, share.PermissionReadOnly)
	defer ts.Close()

	for _, url := range []string{
		ts.URL + "/",
		ts.URL + "/favicon.ico",
		ts.URL + "/api/status",
	} {
		resp, err := http.Get(url)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	logs := s.Logs()
	if len(logs) != 0 {
		t.Fatalf("logs = %#v, want none", logs)
	}
}

func TestAccessLogsUseHumanReadableActions(t *testing.T) {
	root := t.TempDir()
	s, ts := testServer(t, root, share.PermissionManage)
	defer ts.Close()

	resp := postJSON(t, ts.URL+"/api/mkdir", bytes.NewBufferString(`{"path":"","name":"docs"}`))
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("mkdir status = %d", resp.StatusCode)
	}

	resp, err := http.Get(ts.URL + "/api/list?path=docs")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	logs := s.Logs()
	if len(logs) != 2 {
		t.Fatalf("logs = %#v, want two meaningful entries", logs)
	}
	if logs[0].Action != "浏览" || logs[0].Target != "docs" {
		t.Fatalf("list log = %#v", logs[0])
	}
	if logs[0].TargetPath != "docs" {
		t.Fatalf("list target path = %q", logs[0].TargetPath)
	}
	if logs[1].Action != "新建文件夹" || logs[1].Target != "docs" {
		t.Fatalf("mkdir log = %#v", logs[1])
	}
}

func TestAccessLogsIncludeAccessRequestOnly(t *testing.T) {
	root := t.TempDir()
	s, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	defer ts.Close()

	resp := postJSON(t, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	var requested struct {
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&requested); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if requested.ExpiresAt.IsZero() {
		t.Fatal("request expiry should not be zero")
	}
	pending := s.PendingAccessRequests()
	if len(pending) != 1 {
		t.Fatalf("pending = %#v, want one request", pending)
	}
	if err := s.ApproveAccessRequest(pending[0].ID); err != nil {
		t.Fatal(err)
	}

	cookies := resp.Cookies()
	if len(cookies) != 1 || cookies[0].Name != requestCookieName {
		t.Fatalf("request cookies = %#v", cookies)
	}
	pollReq, err := http.NewRequest(http.MethodGet, ts.URL+"/api/access/poll", nil)
	if err != nil {
		t.Fatal(err)
	}
	pollReq.AddCookie(cookies[0])
	pollResp, err := http.DefaultClient.Do(pollReq)
	if err != nil {
		t.Fatal(err)
	}
	pollResp.Body.Close()
	if pollResp.StatusCode != http.StatusOK {
		t.Fatalf("poll status = %d", pollResp.StatusCode)
	}

	logs := s.Logs()
	if len(logs) != 1 {
		t.Fatalf("logs = %#v, want only access request entry", logs)
	}
	if logs[0].Action != "请求访问" {
		t.Fatalf("request log = %#v", logs[0])
	}
}

func TestAccessLogsUseBasenameWithFullTargetPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/download?path=music/live/song.mp3", nil)
	req.RemoteAddr = "10.0.0.2:12345"

	log, ok := newLogEntry(req, http.StatusOK, &accessLogMetadata{})
	if !ok {
		t.Fatal("expected download request to be logged")
	}
	if log.Target != "song.mp3" {
		t.Fatalf("target = %q", log.Target)
	}
	if log.TargetPath != "music/live/song.mp3" {
		t.Fatalf("target path = %q", log.TargetPath)
	}
}

func TestAttachmentDispositionEncodesUTF8Filename(t *testing.T) {
	got := attachmentDisposition("中文-日本語-😀.txt")
	want := `attachment; filename="_-_-_.txt"; filename*=UTF-8''%E4%B8%AD%E6%96%87-%E6%97%A5%E6%9C%AC%E8%AA%9E-%F0%9F%98%80.txt`
	if got != want {
		t.Fatalf("content disposition = %q, want %q", got, want)
	}
}

func TestHTTPDownloadUsesUTF8ContentDisposition(t *testing.T) {
	root := t.TempDir()
	name := "中文-日本語-😀.txt"
	if err := os.WriteFile(filepath.Join(root, name), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	_, ts := testServer(t, root, share.PermissionReadOnly)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/download?path=" + url.QueryEscape(name))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("download status = %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("download body = %q", string(data))
	}
	got := resp.Header.Get("Content-Disposition")
	want := attachmentDisposition(name)
	if got != want {
		t.Fatalf("content disposition = %q, want %q", got, want)
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
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("large message request status = %d", resp.StatusCode)
	}
}

func TestHTTPJSONWriteRejectsLargeRequest(t *testing.T) {
	root := t.TempDir()
	_, ts := testServer(t, root, share.PermissionManage)
	defer ts.Close()

	resp := postJSON(t, ts.URL+"/api/mkdir", bytes.NewBufferString(`{"path":"","name":"`+strings.Repeat("a", 20<<10)+`"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("large mkdir request status = %d", resp.StatusCode)
	}
	var body apiError
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "request_too_large" {
		t.Fatalf("large request error body = %#v", body)
	}
}

func TestHTTPMkdirRejectsOverlongFilename(t *testing.T) {
	root := t.TempDir()
	_, ts := testServer(t, root, share.PermissionManage)
	defer ts.Close()

	resp := postJSON(t, ts.URL+"/api/mkdir", bytes.NewBufferString(`{"path":"","name":"`+strings.Repeat("a", share.MaxFilenameBytes+1)+`"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("overlong mkdir status = %d", resp.StatusCode)
	}
	var body apiError
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "invalid_filename" {
		t.Fatalf("overlong mkdir error body = %#v", body)
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

func TestAccessApprovalProtectsDataRoutes(t *testing.T) {
	root := t.TempDir()
	_, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/list")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized list status = %d", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("static page status = %d", resp.StatusCode)
	}

	resp = postJSON(t, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("access request status = %d", resp.StatusCode)
	}
	var requested map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&requested); err != nil {
		t.Fatal(err)
	}
	if _, ok := requested["id"]; ok {
		t.Fatalf("access request response exposed id: %#v", requested)
	}
	if _, ok := requested["code"]; ok {
		t.Fatalf("access request response exposed code: %#v", requested)
	}
}

func TestAccessRequestDedupesPendingClient(t *testing.T) {
	root := t.TempDir()
	s, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	defer ts.Close()

	first := postJSON(t, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	defer first.Body.Close()
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first request status = %d", first.StatusCode)
	}
	cookies := first.Cookies()
	if len(cookies) != 1 || cookies[0].Name != requestCookieName {
		t.Fatalf("request cookies = %#v", cookies)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "changed-client")
	req.AddCookie(cookies[0])
	second, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Body.Close()
	if second.StatusCode != http.StatusOK {
		t.Fatalf("duplicate request status = %d", second.StatusCode)
	}
	if pending := s.PendingAccessRequests(); len(pending) != 1 {
		t.Fatalf("pending = %#v, want one request", pending)
	} else if pending[0].RequestCount != 2 || pending[0].UserAgent != "changed-client" {
		t.Fatalf("pending metadata = %#v, want count 2 with latest user agent", pending[0])
	}
}

func TestAccessRequestCookieMaxAgeTracksPendingExpiry(t *testing.T) {
	root := t.TempDir()
	s, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	defer ts.Close()

	first := postJSON(t, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	var requested struct {
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if err := json.NewDecoder(first.Body).Decode(&requested); err != nil {
		t.Fatal(err)
	}
	first.Body.Close()
	cookies := first.Cookies()
	if len(cookies) != 1 || cookies[0].Name != requestCookieName {
		t.Fatalf("request cookies = %#v", cookies)
	}
	firstMaxAge := cookies[0].MaxAge
	if firstMaxAge <= 0 || firstMaxAge > int(share.AccessRequestTTL.Seconds()) {
		t.Fatalf("first request cookie max age = %d", firstMaxAge)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookies[0])
	second, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	second.Body.Close()
	nextCookies := second.Cookies()
	if len(nextCookies) != 1 || nextCookies[0].Name != requestCookieName {
		t.Fatalf("duplicate request cookies = %#v", nextCookies)
	}
	if nextCookies[0].MaxAge > firstMaxAge {
		t.Fatalf("duplicate request cookie max age = %d, want at most first %d", nextCookies[0].MaxAge, firstMaxAge)
	}
	pending := s.PendingAccessRequests()
	if len(pending) != 1 {
		t.Fatalf("pending = %#v, want one request", pending)
	}
	if !pending[0].ExpiresAt.Equal(requested.ExpiresAt) {
		t.Fatalf("pending expiry = %s, want original %s", pending[0].ExpiresAt, requested.ExpiresAt)
	}
}

func TestAccessRequestCookieMaxAgeUsesRemainingLifetime(t *testing.T) {
	now := time.Now()
	got, ok := accessRequestCookieMaxAge(now.Add(5*time.Second + 500*time.Millisecond))
	if !ok || got > 5 || got <= 0 {
		t.Fatalf("max age = %d, want positive remaining seconds", got)
	}
	got, ok = accessRequestCookieMaxAge(now.Add(500 * time.Millisecond))
	if !ok || got != 1 {
		t.Fatalf("subsecond max age = %d, want 1", got)
	}
	if got, ok := accessRequestCookieMaxAge(now.Add(-time.Second)); ok || got != 0 {
		t.Fatalf("expired max age = %d ok=%v, want invalid", got, ok)
	}
}

func TestAccessRequestCooldownReturnsTooManyRequests(t *testing.T) {
	root := t.TempDir()
	s, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	defer ts.Close()

	first := postJSON(t, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	first.Body.Close()
	pending := s.PendingAccessRequests()
	if len(pending) != 1 {
		t.Fatalf("pending = %#v, want one request", pending)
	}
	if err := s.DenyAccessRequest(pending[0].ID); err != nil {
		t.Fatal(err)
	}

	resp := postJSON(t, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("cooldown status = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}
	var body apiError
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "access_request_limited" {
		t.Fatalf("cooldown body = %#v", body)
	}
}

func TestAccessPollLimitsInvalidIDs(t *testing.T) {
	root := t.TempDir()
	s, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	s.invalidPolls = newFixedWindowLimiter(time.Minute, 2)
	defer ts.Close()

	for i := 0; i < 2; i++ {
		req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/access/poll", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: requestCookieName, Value: "missing-" + strconv.Itoa(i), Path: requestCookiePath})
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("invalid poll %d status = %d", i, resp.StatusCode)
		}
		resp.Body.Close()
	}

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/access/poll", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: requestCookieName, Value: "missing-overflow", Path: requestCookiePath})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("overflow poll status = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}
	cleared := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == requestCookieName && cookie.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatalf("cookies = %#v, want cleared request cookie", resp.Cookies())
	}
}

func TestAccessPollWithoutRequestCookieDoesNotLog(t *testing.T) {
	root := t.TempDir()
	s, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/access/poll")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("poll status = %d", resp.StatusCode)
	}
	var body struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.State != "expired" {
		t.Fatalf("poll state = %q", body.State)
	}
	if logs := s.Logs(); len(logs) != 0 {
		t.Fatalf("logs = %#v, want none", logs)
	}
}

func TestAccessApprovalApproveSetsCookieAndAllowsAPI(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "note.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	s, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	defer ts.Close()

	requestResp := postJSON(t, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	var requested struct {
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if err := json.NewDecoder(requestResp.Body).Decode(&requested); err != nil {
		t.Fatal(err)
	}
	requestResp.Body.Close()
	requestCookies := requestResp.Cookies()
	if len(requestCookies) != 1 || requestCookies[0].Name != requestCookieName {
		t.Fatalf("request cookies = %#v", requestCookies)
	}
	if requested.ExpiresAt.IsZero() {
		t.Fatal("request expiry should not be zero")
	}
	pending := s.PendingAccessRequests()
	if len(pending) != 1 {
		t.Fatalf("pending = %#v, want one request", pending)
	}
	if err := s.ApproveAccessRequest(pending[0].ID); err != nil {
		t.Fatal(err)
	}

	pollReq, err := http.NewRequest(http.MethodGet, ts.URL+"/api/access/poll", nil)
	if err != nil {
		t.Fatal(err)
	}
	pollReq.AddCookie(requestCookies[0])
	pollResp, err := http.DefaultClient.Do(pollReq)
	if err != nil {
		t.Fatal(err)
	}
	var pollBody struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(pollResp.Body).Decode(&pollBody); err != nil {
		t.Fatal(err)
	}
	pollResp.Body.Close()
	if pollBody.State != "approved" {
		t.Fatalf("poll state = %q", pollBody.State)
	}
	cookies := pollResp.Cookies()
	var sessionCookie *http.Cookie
	var clearedRequest bool
	for _, cookie := range cookies {
		if cookie.Name == sessionCookieName {
			sessionCookie = cookie
		}
		if cookie.Name == requestCookieName && cookie.MaxAge < 0 {
			clearedRequest = true
		}
	}
	if sessionCookie == nil || !clearedRequest {
		t.Fatalf("cookies = %#v, want session and cleared request", cookies)
	}

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(sessionCookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("authorized list status = %d", resp.StatusCode)
	}
	resp.Body.Close()

	s.access.Clear()
	req, err = http.NewRequest(http.MethodGet, ts.URL+"/api/list", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(sessionCookie)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("cleared session status = %d", resp.StatusCode)
	}
}

func TestAccessLogoutClearsStaleSessionCookie(t *testing.T) {
	root := t.TempDir()
	_, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/access/logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "stale-session"})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("logout status = %d", resp.StatusCode)
	}
	cleared := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == sessionCookieName && cookie.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatalf("cookies = %#v, want cleared session cookie", resp.Cookies())
	}
}

func TestAccessApprovalDenyReturnsRejectedPoll(t *testing.T) {
	root := t.TempDir()
	s, ts := testServerWithAccess(t, root, share.PermissionReadOnly)
	defer ts.Close()

	requestResp := postJSON(t, ts.URL+"/api/access/request", bytes.NewBufferString(`{}`))
	requestResp.Body.Close()
	requestCookies := requestResp.Cookies()
	if len(requestCookies) != 1 || requestCookies[0].Name != requestCookieName {
		t.Fatalf("request cookies = %#v", requestCookies)
	}
	pending := s.PendingAccessRequests()
	if len(pending) != 1 {
		t.Fatalf("pending = %#v, want one request", pending)
	}
	if err := s.DenyAccessRequest(pending[0].ID); err != nil {
		t.Fatal(err)
	}

	pollReq, err := http.NewRequest(http.MethodGet, ts.URL+"/api/access/poll", nil)
	if err != nil {
		t.Fatal(err)
	}
	pollReq.AddCookie(requestCookies[0])
	pollResp, err := http.DefaultClient.Do(pollReq)
	if err != nil {
		t.Fatal(err)
	}
	defer pollResp.Body.Close()
	var pollBody struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(pollResp.Body).Decode(&pollBody); err != nil {
		t.Fatal(err)
	}
	if pollBody.State != "denied" {
		t.Fatalf("poll state = %q", pollBody.State)
	}
	cleared := false
	for _, cookie := range pollResp.Cookies() {
		if cookie.Name == requestCookieName && cookie.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatalf("cookies = %#v, want cleared request cookie", pollResp.Cookies())
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
	return testServerConfigured(t, root, permission, false)
}

func testServerWithAccess(t *testing.T, root string, permission share.Permission) (*Server, *httptest.Server) {
	return testServerConfigured(t, root, permission, true)
}

func testServerConfigured(t *testing.T, root string, permission share.Permission, accessApproval bool) (*Server, *httptest.Server) {
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
	s.config = Config{Host: "127.0.0.1", Port: 8899, Root: root, Permission: permission, AccessApproval: accessApproval}
	mux := http.NewServeMux()
	s.routes(mux)
	return s, httptest.NewServer(s.logMiddleware(s.secureMiddleware(s.accessMiddleware(mux))))
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
