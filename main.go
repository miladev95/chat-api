package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal %v — shutting down gracefully...", sig)

	// Give active requests up to 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close database connection
	sqlDB, err := db.GetDB().DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}

	log.Println("Server exited gracefully")
}
