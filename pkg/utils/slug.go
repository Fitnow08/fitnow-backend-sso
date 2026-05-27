package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

func Slugify(s string) (string, error) {
	str := strings.ToLower(s)

	var b strings.Builder
	prevWasDash := false
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			prevWasDash = false
		} else {
			if !prevWasDash {
				b.WriteRune('-')
				prevWasDash = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")

	// Удаляем повторяющиеся дефисы.
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	if slug == "" {
		return "", errors.New("строка пуста")
	}
	return slug, nil
}
