package server

import "strings"

func attachmentDisposition(filename string) string {
	name := strings.ToValidUTF8(filename, "_")
	return `attachment; filename="` + asciiFilenameFallback(name) + `"; filename*=UTF-8''` + encodeRFC5987Value(name)
}

func asciiFilenameFallback(filename string) string {
	var builder strings.Builder
	lastWasReplacement := false
	for _, r := range filename {
		if r >= 0x20 && r <= 0x7e && r != '"' && r != '\\' && r != '/' {
			builder.WriteRune(r)
			lastWasReplacement = false
			continue
		}
		if !lastWasReplacement {
			builder.WriteByte('_')
			lastWasReplacement = true
		}
	}
	fallback := strings.TrimSpace(builder.String())
	if fallback == "" {
		return "download"
	}
	return fallback
}

func encodeRFC5987Value(value string) string {
	var builder strings.Builder
	const hex = "0123456789ABCDEF"
	for i := 0; i < len(value); i++ {
		c := value[i]
		if isRFC5987AttrChar(c) {
			builder.WriteByte(c)
			continue
		}
		builder.WriteByte('%')
		builder.WriteByte(hex[c>>4])
		builder.WriteByte(hex[c&0x0f])
	}
	return builder.String()
}

func isRFC5987AttrChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '!' ||
		c == '#' ||
		c == '$' ||
		c == '&' ||
		c == '+' ||
		c == '-' ||
		c == '.' ||
		c == '^' ||
		c == '_' ||
		c == '`' ||
		c == '|' ||
		c == '~'
}
