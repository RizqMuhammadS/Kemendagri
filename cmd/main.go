package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/yourusername/meeting-minutes-ai/docs"
	"github.com/yourusername/meeting-minutes-ai/internal/config"
	"github.com/yourusername/meeting-minutes-ai/internal/controllers"
	"github.com/yourusername/meeting-minutes-ai/internal/middleware"
	"github.com/yourusername/meeting-minutes-ai/internal/models"
	"github.com/yourusername/meeting-minutes-ai/internal/repositories"
	"github.com/yourusername/meeting-minutes-ai/internal/routes"
	"github.com/yourusername/meeting-minutes-ai/internal/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Meeting Minutes AI API
// @version 1.0
// @description API untuk sistem notulensi rapat otomatis dengan AI
// @termsOfService http://swagger.io/terms/

// @contact.name Tim Developer
// @contact.email developer@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg := config.Load()

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Ensure export directory exists
	if err := os.MkdirAll(cfg.ExportDir, 0755); err != nil {
		log.Fatalf("Failed to create export directory: %v", err)
	}

	// Connect to PostgreSQL database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to PostgreSQL database")

	// Auto-migrate database schemas
	if err := db.AutoMigrate(
		&models.User{},
		&models.Meeting{},
		&models.Participant{},
		&models.DiscussionPoint{},
		&models.Decision{},
		&models.ActionItem{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed")

	// Initialize repositories (using PostgreSQL with GORM)
	userRepo := repositories.NewUserRepository(db)
	meetingRepo := repositories.NewMeetingRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	sttService := services.NewSTTService(cfg)
	cleanerService := services.NewTextCleanerService()
	llmService := services.NewLLMService(cfg)
	exportService := services.NewExportService(cfg)
	emailService := services.NewEmailService(cfg)

	meetingService := services.NewMeetingService(
		meetingRepo,
		userRepo,
		sttService,
		cleanerService,
		llmService,
		exportService,
		emailService,
		cfg,
	)

	// Initialize controllers
	authController := controllers.NewAuthController(authService)
	meetingController := controllers.NewMeetingController(meetingService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Setup router
	router := routes.SetupRouter(authController, meetingController, authMiddleware)

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}