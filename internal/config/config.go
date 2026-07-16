package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    int
	ServerHost    string
	DBDriver      string
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
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
	// Try to load .env from current working directory first
	err := godotenv.Load()
	if err != nil {
		// If that fails, try loading from the executable's directory
		exePath, exeErr := os.Executable()
		if exeErr == nil {
			exeDir := filepath.Dir(exePath)
			envPath := filepath.Join(exeDir, ".env")
			log.Printf("Trying to load .env from: %s", envPath)
			err = godotenv.Load(envPath)
		}
	}
	if err != nil {
		log.Printf("Warning: .env file not loaded: %v", err)
	}

	cfg := &Config{
		ServerPort:    getEnvInt("SERVER_PORT", 8080),
		ServerHost:    getEnv("SERVER_HOST", "0.0.0.0"),
		DBDriver:      getEnv("DB_DRIVER", "postgres"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnvInt("DB_PORT", 5432),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", "kemendagri"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
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

	fmt.Println("======================================")
	fmt.Println("LLM_API_KEY dari env:", os.Getenv("LLM_API_KEY"))

	if len(cfg.LLMApiKey) > 20 {
		fmt.Println("Config LLM Key Prefix :", cfg.LLMApiKey[:20])
	} else {
		fmt.Println("Config LLM Key :", cfg.LLMApiKey)
	}

	fmt.Println("LLM Model :", cfg.LLMModel)
	fmt.Println("LLM URL   :", cfg.LLMApiUrl)
	fmt.Println("======================================")

	// Debug: log LLM config status (key masked for security)
	apiKeyStatus := "not set"
	if cfg.LLMApiKey != "" {
		if len(cfg.LLMApiKey) >= 8 {
			apiKeyStatus = fmt.Sprintf("set (masked: %s...%s)", cfg.LLMApiKey[:4], cfg.LLMApiKey[len(cfg.LLMApiKey)-4:])
		} else {
			apiKeyStatus = "set (too short)"
		}
	}
	log.Printf("LLM API Key: %s", apiKeyStatus)
	log.Printf("LLM Model: %s", cfg.LLMModel)
	log.Printf("LLM API URL: %s", cfg.LLMApiUrl)

	return cfg
}

func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
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