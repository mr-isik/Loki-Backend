package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/mr-isik/loki-backend/internal/handler"
	"github.com/mr-isik/loki-backend/internal/middleware"
	"github.com/mr-isik/loki-backend/internal/util"
	
	_ "github.com/mr-isik/loki-backend/docs"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, jwtManager *util.JWTManager, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, workspaceHandler *handler.WorkspaceHandler, workflowHandler *handler.WorkflowHandler, workflowEdgeHandler *handler.WorkflowEdgeHandler, workflowNodeHandler *handler.WorkflowNodeHandler, nodeTemplateHandler *handler.NodeTemplateHandler, workflowRunHandler *handler.WorkflowRunHandler, nodeRunLogHandler *handler.NodeRunLogHandler) {
	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-User-ID",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "loki-backend",
		})
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.New(swagger.Config{
		Title:        "Loki Backend API",
		DeepLinking:  true,
		DocExpansion: "list",
	}))

	// API routes
	api := app.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/me", middleware.AuthMiddleware(jwtManager), authHandler.GetMe)

	// Create auth middleware
	authMiddleware := middleware.AuthMiddleware(jwtManager)

	// User routes (protected)
	users := api.Group("/users", authMiddleware)
	users.Post("/", userHandler.CreateUser)
	users.Get("/:id", userHandler.GetUser)
	users.Patch("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	// Workspace routes (protected)
	workspaces := api.Group("/workspaces", authMiddleware)
	workspaces.Post("/", workspaceHandler.CreateWorkspace)
	workspaces.Get("/my", workspaceHandler.GetMyWorkspaces)
	workspaces.Get("/:id", workspaceHandler.GetWorkspace)
	workspaces.Put("/:id", workspaceHandler.UpdateWorkspace)
	workspaces.Delete("/:id", workspaceHandler.DeleteWorkspace)

	// Workspace workflows routes (nested, protected)
	workspaces.Get("/:workspace_id/workflows", workflowHandler.GetWorkspaceWorkflows)
	workspaces.Post("/:workspace_id/workflows", workflowHandler.CreateWorkflow)

	// Workflow routes (protected)
	workflows := api.Group("/workflows", authMiddleware)
	workflows.Get("/:id", workflowHandler.GetWorkflow)
	workflows.Put("/:id", workflowHandler.UpdateWorkflow)
	workflows.Delete("/:id", workflowHandler.DeleteWorkflow)
	workflows.Post("/:id/publish", workflowHandler.PublishWorkflow)
	workflows.Post("/:id/archive", workflowHandler.ArchiveWorkflow)
	workflows.Get("/:workflow_id/edges", workflowEdgeHandler.GetWorkflowEdgesByWorkflow)
	workflows.Get("/:workflow_id/nodes", workflowNodeHandler.GetWorkflowNodes)
	workflows.Post("/:workflow_id/runs", workflowRunHandler.StartWorkflowRun)
	workflows.Get("/:workflow_id/runs", workflowRunHandler.ListWorkflowRuns)

	// Workflow Edge routes (protected)
	edges := api.Group("/workflow-edges", authMiddleware)
	edges.Post("/", workflowEdgeHandler.CreateWorkflowEdge)
	edges.Get("/:id", workflowEdgeHandler.GetWorkflowEdge)
	edges.Put("/:id", workflowEdgeHandler.UpdateWorkflowEdge)
	edges.Delete("/:id", workflowEdgeHandler.DeleteWorkflowEdge)

	// Workflow Node routes (protected)
	nodes := api.Group("/workflow-nodes", authMiddleware)
	nodes.Post("/", workflowNodeHandler.CreateWorkflowNode)
	nodes.Get("/:id", workflowNodeHandler.GetWorkflowNode)
	nodes.Put("/:id", workflowNodeHandler.UpdateWorkflowNode)
	nodes.Delete("/:id", workflowNodeHandler.DeleteWorkflowNode)

	// Workflow Run routes (protected)
	workflowRuns := api.Group("/workflow-runs", authMiddleware)
	workflowRuns.Get("/:id", workflowRunHandler.GetWorkflowRun)
	workflowRuns.Patch("/:id/status", workflowRunHandler.UpdateWorkflowRunStatus)
	workflowRuns.Get("/:run_id/logs", nodeRunLogHandler.GetNodeRunLogsByRunID)

	// Node Run Log routes (protected)
	nodeRunLogs := api.Group("/node-run-logs", authMiddleware)
	nodeRunLogs.Post("/", nodeRunLogHandler.CreateNodeRunLog)
	nodeRunLogs.Get("/:id", nodeRunLogHandler.GetNodeRunLog)
	nodeRunLogs.Patch("/:id", nodeRunLogHandler.UpdateNodeRunLog)

	// Node Template routes (protected)
	nodeTemplates := api.Group("/node-templates", authMiddleware)
	nodeTemplates.Get("/", nodeTemplateHandler.ListNodeTemplates)
	nodeTemplates.Get("/:id", nodeTemplateHandler.GetNodeTemplate)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "Route not found",
		})
	})
}
