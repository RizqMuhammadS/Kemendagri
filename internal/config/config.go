package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    int
	ServerHost    string
	DatabasePath  string
	UploadDir     string
	MaxUploadSize int64

	LLMApiKey string
	LLMModel  string
	LLMApiUrl string

	STTEngine string
	STTApiKey string
	STTApiUrl string

	JWTSecret     string
	JWTExpiration time.Duration

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string

	ExportDir string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:    getEnvInt("SERVER_PORT", 8080),
		ServerHost:    getEnv("SERVER_HOST", "0.0.0.0"),
		DatabasePath:  getEnv("DB_PATH", "./meeting-minutes.db"),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSize: int64(getEnvInt("MAX_UPLOAD_SIZE", 50)) * 1024 * 1024,

		LLMApiKey: getEnv("LLM_API_KEY", ""),
		LLMModel:  getEnv("LLM_MODEL", "gpt-4"),
		LLMApiUrl: getEnv("LLM_API_URL", "https://api.openai.com/v1/chat/completions"),

		STTEngine: getEnv("STT_ENGINE", "whisper"),
		STTApiKey: getEnv("STT_API_KEY", ""),
		STTApiUrl: getEnv("STT_API_URL", ""),

		JWTSecret: getEnv("JWT_SECRET", "default-secret"),
		JWTExpiration: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),

		SMTPHost: getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort: getEnvInt("SMTP_PORT", 587),
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),

		ExportDir: getEnv("EXPORT_DIR", "./exports"),
	}

	return cfg
}

func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if dur, err := time.ParseDuration(value); err == nil {
			return dur
		}
	}
	return defaultValue
}