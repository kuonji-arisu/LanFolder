package share

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var invalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

func cleanRel(input string) (string, error) {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\\", "/")
	if input == "" || input == "." || input == "/" {
		return "", nil
	}
	if strings.Contains(input, "\x00") {
		return "", ErrInvalidPath
	}
	if path.IsAbs(input) || filepath.IsAbs(input) || strings.Contains(input, ":") {
		return "", ErrInvalidPath
	}
	for _, part := range strings.Split(input, "/") {
		if part == ".." {
			return "", ErrInvalidPath
		}
	}
	cleaned := path.Clean("/" + input)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "." {
		return "", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", ErrInvalidPath
	}
	return cleaned, nil
}

func withinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func containsHiddenSegment(rel string) bool {
	for _, part := range strings.Split(rel, "/") {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(filepath.Base(strings.ReplaceAll(name, "\\", "/")))
	name = invalidFilenameChars.ReplaceAllString(name, "_")
	name = strings.Trim(name, ". ")
	if name == "" {
		name = "file"
	}
	if runtime.GOOS == "windows" {
		reserved := map[string]bool{
			"CON": true, "PRN": true, "AUX": true, "NUL": true,
			"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
			"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		if reserved[strings.ToUpper(base)] {
			name = "_" + name
		}
	}
	return name
}

func uniquePath(dir, name string) string {
	candidate := filepath.Join(dir, name)
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate
	}
	for i := 1; ; i++ {
		next := filepath.Join(dir, uniqueName(name, i))
		if _, err := os.Stat(next); os.IsNotExist(err) {
			return next
		}
	}
}

func uniqueName(name string, index int) string {
	if index <= 0 {
		return name
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s (%d)%s", base, index, ext)
}

func createUniqueFile(dir, name string, mode os.FileMode) (*os.File, string, error) {
	for i := 0; ; i++ {
		candidateName := uniqueName(name, i)
		file, err := os.OpenFile(filepath.Join(dir, candidateName), os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
		if os.IsExist(err) {
			continue
		}
		return file, candidateName, err
	}
}

func createUniqueDir(dir, name string, mode os.FileMode) (string, error) {
	for i := 0; ; i++ {
		candidateName := uniqueName(name, i)
		err := os.Mkdir(filepath.Join(dir, candidateName), mode)
		if os.IsExist(err) {
			continue
		}
		return candidateName, err
	}
}

func isTrashPath(rel string) bool {
	return rel == ".trash" || strings.HasPrefix(rel, ".trash/")
}

func slashRel(rel string) string {
	return filepath.ToSlash(rel)
}
