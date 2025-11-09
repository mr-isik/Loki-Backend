package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mr-isik/loki-backend/internal/database"
	"github.com/mr-isik/loki-backend/internal/handler"
	"github.com/mr-isik/loki-backend/internal/repository"
	"github.com/mr-isik/loki-backend/internal/router"
	"github.com/mr-isik/loki-backend/internal/service"
	"github.com/mr-isik/loki-backend/internal/util"

	_ "github.com/mr-isik/loki-backend/docs" // Swagger docs
)

// @title Loki Backend API
// @version 1.0
// @description Workflow automation backend service with authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@loki.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {

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

	jwtManager := util.NewJWTManager(
		getEnv("JWT_ACCESS_SECRET", "your-super-secret-access-key-change-this-in-production"),
		getEnv("JWT_REFRESH_SECRET", "your-super-secret-refresh-key-change-this-in-production"),
		15*time.Minute,
		7*24*time.Hour,
	)

	userRepo := repository.NewUserRepository(db.Pool)
	workspaceRepo := repository.NewWorkspaceRepository(db.Pool)
	workflowRepo := repository.NewWorkflowRepository(db.Pool)
	workflowEdgeRepo := repository.NewWorkflowEdgeRepository(db.Pool)
	workflowNodeRepo := repository.NewWorkflowNodeRepository(db.Pool)
	nodeTemplateRepo := repository.NewNodeTemplateRepository(db.Pool)
	workflowRunRepo := repository.NewWorkflowRunRepository(db.Pool)
	nodeRunLogRepo := repository.NewNodeRunLogRepository(db.Pool)

	authService := service.NewAuthService(userRepo, jwtManager)
	userService := service.NewUserService(userRepo)
	workspaceService := service.NewWorkspaceService(workspaceRepo)
	workflowService := service.NewWorkflowService(workflowRepo, workspaceRepo)
	workflowEdgeService := service.NewWorkflowEdgeService(workflowEdgeRepo)
	workflowNodeService := service.NewWorkflowNodeService(workflowNodeRepo)
	nodeTemplateService := service.NewNodeTemplateService(nodeTemplateRepo)
	workflowRunService := service.NewWorkflowRunService(workflowRunRepo)
	nodeRunLogService := service.NewNodeRunLogService(nodeRunLogRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
	workflowHandler := handler.NewWorkflowHandler(workflowService)
	workflowEdgeHandler := handler.NewWorkflowEdgeHandler(workflowEdgeService)
	workflowNodeHandler := handler.NewWorkflowNodeHandler(workflowNodeService)
	nodeTemplateHandler := handler.NewNodeTemplateHandler(nodeTemplateService)
	workflowRunHandler := handler.NewWorkflowRunHandler(workflowRunService)
	nodeRunLogHandler := handler.NewNodeRunLogHandler(nodeRunLogService)

	app := fiber.New(fiber.Config{
		AppName:      "Loki Backend API",
		ServerHeader: "Loki",
		ErrorHandler: customErrorHandler,
	})

	router.SetupRoutes(app, jwtManager, authHandler, userHandler, workspaceHandler, workflowHandler, workflowEdgeHandler, workflowNodeHandler, nodeTemplateHandler, workflowRunHandler, nodeRunLogHandler)

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
