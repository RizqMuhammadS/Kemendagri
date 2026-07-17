package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/yourusername/meeting-minutes-ai/internal/controllers"
	"github.com/yourusername/meeting-minutes-ai/internal/middleware"
)

// SetupRouter configures all API routes
func SetupRouter(
	authCtrl *controllers.AuthController,
	meetingCtrl *controllers.MeetingController,
	authMiddleware *middleware.AuthMiddleware,
) *gin.Engine {
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("internal/templates/*.html")

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Initialize frontend controller
	frontendCtrl := controllers.NewFrontendController()

	// Frontend routes (SPA - auth handled client-side via localStorage)
	router.GET("/", frontendCtrl.ServeFrontend)
	router.GET("/login", frontendCtrl.LoginPage)
	router.GET("/register", frontendCtrl.RegisterPage)
	router.GET("/dashboard", frontendCtrl.ServeFrontend)
	router.GET("/meetings", frontendCtrl.ServeFrontend)
	router.GET("/meetings/create", frontendCtrl.ServeFrontend)

	// Meeting detail page must be defined before API route to avoid conflict
	router.GET("/meetings/:id", func(c *gin.Context) {
		// Check if it's a numeric ID (for frontend page) or something else
		frontendCtrl.ServeFrontend(c)
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "Meeting Minutes AI",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authCtrl.Register)
			auth.POST("/login", authCtrl.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			// Meeting routes
			meetings := protected.Group("/meetings")
			{
				meetings.POST("", meetingCtrl.CreateMeeting)
				meetings.GET("", meetingCtrl.ListMeetings)
				meetings.GET("/:id", meetingCtrl.GetMeeting)
				meetings.POST("/upload-audio", meetingCtrl.UploadAudio)
				meetings.POST("/process-transcript", meetingCtrl.ProcessTranscript)
				meetings.POST("/export", meetingCtrl.ExportMeeting)
				meetings.POST("/send-email", meetingCtrl.SendEmail)
				meetings.PUT("/:id", meetingCtrl.UpdateMeeting)
				meetings.DELETE("/:id", meetingCtrl.DeleteMeeting)
			}

			// Dashboard
			protected.GET("/dashboard", meetingCtrl.GetDashboard)
		}
	}

	return router
}
