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
	fileRepo := repository.NewFileRepository(dbConn)

	// Services
	userSvc := service.NewUserService(userRepo)
	groupSvc := service.NewGroupService(groupRepo)
	msgSvc := service.NewMessageService(msgRepo, groupRepo)
	fileSvc := service.NewFileService(fileRepo, "./uploads")

	// Handlers
	userHandler := handler.NewUserHandler(userSvc)
	groupHandler := handler.NewGroupHandler(groupSvc)
	messageHandler := handler.NewMessageHandler(msgSvc, fileSvc)
	fileHandler := handler.NewFileHandler(fileSvc)

	// Router
	r := routes.SetupRouter(userHandler, groupHandler, messageHandler, fileHandler)

	// Run server
	r.Run(":8080")
}
