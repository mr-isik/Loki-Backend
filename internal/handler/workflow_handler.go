package handler

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/mr-isik/loki-backend/internal/engine"
)

type WorkflowHandler struct {
	service     domain.WorkflowService
	nodeService domain.WorkflowNodeService
	edgeService domain.WorkflowEdgeService
	runService  domain.WorkflowRunService
	logRepo     domain.NodeRunLogRepository
	runRepo     domain.WorkflowRunRepository
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(
	service domain.WorkflowService,
	nodeService domain.WorkflowNodeService,
	edgeService domain.WorkflowEdgeService,
	runService domain.WorkflowRunService,
	logRepo domain.NodeRunLogRepository,
	runRepo domain.WorkflowRunRepository,
) *WorkflowHandler {
	return &WorkflowHandler{
		service:     service,
		nodeService: nodeService,
		edgeService: edgeService,
		runService:  runService,
		logRepo:     logRepo,
		runRepo:     runRepo,
	}
}

// CreateWorkflow handles workflow creation
// @Summary Create a new workflow
// @Description Create a new workflow in a workspace
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace ID (UUID)"
// @Param request body domain.CreateWorkflowRequest true "Workflow details"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{workspace_id}/workflows [post]
func (h *WorkflowHandler) CreateWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	workspaceIDParam := c.Params("workspace_id")
	workspaceID, err := uuid.Parse(workspaceIDParam)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_workspace_id",
			Message: "Invalid workspace ID format",
		})
	}

	var req domain.CreateWorkflowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	_, err = h.service.CreateWorkflow(c.Context(), workspaceID, userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You are not the owner of this workspace",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create workflow",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetWorkflow handles retrieving a workflow by ID
// @Summary Get workflow by ID
// @Description Retrieve workflow information by ID
// @Tags Workflows
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID (UUID)"
// @Success 200 {object} domain.WorkflowResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{id} [get]
func (h *WorkflowHandler) GetWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow ID format",
		})
	}

	workflow, err := h.service.GetWorkflow(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow not found",
			})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this workflow",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get workflow",
		})
	}

	return c.JSON(workflow)
}

// GetWorkspaceWorkflows handles retrieving all workflows in a workspace
// @Summary Get workspace workflows
// @Description Retrieve all workflows in a workspace with pagination
// @Tags Workflows
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace ID (UUID)"
// @Param page query int false "Page number (1-based)" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} domain.PaginatedResponse "Returns paginated workflows"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{workspace_id}/workflows [get]
func (h *WorkflowHandler) GetWorkspaceWorkflows(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	workspaceIDParam := c.Params("workspace_id")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_workspace_id",
			Message: "Invalid workspace ID format",
		})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	workflows, total, err := h.service.GetWorkspaceWorkflows(c.Context(), workspaceID, userID, page, pageSize)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this workspace",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get workflows",
		})
	}

	response := domain.NewPaginatedResponse(workflows, int(total), page, pageSize)
	return c.JSON(response)
}

// UpdateWorkflow handles updating a workflow
// @Summary Update workflow
// @Description Update workflow information
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID (UUID)"
// @Param request body domain.UpdateWorkflowRequest true "Workflow update details"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{id} [put]
func (h *WorkflowHandler) UpdateWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow ID format",
		})
	}

	var req domain.UpdateWorkflowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	_, err = h.service.UpdateWorkflow(c.Context(), id, userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow not found",
			})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this workflow",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update workflow",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteWorkflow handles deleting a workflow
// @Summary Delete workflow
// @Description Delete a workflow (soft delete)
// @Tags Workflows
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{id} [delete]
func (h *WorkflowHandler) DeleteWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow ID format",
		})
	}

	if err := h.service.DeleteWorkflow(c.Context(), id, userID); err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow not found",
			})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this workflow",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete workflow",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// PublishWorkflow handles publishing a workflow
// @Summary Publish workflow
// @Description Publish a workflow to make it active
// @Tags Workflows
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{id}/publish [post]
func (h *WorkflowHandler) PublishWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow ID format",
		})
	}

	err = h.service.PublishWorkflow(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow not found",
			})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this workflow",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to publish workflow",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ArchiveWorkflow handles archiving a workflow
