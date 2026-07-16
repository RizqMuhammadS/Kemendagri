package services

import (
	"strings"
	"regexp"
)

// TextCleanerService handles cleaning and preprocessing transcribed text
type TextCleanerService struct {
	fillerWords []string
}

// NewTextCleanerService creates a new text cleaner service
func NewTextCleanerService() *TextCleanerService {
	return &TextCleanerService{
		fillerWords: []string{
			"eh", "anu", "ah", "oh", "nah",
			"um", "uh", "er", "ah",
			"a", "hm", "hmm", "mmm",
			"you know", "i mean", "you see",
			"actually", "basically", "literally",
			"like", "so", "well",
			"see", "anyway",
		},
	}
}

// Clean performs text cleaning operations
func (s *TextCleanerService) Clean(text string) string {
	if text == "" {
		return ""
	}

	// 1. Remove filler words (case insensitive)
	text = s.removeFillerWords(text)

	// 2. Remove repeated punctuation
	text = s.removeRepeatedPunctuation(text)

	// 3. Normalize whitespace
	text = s.normalizeWhitespace(text)

	// 4. Remove stuttering (e.g., "I I think" -> "I think")
	text = s.removeStuttering(text)

	// 5. Trim spaces
	text = strings.TrimSpace(text)

	return text
}

// CleanIndonesian provides cleaning tailored for Indonesian language filler words
func (s *TextCleanerService) CleanIndonesian(text string) string {
	indonesianFillers := []string{
		"eh", "anu", "ah", "oh", "nah", "loh", "dong", "sih", "kok",
		"gitu", "gituloh", "anu", "ehm",
	}

	s.fillerWords = append(s.fillerWords, indonesianFillers...)
	return s.Clean(text)
}

func (s *TextCleanerService) removeFillerWords(text string) string {
	// Build regex pattern for filler words (word boundary aware)
	pattern := `\b(` + strings.Join(s.fillerWords, `|`) + `)\b`
	re := regexp.MustCompile(`(?i)` + pattern)
	return re.ReplaceAllString(text, "")
}

func (s *TextCleanerService) removeRepeatedPunctuation(text string) string {
	// Replace multiple periods with single period
	re := regexp.MustCompile(`\.{3,}`)
	text = re.ReplaceAllString(text, ".")

	// Replace multiple commas
	re = regexp.MustCompile(`,{2,}`)
	text = re.ReplaceAllString(text, ",")

	// Replace multiple spaces around punctuation
	re = regexp.MustCompile(`\s+([.,!?;:])`)
	text = re.ReplaceAllString(text, "$1")

	return text
}

func (s *TextCleanerService) normalizeWhitespace(text string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// Replace multiple newlines with single newline
	re = regexp.MustCompile(`\n{3,}`)
	text = re.ReplaceAllString(text, "\n\n")

	return text
}

func (s *TextCleanerService) removeStuttering(text string) string {
	// Remove repeated words (e.g., "I I think" -> "I think")
	// Note: Go RE2 does not support backreferences, so we do this manually
	words := strings.Fields(text)
	if len(words) <= 1 {
		return text
	}
	result := []string{words[0]}
	for i := 1; i < len(words); i++ {
		if strings.ToLower(words[i]) != strings.ToLower(words[i-1]) {
			result = append(result, words[i])
		}
	}
	return strings.Join(result, " ")
}
