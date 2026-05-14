package share

import (
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestListRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	m := NewManager()
	if err := m.Configure(root, PermissionReadOnly, false); err != nil {
		t.Fatal(err)
	}
	if _, err := m.List("../"); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
	if _, err := m.List(filepath.VolumeName(root) + `\Windows`); err == nil {
		t.Fatal("expected absolute path to be rejected")
	}
}

func TestDeleteMovesToTrash(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "note.txt")
	if err := os.WriteFile(target, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewManager()
	if err := m.Configure(root, PermissionManage, false); err != nil {
		t.Fatal(err)
	}
	if err := m.Delete("note.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected original file to be moved, got %v", err)
	}
	trashEntries, err := os.ReadDir(filepath.Join(root, ".lanfolder", "trash"))
	if err != nil {
		t.Fatal(err)
	}
	if len(trashEntries) != 1 {
		t.Fatalf("expected one trash entry, got %d", len(trashEntries))
	}
}

func TestDeleteSymlinkMovesLinkAndKeepsTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	link := filepath.Join(root, "link.txt")
	if err := os.WriteFile(target, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink not available: %v", err)
	}

	m := NewManager()
	if err := m.Configure(root, PermissionManage, false); err != nil {
		t.Fatal(err)
	}
	if err := m.Delete("link.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected symlink target to remain, got %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("expected symlink to move to trash, got %v", err)
	}
	trashEntries, err := os.ReadDir(filepath.Join(root, ".lanfolder", "trash"))
	if err != nil {
		t.Fatal(err)
	}
	if len(trashEntries) != 1 {
		t.Fatalf("expected one trash entry, got %d", len(trashEntries))
	}
}

func TestDeleteManagedTrashPathIsAlwaysRejected(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".lanfolder", "trash"), 0755); err != nil {
		t.Fatal(err)
	}

	m := NewManager()
	if err := m.Configure(root, PermissionManage, false); err != nil {
		t.Fatal(err)
	}
	if err := m.Delete(".lanfolder/trash"); !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected managed trash deletion to be rejected, got %v", err)
	}
}

func TestDeleteManagedPathsAreAlwaysRejected(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".lanfolder"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".lanfolder", "messages.jsonl"), []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager()
	if err := m.Configure(root, PermissionManage, true); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{".lanfolder", ".lanfolder/messages.jsonl"} {
		if err := m.Delete(target); !errors.Is(err, ErrInvalidPath) {
			t.Fatalf("Delete(%q) error = %v, want ErrInvalidPath", target, err)
		}
	}
}

func TestListAlwaysHidesManagedPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".lanfolder"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".secret"), []byte("hidden"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager()
	if err := m.Configure(root, PermissionReadOnly, true); err != nil {
		t.Fatal(err)
	}
	result, err := m.List("")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range result.Entries {
		if entry.Name == ".lanfolder" {
			t.Fatalf("expected managed path %q to be hidden from listings", entry.Name)
		}
	}
	if len(result.Entries) != 1 || result.Entries[0].Name != ".secret" {
		t.Fatalf("expected non-managed hidden files to remain visible, got %#v", result.Entries)
	}
}

func TestMessagesValidateAndRoundTrip(t *testing.T) {
	root := t.TempDir()
	store := NewMessageStore()
	if _, err := store.Send(root, "client-1", "   "); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("empty message error = %v, want ErrInvalidMessage", err)
	}
	if _, err := store.Send(root, "client-1", strings.Repeat("a", MaxMessageTextChars+1)); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("long message error = %v, want ErrInvalidMessage", err)
	}
	if _, err := store.Send(root, "", "hello"); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("empty client id error = %v, want ErrInvalidMessage", err)
	}
	if _, err := store.Send(root, strings.Repeat("a", maxClientIDChars+1), "hello"); !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("long client id error = %v, want ErrInvalidMessage", err)
	}

	first, err := store.Send(root, "client-1", " hello ")
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Send(root, "client-2", "收到")
	if err != nil {
		t.Fatal(err)
	}
	messages, err := store.List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(messages))
	}
	if messages[0].ID != first.ID || messages[0].Text != "hello" || messages[1].ID != second.ID {
		t.Fatalf("messages were not returned in write order: %#v", messages)
	}
}

func TestClearMessagesRemovesMessageFile(t *testing.T) {
	root := t.TempDir()
	store := NewMessageStore()
	if _, err := store.Send(root, "client-1", "hello"); err != nil {
		t.Fatal(err)
	}
	if err := store.Clear(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".lanfolder", "messages.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("messages file still exists after clear: %v", err)
	}
	messages, err := store.List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 0 {
		t.Fatalf("messages = %#v, want empty", messages)
	}
	if err := store.Clear(root); err != nil {
		t.Fatalf("clear should ignore missing messages file: %v", err)
	}
}

