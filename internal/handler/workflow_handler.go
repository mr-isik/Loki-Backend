package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type WorkflowHandler struct {
	service domain.WorkflowService
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(service domain.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		service: service,
	}
}

// CreateWorkflow handles workflow creation
// POST /api/workspaces/:workspace_id/workflows
func (h *WorkflowHandler) CreateWorkflow(c *fiber.Ctx) error {
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
	}

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
// GET /api/workflows/:id
func (h *WorkflowHandler) GetWorkflow(c *fiber.Ctx) error {
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
	}

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
// GET /api/workspaces/:workspace_id/workflows?page=1&page_size=10
func (h *WorkflowHandler) GetWorkspaceWorkflows(c *fiber.Ctx) error {
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
	}

	workspaceIDParam := c.Params("workspace_id")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_workspace_id",
			Message: "Invalid workspace ID format",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

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

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	response := PaginatedResponse{
		Data:       workflows,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	return c.JSON(response)
}

// UpdateWorkflow handles updating a workflow
// PUT /api/workflows/:id
func (h *WorkflowHandler) UpdateWorkflow(c *fiber.Ctx) error {
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
	}

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
// DELETE /api/workflows/:id
func (h *WorkflowHandler) DeleteWorkflow(c *fiber.Ctx) error {
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
	}

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
// POST /api/workflows/:id/publish
func (h *WorkflowHandler) PublishWorkflow(c *fiber.Ctx) error {
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
	}

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow ID format",
		})
	}

	workflow, err := h.service.PublishWorkflow(c.Context(), id, userID)
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

	return c.JSON(workflow)
}

// ArchiveWorkflow handles archiving a workflow
// POST /api/workflows/:id/archive
func (h *WorkflowHandler) ArchiveWorkflow(c *fiber.Ctx) error {
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
	}

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow ID format",
		})
	}

	workflow, err := h.service.ArchiveWorkflow(c.Context(), id, userID)
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

	return c.JSON(workflow)
}
