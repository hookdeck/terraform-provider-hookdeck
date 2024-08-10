package codegen

import (
	"strings"
	"unicode"
)

func toConfigCase(input string) string {
	if input == "AWS SNS" {
		return "Awssns"
	} else if input == "3dEye" {
		return "3DEye"
	}

	words := splitWords(input)
	var result strings.Builder

	for _, word := range words {
		word = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		result.WriteString(word)
	}

	return result.String()
}

// splitWords splits a string into words based on uppercase letters or spaces
func splitWords(s string) []string {
	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if i > 0 && (unicode.IsUpper(r) && !unicode.IsUpper(rune(s[i-1]))) {
			// Split at uppercase letters following lowercase letters or after spaces
			words = append(words, currentWord.String())
			currentWord.Reset()
		}
		if currentWord.Len() > 0 && i < len(s)-1 && (unicode.IsUpper(r) && unicode.IsLower(rune(s[i+1]))) {
			// Split at uppercase letters where the next letter is lower case
			words = append(words, currentWord.String())
			currentWord.Reset()
		}
		if !unicode.IsSpace(r) {
			currentWord.WriteRune(r)
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}
