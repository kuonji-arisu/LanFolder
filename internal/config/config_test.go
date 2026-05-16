package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"lanfolder/internal/share"
)

func TestSaveToPathReplacesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"port":1}`), 0600); err != nil {
		t.Fatal(err)
	}

	if err := saveToPath(path, Config{
		Port:           9000,
		Permission:     share.PermissionManage,
		AccessApproval: true,
		AutoShare:      true,
	}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got Config
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Port != 9000 || got.Permission != share.PermissionManage || !got.AccessApproval || !got.AutoShare {
		t.Fatalf("saved config = %#v", got)
	}
	assertNoConfigTemps(t, dir)
}

func TestSaveToPathCleansTempFileWhenReplaceFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}

	err := saveToPath(path, Config{Port: 9000, Permission: share.PermissionReadOnly})
	if err == nil {
		t.Fatal("expected save to fail when target path is a directory")
	}
	assertNoConfigTemps(t, dir)
	if info, statErr := os.Stat(path); statErr != nil || !info.IsDir() {
		t.Fatalf("target directory was not preserved: info=%#v err=%v", info, statErr)
	}
}

func assertNoConfigTemps(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "config-*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary config files remain: %v", matches)
	}
}
