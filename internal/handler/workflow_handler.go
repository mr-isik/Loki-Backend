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
// @Summary Create a new workflow
// @Description Create a new workflow in a workspace
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace ID (UUID)"
// @Param request body domain.CreateWorkflowRequest true "Workflow details"
// @Success 201 {object} domain.WorkflowResponse
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
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} map[string]interface{} "Returns paginated workflows"
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
// @Summary Archive workflow
// @Description Archive a workflow to make it inactive
// @Tags Workflows
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID (UUID)"
// @Success 200 {object} domain.WorkflowResponse
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
