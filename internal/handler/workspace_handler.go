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
// POST /api/workspaces
func (h *WorkspaceHandler) CreateWorkspace(c *fiber.Ctx) error {
	// TODO: Get user ID from authentication context
	// For now, using a dummy user ID from header
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

	var req domain.CreateWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	workspace, err := h.service.CreateWorkspace(c.Context(), userID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create workspace",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(workspace)
}

// GetWorkspace handles retrieving a workspace by ID
// GET /api/workspaces/:id
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
// GET /api/workspaces/my
func (h *WorkspaceHandler) GetMyWorkspaces(c *fiber.Ctx) error {
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
// PUT /api/workspaces/:id
func (h *WorkspaceHandler) UpdateWorkspace(c *fiber.Ctx) error {
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

	workspace, err := h.service.UpdateWorkspace(c.Context(), id, userID, &req)
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

	return c.JSON(workspace)
}

// DeleteWorkspace handles deleting a workspace
// DELETE /api/workspaces/:id
func (h *WorkspaceHandler) DeleteWorkspace(c *fiber.Ctx) error {
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
