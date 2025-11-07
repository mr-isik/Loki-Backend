package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mr-isik/loki-backend/internal/handler"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, userHandler *handler.UserHandler, workspaceHandler *handler.WorkspaceHandler, workflowHandler *handler.WorkflowHandler, workflowEdgeHandler *handler.WorkflowEdgeHandler) {
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

	// API routes
	api := app.Group("/api")

	// User routes
	users := api.Group("/users")
	users.Post("/", userHandler.CreateUser)
	users.Get("/:id", userHandler.GetUser)
	users.Patch("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	// Workspace routes
	workspaces := api.Group("/workspaces")
	workspaces.Post("/", workspaceHandler.CreateWorkspace)
	workspaces.Get("/my", workspaceHandler.GetMyWorkspaces)
	workspaces.Get("/:id", workspaceHandler.GetWorkspace)
	workspaces.Put("/:id", workspaceHandler.UpdateWorkspace)
	workspaces.Delete("/:id", workspaceHandler.DeleteWorkspace)

	// Workspace workflows routes (nested)
	workspaces.Get("/:workspace_id/workflows", workflowHandler.GetWorkspaceWorkflows)
	workspaces.Post("/:workspace_id/workflows", workflowHandler.CreateWorkflow)

	// Workflow routes
	workflows := api.Group("/workflows")
	workflows.Get("/:id", workflowHandler.GetWorkflow)
	workflows.Put("/:id", workflowHandler.UpdateWorkflow)
	workflows.Delete("/:id", workflowHandler.DeleteWorkflow)
	workflows.Post("/:id/publish", workflowHandler.PublishWorkflow)
	workflows.Post("/:id/archive", workflowHandler.ArchiveWorkflow)
	workflows.Get("/:workflow_id/edges", workflowEdgeHandler.GetWorkflowEdgesByWorkflow)

	// Workflow Edge routes
	edges := api.Group("/workflow-edges")
	edges.Post("/", workflowEdgeHandler.CreateWorkflowEdge)
	edges.Get("/:id", workflowEdgeHandler.GetWorkflowEdge)
	edges.Put("/:id", workflowEdgeHandler.UpdateWorkflowEdge)
	edges.Delete("/:id", workflowEdgeHandler.DeleteWorkflowEdge)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "Route not found",
		})
	})
}
