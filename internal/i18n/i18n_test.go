package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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

func TestGeneratedCatalogMatchesLocaleFiles(t *testing.T) {
	for _, language := range []string{English, Chinese} {
		data, err := os.ReadFile(filepath.Join("..", "..", "locales", language+".json"))
		if err != nil {
			t.Fatal(err)
		}
		var want map[string]string
		if err := json.Unmarshal(data, &want); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(dictionaries[language], want) {
			t.Fatalf("%s catalog is stale; run go generate ./internal/i18n", language)
		}
	}
}
