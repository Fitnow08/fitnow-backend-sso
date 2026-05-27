package utils

import (
	"strings"
)

func escapeMarkdownV2(text string) string {
	specialChars := "_*[]()~`>#+-=|{}.!"
	escapedText := strings.Builder{}

	for _, char := range text {
		if strings.ContainsRune(specialChars, char) {
			escapedText.WriteRune('\\')
		}
		escapedText.WriteRune(char)
	}

	return escapedText.String()
}