func TestConcurrentMessagesDoNotCorruptJSONL(t *testing.T) {
	root := t.TempDir()
	store := NewMessageStore()
	const count = 12
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Send(root, "client-1", "hello")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	messages, err := store.List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != count {
		t.Fatalf("messages = %d, want %d", len(messages), count)
	}
}

func TestListMessagesSkipsInvalidJSONLLines(t *testing.T) {
	root := t.TempDir()
	store := NewMessageStore()
	first, err := store.Send(root, "client-1", "hello")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(root, ".lanfolder", "messages.jsonl"), os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("{bad json}\n"); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := store.Send(root, "client-1", "again")
	if err != nil {
		t.Fatal(err)
	}

	messages, err := store.List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[0].ID != first.ID || messages[1].ID != second.ID {
		t.Fatalf("messages = %#v, want valid lines only", messages)
	}
}

func TestUploadSanitizesAndAvoidsOverwrite(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bad_name.txt"), []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewManager()
	if err := m.Configure(root, PermissionUpload, false); err != nil {
		t.Fatal(err)
	}
	header := multipartHeader(t, `..\bad:name.txt`, "new")
	entry, err := m.SaveUpload("", header)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Name == "bad_name.txt" {
		t.Fatal("expected upload to avoid overwriting existing file")
	}
	if strings.ContainsAny(entry.Name, `<>:"/\|?*`) {
		t.Fatalf("filename was not sanitized: %q", entry.Name)
	}
}

func TestUploadSanitizesWindowsReservedBasename(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows reserved device filenames are platform-specific")
	}
	root := t.TempDir()
	m := NewManager()
	if err := m.Configure(root, PermissionUpload, false); err != nil {
		t.Fatal(err)
	}
	entry, err := m.SaveUpload("", multipartHeader(t, "CON.txt", "hello"))
	if err != nil {
		t.Fatal(err)
	}
	if entry.Name != "_CON.txt" {
		t.Fatalf("reserved filename = %q, want _CON.txt", entry.Name)
	}
}

func TestConcurrentUploadsUseUniqueNames(t *testing.T) {
	root := t.TempDir()
	m := NewManager()
	if err := m.Configure(root, PermissionUpload, false); err != nil {
		t.Fatal(err)
	}

	const count = 8
	headers := make([]*multipart.FileHeader, count)
	for i := range headers {
		headers[i] = multipartHeader(t, "note.txt", "hello")
	}
	var wg sync.WaitGroup
	names := make(chan string, count)
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(header *multipart.FileHeader) {
			defer wg.Done()
			entry, err := m.SaveUpload("", header)
			if err != nil {
				errs <- err
				return
			}
			names <- entry.Name
		}(headers[i])
	}
	wg.Wait()
	close(names)
	close(errs)

	for err := range errs {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for name := range names {
		if seen[name] {
			t.Fatalf("duplicate upload name %q", name)
		}
		seen[name] = true
	}
	if len(seen) != count {
		t.Fatalf("created files = %d, want %d", len(seen), count)
	}
}

func TestConcurrentMkdirUsesUniqueNames(t *testing.T) {
	root := t.TempDir()
	m := NewManager()
	if err := m.Configure(root, PermissionUpload, false); err != nil {
		t.Fatal(err)
	}

	const count = 8
	var wg sync.WaitGroup
	names := make(chan string, count)
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			entry, err := m.Mkdir("", "docs")
			if err != nil {
				errs <- err
				return
			}
			names <- entry.Name
		}()
	}
	wg.Wait()
	close(names)
	close(errs)

	for err := range errs {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for name := range names {
		if seen[name] {
			t.Fatalf("duplicate folder name %q", name)
		}
		seen[name] = true
	}
	if len(seen) != count {
		t.Fatalf("created folders = %d, want %d", len(seen), count)
	}
}

func TestHiddenFilesRequireShowHidden(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".secret"), []byte("hidden"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager()
	if err := m.Configure(root, PermissionReadOnly, false); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.OpenForDownload(".secret"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected hidden file to be blocked, got %v", err)
	}

	if err := m.Configure(root, PermissionReadOnly, true); err != nil {
		t.Fatal(err)
	}
	file, _, err := m.OpenForDownload(".secret")
	if err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
}

func multipartHeader(t *testing.T, name, content string) *multipart.FileHeader {
	t.Helper()
	var body strings.Builder
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("files", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/", strings.NewReader(body.String()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(1024); err != nil {
		t.Fatal(err)
	}
	return req.MultipartForm.File["files"][0]
}
