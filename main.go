package main

import (
	"chat/db"
	"chat/handler"
	"chat/repository"
	"chat/routes"
	"chat/service"
)

func main() {
	// Initialize DB connection
	dbConn := db.InitDB()

	// Repositories
	userRepo := repository.NewUserRepository(dbConn)
	groupRepo := repository.NewGroupRepository(dbConn)
	msgRepo := repository.NewMessageRepository(dbConn)

	// Services
	userSvc := service.NewUserService(userRepo)
	groupSvc := service.NewGroupService(groupRepo)
	msgSvc := service.NewMessageService(msgRepo)

	// Handlers
	userHandler := handler.NewUserHandler(userSvc)
	groupHandler := handler.NewGroupHandler(groupSvc)
	messageHandler := handler.NewMessageHandler(msgSvc)

	// Router
	r := routes.SetupRouter(userHandler, groupHandler, messageHandler)

	// Run server
	r.Run(":8080")
}
