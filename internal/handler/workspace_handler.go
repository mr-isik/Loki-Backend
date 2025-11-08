package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type WorkspaceHandler struct {
	service domain.WorkspaceService
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(service domain.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{
		service: service,
	}
}

// CreateWorkspace handles workspace creation
// @Summary Create a new workspace
// @Description Create a new workspace for the authenticated user
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateWorkspaceRequest true "Workspace details"
// @Success 201 {object} domain.WorkspaceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces [post]
func (h *WorkspaceHandler) CreateWorkspace(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req domain.CreateWorkspaceRequest
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	_, err = h.service.CreateWorkspace(c.Context(), userID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create workspace",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetWorkspace handles retrieving a workspace by ID
// @Summary Get workspace by ID
// @Description Retrieve workspace information by ID
// @Tags Workspaces
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID (UUID)"
// @Success 200 {object} domain.WorkspaceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{id} [get]
func (h *WorkspaceHandler) GetWorkspace(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workspace ID format",
		})
	}

	workspace, err := h.service.GetWorkspace(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workspace not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get workspace",
		})
	}

	return c.JSON(workspace)
}

// GetMyWorkspaces handles retrieving all workspaces for the authenticated user
// @Summary Get my workspaces
// @Description Retrieve all workspaces owned by the authenticated user
// @Tags Workspaces
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Returns array of workspaces"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/my [get]
func (h *WorkspaceHandler) GetMyWorkspaces(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	workspaces, err := h.service.GetUserWorkspaces(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get workspaces",
		})
	}

	return c.JSON(workspaces)
}

// ListWorkspaces handles retrieving all workspaces with pagination
// GET /api/workspaces?page=1&page_size=10
func (h *WorkspaceHandler) ListWorkspaces(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	workspaces, total, err := h.service.ListWorkspaces(c.Context(), page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list workspaces",
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	response := PaginatedResponse{
		Data:       workspaces,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	return c.JSON(response)
}

// UpdateWorkspace handles updating a workspace
// @Summary Update workspace
// @Description Update workspace information (owner only)
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID (UUID)"
// @Param request body domain.UpdateWorkspaceRequest true "Workspace update details"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{id} [put]
func (h *WorkspaceHandler) UpdateWorkspace(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workspace ID format",
		})
	}

	var req domain.UpdateWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	_, err = h.service.UpdateWorkspace(c.Context(), id, userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workspace not found",
			})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You are not the owner of this workspace",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update workspace",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteWorkspace handles deleting a workspace
// @Summary Delete workspace
// @Description Delete a workspace (owner only)
// @Tags Workspaces
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workspaces/{id} [delete]
func (h *WorkspaceHandler) DeleteWorkspace(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workspace ID format",
		})
	}

	if err := h.service.DeleteWorkspace(c.Context(), id, userID); err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workspace not found",
			})
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
				Error:   "forbidden",
				Message: "You are not the owner of this workspace",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete workspace",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
