package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	English = "en"
	Chinese = "zh-CN"
)

//go:embed *.json
var localeFiles embed.FS

var dictionaries = loadDictionaries()

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

func loadDictionaries() map[string]map[string]string {
	out := map[string]map[string]string{}
	for _, language := range []string{English, Chinese} {
		data, err := localeFiles.ReadFile(language + ".json")
		if err != nil {
			out[language] = map[string]string{}
			continue
		}
		var dict map[string]string
		if err := json.Unmarshal(data, &dict); err != nil {
			out[language] = map[string]string{}
			continue
		}
		out[language] = dict
	}
	return out
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
