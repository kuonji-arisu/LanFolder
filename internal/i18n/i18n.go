package i18n

import (
	"fmt"
	"strings"
)

const (
	English = "en"
	Chinese = "zh-CN"
)

//go:generate go run ../../tools/geni18n

func NormalizeLanguage(language string) string {
	value := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(language), "_", "-"))
	switch {
	case strings.HasPrefix(value, "zh"):
		return Chinese
	case strings.HasPrefix(value, "en"):
		return English
	default:
		return Chinese
	}
}

func ValidLanguage(language string) bool {
	normalized := NormalizeLanguage(language)
	return language == normalized
}

func T(language, key string, params map[string]any) string {
	normalized := NormalizeLanguage(language)
	if value, ok := dictionaries[normalized][key]; ok {
		return interpolate(value, params)
	}
	if normalized != English {
		if value, ok := dictionaries[English][key]; ok {
			return interpolate(value, params)
		}
	}
	return key
}

func interpolate(value string, params map[string]any) string {
	if len(params) == 0 {
		return value
	}
	for key, param := range params {
		value = strings.ReplaceAll(value, "{"+key+"}", fmt.Sprint(param))
	}
	return value
}
