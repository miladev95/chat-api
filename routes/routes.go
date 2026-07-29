package routes

import (
	"time"

	"github.com/gin-gonic/gin"

	"chat/handler"
	"chat/middleware"
)

// Default rate limit constants
const (
	defaultRequests = 100                // max requests per window
	defaultWindow   = time.Minute        // sliding window duration
	uploadRequests  = 10                 // stricter limit for file uploads
)

func SetupRouter(
	userHandler *handler.UserHandler,
	groupHandler *handler.GroupHandler,
	messageHandler *handler.MessageHandler,
	fileHandler *handler.FileHandler,
) *gin.Engine {
	r := gin.Default()

	// Create rate limiter instances
	globalLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Requests: defaultRequests,
		Window:   defaultWindow,
	})
	uploadLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Requests: uploadRequests,
		Window:   defaultWindow,
	})

	// Health check — no rate limit
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// User routes
	userRoutes := r.Group("/users", globalLimiter)
	{
		userRoutes.POST("", userHandler.CreateUser)
		userRoutes.GET("", userHandler.GetAllUsers)
		userRoutes.GET("/:id", userHandler.GetUserByID)
	}

	// Group routes
	groupRoutes := r.Group("/groups", globalLimiter)
	{
		groupRoutes.POST("", groupHandler.CreateGroup)
		groupRoutes.DELETE("/:id", groupHandler.DeleteGroup)
		groupRoutes.POST("/:id/members", groupHandler.AddMember)
		groupRoutes.DELETE("/:id/members/:user_id", groupHandler.RemoveMember)
		groupRoutes.GET("/:id/members", groupHandler.GetMembers)
	}

	// Message routes
	messageRoutes := r.Group("/messages", globalLimiter)
	{
		messageRoutes.POST("", messageHandler.SendMessage)
		messageRoutes.POST("/upload", messageHandler.SendFileMessage)
		messageRoutes.GET("", messageHandler.GetConversation)
		messageRoutes.POST("/:id/seen", messageHandler.MarkSeen)
		messageRoutes.DELETE("/:id", messageHandler.DeleteMessage)
		messageRoutes.GET("/unseen/:user_id", messageHandler.GetUnseenCount)
	}

	// Upload route — stricter limit (10 req/min)
	r.POST("/upload", uploadLimiter, fileHandler.Upload)

	// Serve uploaded files statically
	r.Static("/uploads", "./uploads")

	return r
}