// @Summary Archive workflow
// @Description Archive a workflow to make it inactive
// @Tags Workflows
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{id}/archive [post]
func (h *WorkflowHandler) ArchiveWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow ID format",
		})
	}

	err = h.service.ArchiveWorkflow(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow not found",
			})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this workflow",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to archive workflow",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RunWorkflow handles executing a workflow
// @Summary Run workflow
// @Description Execute a workflow immediately
// @Tags Workflows
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID (UUID)"
// @Success 200 {object} domain.WorkflowRunResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{id}/run [post]
func (h *WorkflowHandler) RunWorkflow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	idParam := c.Params("id")
	workflowID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow ID format",
		})
	}

	// 1. Check access
	_, err = h.service.GetWorkflow(c.Context(), workflowID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow not found",
			})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this workflow",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to check workflow access",
		})
	}

	// 2. Create Run
	runResponse, err := h.runService.StartWorkflowRun(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create workflow run",
		})
	}

	// 3. Fetch Nodes and Edges
	// We need domain structs, but services return Response structs.
	// We might need to use Repositories directly if Services only return Responses,
	// OR map Responses back to Domain models.
	// Looking at the code, Services return *Response.
	// The Engine needs domain.WorkflowNode and domain.WorkflowEdge.
	// Let's see if we can map them or if we should use Repositories.
	// Using Repositories in Handler is generally discouraged if Service layer exists,
	// but for the Engine execution which is internal logic, it might be acceptable.
	// However, `WorkflowHandler` now has `nodeService` and `edgeService`.
	// Let's assume we can map them or the service has a method to get domain models (unlikely based on standard patterns).
	// Actually, `WorkflowEngine` expects `[]domain.WorkflowNode`.
	// The `WorkflowNodeResponse` is very similar to `WorkflowNode`.
	// Let's implement a mapper here or fetch via repository if we had access.
	// Since we injected Services, let's use them and map.

	nodeResponses, err := h.nodeService.GetWorkflowNodesByWorkflowID(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch workflow nodes",
		})
	}

	edgeResponses, err := h.edgeService.GetWorkflowEdgesByWorkflowID(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch workflow edges",
		})
	}

	// Map to Domain Models
	var nodes []domain.WorkflowNode
	for _, nr := range nodeResponses {
		nodes = append(nodes, domain.WorkflowNode{
			ID:         nr.ID,
			WorkflowID: nr.WorkflowID,
			TemplateID: nr.TemplateID,
			PositionX:  nr.PositionX,
			PositionY:  nr.PositionY,
			Data:       nr.Data,
		})
	}

	var edges []domain.WorkflowEdge
	for _, er := range edgeResponses {
		edges = append(edges, domain.WorkflowEdge{
			ID:           er.ID,
			WorkflowID:   er.WorkflowID,
			SourceNodeID: er.SourceNodeID,
			TargetNodeID: er.TargetNodeID,
			SourceHandle: er.SourceHandle,
			TargetHandle: er.TargetHandle,
		})
	}

	// 4. Initialize and Run Engine
	// Note: We are running this synchronously for now as requested/implied.
	// In production, this should likely be a background job (goroutine or worker queue).
	eng := engine.NewWorkflowEngine(
		nodes,
		edges,
		runResponse.ID,
		workflowID,
		h.logRepo,
		h.runRepo,
	)

	// Run in a goroutine to not block the response, OR run sync?
	// "Run'ları oluşturmalı... Node'leri ... çalıştırmalı"
	// If we run sync, the user waits. If async, we return "Running".
	// Let's run Async so the API returns quickly.
	go func() {
		// Create a new context for the background execution
		// because c.Context() will be cancelled when request ends.
		bgCtx := context.Background()
		if err := eng.Execute(bgCtx); err != nil {
			// Log error (we don't have a logger injected here, maybe fmt.Println for now)
			// The engine logs to DB, so we are good.
		}
	}()

	return c.JSON(runResponse)
}
