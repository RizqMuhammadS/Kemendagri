package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yourusername/meeting-minutes-ai/internal/config"
	"github.com/yourusername/meeting-minutes-ai/internal/models"
)

// LLMService handles AI summarization using LLM
type LLMService struct {
	cfg *config.Config
}

// LLMRequest represents the request body for OpenAI API
type LLMRequest struct {
	Model    string        `json:"model"`
	Messages []LLMMessage  `json:"messages"`
}

// LLMMessage represents a message in the conversation
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse represents the response from OpenAI API
type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// SummarizeResult holds the structured output from LLM summarization
type SummarizeResult struct {
	Summary         string               `json:"summary"`
	DiscussionPoints []string             `json:"discussion_points"`
	Decisions       []string             `json:"decisions"`
	ActionItems     []ActionItemResult   `json:"action_items"`
}

// ActionItemResult holds action item data from LLM
type ActionItemResult struct {
	Task     string `json:"task"`
	Assignee string `json:"assignee"`
	Deadline string `json:"deadline"`
}

// NewLLMService creates a new LLM service
func NewLLMService(cfg *config.Config) *LLMService {
	return &LLMService{cfg: cfg}
}

// GenerateMinutes generates structured meeting minutes from cleaned transcript
func (s *LLMService) GenerateMinutes(meetingTitle string, participants []models.Participant, cleanedText string) (*SummarizeResult, error) {
	// If no API key configured, use local text extraction fallback
	if s.cfg.LLMApiKey == "" {
		return s.localGenerate(meetingTitle, participants, cleanedText), nil
	}

	prompt := s.buildPrompt(meetingTitle, participants, cleanedText)

	resp, err := s.callLLM(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	result, err := s.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return result, nil
}

func (s *LLMService) buildPrompt(title string, participants []models.Participant, text string) string {
	participantNames := ""
	for i, p := range participants {
		if i > 0 {
			participantNames += ", "
		}
		participantNames += p.Name
	}
	if participantNames == "" {
		participantNames = "Tidak disebutkan"
	}

	return fmt.Sprintf(`Anda adalah asisten notulensi rapat profesional. Buatlah notulensi rapat yang terstruktur dari transkrip berikut.

Judul Rapat: %s
Peserta: %s

Transkrip Rapat:
%s

Berdasarkan transkrip di atas, buatlah:

1. RINGKASAN RAPAT (paragraf singkat yang mencakup esensi rapat)
2. POIN PEMBAHASAN (daftar poin-poin yang dibahas)
3. KEPUTUSAN (daftar keputusan yang diambil)
4. ACTION ITEMS (daftar tugas yang harus dilakukan, beserta penanggung jawab dan deadline jika disebutkan)

Format output dalam JSON:
{
  "summary": "ringkasan rapat",
  "discussion_points": ["poin 1", "poin 2", ...],
  "decisions": ["keputusan 1", "keputusan 2", ...],
  "action_items": [
    {"task": "deskripsi tugas", "assignee": "penanggung jawab", "deadline": "deadline jika ada"}
  ]
}`, title, participantNames, text)
}

func (s *LLMService) callLLM(prompt string) (string, error) {
	requestBody := LLMRequest{
		Model: s.cfg.LLMModel,
		Messages: []LLMMessage{
			{
				Role:    "system",
				Content: "Anda adalah asisten notulensi yang ahli. Selalu merespon dengan format JSON yang valid.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.cfg.LLMApiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.LLMApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var llmResp LLMResponse
	if err := json.Unmarshal(body, &llmResp); err != nil {
		return "", fmt.Errorf("failed to decode LLM response: %w", err)
	}

	if llmResp.Error != nil {
		return "", fmt.Errorf("LLM API error: %s", llmResp.Error.Message)
	}

	if len(llmResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in LLM response")
	}

	return llmResp.Choices[0].Message.Content, nil
}

func (s *LLMService) parseResponse(response string) (*SummarizeResult, error) {
	// Try to parse the JSON response directly
	var result SummarizeResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If direct parsing fails, try to extract JSON from the response
		extracted := extractJSON(response)
		if extracted == "" {
			return nil, fmt.Errorf("failed to parse LLM response as JSON")
		}
		if err := json.Unmarshal([]byte(extracted), &result); err != nil {
			return nil, fmt.Errorf("failed to parse extracted JSON: %w", err)
		}
	}

	return &result, nil
}

// localGenerate provides a fallback summarization without external API
func (s *LLMService) localGenerate(meetingTitle string, participants []models.Participant, cleanedText string) *SummarizeResult {
	// Perform basic text extraction: use first 500 chars as summary
	summary := cleanedText
	if len(summary) > 500 {
		summary = summary[:500] + "..."
	}
	if summary == "" {
		summary = "Transkrip rapat tidak tersedia atau kosong."
	}

	// Split text into sentences for discussion points
	sentences := splitSentences(cleanedText)
	discussionPoints := []string{}
	for i, s := range sentences {
		if i >= 10 {
			break
		}
		trimmed := s
		if len(trimmed) > 200 {
			trimmed = trimmed[:200] + "..."
		}
		if trimmed != "" {
			discussionPoints = append(discussionPoints, trimmed)
		}
	}
	if len(discussionPoints) == 0 {
		discussionPoints = []string{"Transkrip tidak tersedia untuk dianalisis"}
	}

	participantNames := ""
	for i, p := range participants {
		if i > 0 {
			participantNames += ", "
		}
		participantNames += p.Name
	}
	if participantNames == "" {
		participantNames = "Tidak disebutkan"
	}

	return &SummarizeResult{
		Summary:          fmt.Sprintf("Ringkasan Rapat \"%s\"\nPeserta: %s\n\n%s", meetingTitle, participantNames, summary),
		DiscussionPoints: discussionPoints,
		Decisions:        []string{"Keputusan tidak dapat diekstrak secara otomatis tanpa AI. Silakan upload transkrip yang lebih lengkap atau konfigurasikan API key LLM."},
		ActionItems:      []ActionItemResult{},
	}
}

// splitSentences splits text into sentence-like chunks
func splitSentences(text string) []string {
	var sentences []string
	current := ""
	for _, ch := range text {
		current += string(ch)
		if ch == '.' || ch == '!' || ch == '?' || ch == '\n' {
			// Trim whitespace and newlines
			trimmed := ""
			for _, c := range current {
				if c != '\n' && c != '\r' {
					trimmed += string(c)
				}
			}
			trimmed = ""
			for _, c := range current {
				if c == '\n' || c == '\r' {
					if trimmed != "" {
						// Avoid adding empty sentences
						sentences = append(sentences, trimmed)
					}
					trimmed = ""
				} else {
					trimmed += string(c)
				}
			}
			if trimmed != "" {
				sentences = append(sentences, trimmed)
			}
			current = ""
		}
	}
	if current != "" {
		trimmed := ""
		for _, c := range current {
			if c != '\n' && c != '\r' {
				trimmed += string(c)
			}
		}
		if trimmed != "" {
			sentences = append(sentences, trimmed)
		}
	}
	// If no sentence delimiters found, split by words
	if len(sentences) == 0 && len(text) > 0 {
		words := []string{}
		word := ""
		for _, ch := range text {
			if ch == ' ' || ch == '\n' || ch == '\r' {
				if word != "" {
					words = append(words, word)
					word = ""
				}
			} else {
				word += string(ch)
			}
		}
		if word != "" {
			words = append(words, word)
		}
		// Group words into ~50 word chunks
		chunkSize := 50
		for i := 0; i < len(words); i += chunkSize {
			end := i + chunkSize
			if end > len(words) {
				end = len(words)
			}
			sentence := ""
			for j := i; j < end; j++ {
				if j > i {
					sentence += " "
				}
				sentence += words[j]
			}
			if sentence != "" {
				sentences = append(sentences, sentence)
			}
		}
	}
	if len(sentences) == 0 {
		sentences = append(sentences, text)
	}
	return sentences
}

// extractJSON attempts to extract a JSON object from text
func extractJSON(text string) string {
	start := -1
	end := -1

	for i := 0; i < len(text); i++ {
		if text[i] == '{' {
			start = i
			break
		}
	}

	if start == -1 {
		return ""
	}

	depth := 0
	for i := start; i < len(text); i++ {
		if text[i] == '{' {
			depth++
		} else if text[i] == '}' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}

	if end == -1 {
		return ""
	}

	return text[start : end+1]
}