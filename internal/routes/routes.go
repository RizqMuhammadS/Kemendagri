package routes

import (
	"github.com/gin-gonic/gin"
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

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "Meeting Minutes AI",
		})
	})

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
			}

			// Dashboard
			protected.GET("/dashboard", meetingCtrl.GetDashboard)
		}
	}

	return router
}