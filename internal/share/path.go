package share

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"
)

var invalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

const managedDirName = ".lanfolder"
const MaxFilenameBytes = 255
const MaxFilenameRunes = 255

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

func cleanFilename(name string) (string, error) {
	name = sanitizeFilename(name)
	if filenameTooLong(name) {
		return "", ErrInvalidFilename
	}
	return name, nil
}

func filenameTooLong(name string) bool {
	return filenameLength(name) > filenameLimit()
}

func filenameLimit() int {
	if runtime.GOOS == "windows" {
		return MaxFilenameRunes
	}
	return MaxFilenameBytes
}

func filenameLength(name string) int {
	if runtime.GOOS == "windows" {
		return len([]rune(name))
	}
	return len([]byte(name))
}

func trimFilename(name string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if runtime.GOOS == "windows" {
		runes := []rune(name)
		if len(runes) <= limit {
			return name
		}
		return string(runes[:limit])
	}
	for len(name) > limit {
		_, size := utf8.DecodeLastRuneInString(name)
		name = name[:len(name)-size]
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
	suffix := fmt.Sprintf(" (%d)%s", index, ext)
	if filenameLength(suffix) >= filenameLimit() {
		suffix = fmt.Sprintf(" (%d)", index)
	}
	if filenameLength(suffix) >= filenameLimit() {
		return trimFilename(suffix, filenameLimit())
	}
	base = trimFilename(base, filenameLimit()-filenameLength(suffix))
	if base == "" {
		base = trimFilename("file", filenameLimit()-filenameLength(suffix))
	}
	return base + suffix
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

func isManagedPath(rel string) bool {
	return rel == managedDirName ||
		strings.HasPrefix(rel, managedDirName+"/")
}

func slashRel(rel string) string {
	return filepath.ToSlash(rel)
}
