package pungi

import "strings"

func firstWord(spaceSeparatedText string) string {
	first := strings.Split(spaceSeparatedText, " ")[0]
	return strings.TrimSpace(first)
}
