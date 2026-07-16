package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/yourusername/meeting-minutes-ai/internal/config"
)

// STTService handles Speech-to-Text conversion
type STTService struct {
	cfg *config.Config
}

// WhisperResponse represents OpenAI Whisper API response
type WhisperResponse struct {
	Text string `json:"text"`
}

// NewSTTService creates a new STT service
func NewSTTService(cfg *config.Config) *STTService {
	return &STTService{cfg: cfg}
}

// Transcribe converts audio file to text using the configured STT engine
func (s *STTService) Transcribe(audioPath string) (string, error) {
	switch s.cfg.STTEngine {
	case "whisper":
		return s.transcribeWhisper(audioPath)
	case "google":
		return s.transcribeGoogle(audioPath)
	case "azure":
		return s.transcribeAzure(audioPath)
	default:
		return "", fmt.Errorf("unsupported STT engine: %s", s.cfg.STTEngine)
	}
}

func (s *STTService) transcribeWhisper(audioPath string) (string, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add model parameter
	_ = writer.WriteField("model", "whisper-1")

	// Add audio file
	part, err := writer.CreateFormFile("file", audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Use STT API key, fallback to LLM API key if not set
	apiKey := s.cfg.STTApiKey
	if apiKey == "" {
		apiKey = s.cfg.LLMApiKey
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("========== STT RESPONSE ==========")
	fmt.Println("HTTP :", resp.Status)
	fmt.Println(string(body))
	fmt.Println("==================================")

	var whisperResp WhisperResponse

	if err := json.Unmarshal(body, &whisperResp); err != nil {
		return "", err
	}

	return whisperResp.Text, nil
}

	func (s *STTService) transcribeGoogle(audioPath string) (string, error) {
	// Google Speech-to-Text implementation placeholder
	return "", fmt.Errorf("Google STT not yet implemented")
}

	func (s *STTService) transcribeAzure(audioPath string) (string, error) {
	// Azure Speech-to-Text implementation placeholder
	return "", fmt.Errorf("Azure STT not yet implemented")
}