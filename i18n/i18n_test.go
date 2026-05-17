package i18n

import "testing"

func TestTranslateUsesSelectedLanguage(t *testing.T) {
	if got := T(English, "settings.language.label", nil); got != "Language" {
		t.Fatalf("english label = %q", got)
	}
	if got := T(Chinese, "settings.language.label", nil); got != "界面语言" {
		t.Fatalf("chinese label = %q", got)
	}
}

func TestTranslateInterpolatesParams(t *testing.T) {
	if got := T(English, "file.deleteConfirm", map[string]any{"name": "note.txt"}); got != "Delete note.txt? The file will be moved to .lanfolder/trash." {
		t.Fatalf("interpolated = %q", got)
	}
}

func TestTranslateFallsBackToKey(t *testing.T) {
	if got := T(Chinese, "missing.translation.key", nil); got != "missing.translation.key" {
		t.Fatalf("missing key = %q", got)
	}
}
