package aliases

import (
	"strings"
	"unicode"
)

func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxNameLen {
		return ErrInvalidName
	}
	if strings.Contains(name, "\x00") {
		return ErrInvalidName
	}
	for _, r := range name {
		if r > unicode.MaxASCII || (!unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-') {
			return ErrInvalidName
		}
	}
	return nil
}

func validateValue(value string) error {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxValueLen {
		return ErrInvalidValue
	}
	if strings.Contains(value, "\x00") {
		return ErrInvalidValue
	}
	return nil
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
