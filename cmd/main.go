package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/mr-isik/loki-backend/internal/database"
	"github.com/mr-isik/loki-backend/internal/handler"
	"github.com/mr-isik/loki-backend/internal/repository"
	"github.com/mr-isik/loki-backend/internal/router"
	"github.com/mr-isik/loki-backend/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Error loading .env file, proceeding with system environment variables")
	}

	dbConfig := database.NewConfig(
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "loki"),
		getEnv("DB_PASSWORD", "loki_password"),
		getEnv("DB_NAME", "loki_db"),
	)

	log.Println("🔌 Connecting to database...")
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.RunMigrations(ctx); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	userRepo := repository.NewUserRepository(db.Pool)
	workspaceRepo := repository.NewWorkspaceRepository(db.Pool)
	workflowRepo := repository.NewWorkflowRepository(db.Pool)
	workflowEdgeRepo := repository.NewWorkflowEdgeRepository(db.Pool)
	workflowNodeRepo := repository.NewWorkflowNodeRepository(db.Pool)

	userService := service.NewUserService(userRepo)
	workspaceService := service.NewWorkspaceService(workspaceRepo)
	workflowService := service.NewWorkflowService(workflowRepo, workspaceRepo)
	workflowEdgeService := service.NewWorkflowEdgeService(workflowEdgeRepo)
	workflowNodeService := service.NewWorkflowNodeService(workflowNodeRepo)

	userHandler := handler.NewUserHandler(userService)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
	workflowHandler := handler.NewWorkflowHandler(workflowService)
	workflowEdgeHandler := handler.NewWorkflowEdgeHandler(workflowEdgeService)
	workflowNodeHandler := handler.NewWorkflowNodeHandler(workflowNodeService)

	app := fiber.New(fiber.Config{
		AppName:      "Loki Backend API",
		ServerHeader: "Loki",
		ErrorHandler: customErrorHandler,
	})

	router.SetupRoutes(app, userHandler, workspaceHandler, workflowHandler, workflowEdgeHandler, workflowNodeHandler)

	port := getEnv("PORT", ":3000")
	if port[0] != ':' {
		port = ":" + port
	}

	go func() {
		log.Printf("🚀 Server is running on http://localhost%s", port)
		if err := app.Listen(port); err != nil {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": message,
	})
}