package routes

import (
	"github.com/gin-gonic/gin"

	"chat/handler"
)

func SetupRouter(
	userHandler *handler.UserHandler,
	groupHandler *handler.GroupHandler,
	messageHandler *handler.MessageHandler,
) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// User routes
	userRoutes := r.Group("/users")
	{
		userRoutes.POST("", userHandler.CreateUser)
		userRoutes.GET("", userHandler.GetAllUsers)
		userRoutes.GET("/:id", userHandler.GetUserByID)
	}

	// Group routes
	groupRoutes := r.Group("/groups")
	{
		groupRoutes.POST("", groupHandler.CreateGroup)
		groupRoutes.DELETE("/:id", groupHandler.DeleteGroup)
		groupRoutes.POST("/:id/members", groupHandler.AddMember)
		groupRoutes.DELETE("/:id/members/:user_id", groupHandler.RemoveMember)
		groupRoutes.GET("/:id/members", groupHandler.GetMembers)
	}

	// Message routes
	messageRoutes := r.Group("/messages")
	{
		messageRoutes.POST("", messageHandler.SendMessage)
		messageRoutes.GET("", messageHandler.GetConversation)
		messageRoutes.POST("/:id/seen", messageHandler.MarkSeen)
		messageRoutes.DELETE("/:id", messageHandler.DeleteMessage)
		messageRoutes.GET("/unseen/:user_id", messageHandler.GetUnseenCount)
	}

	return r
}

