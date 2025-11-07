package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type NodeRunLogHandler struct {
	service domain.NodeRunLogService
}

func NewNodeRunLogHandler(service domain.NodeRunLogService) *NodeRunLogHandler {
	return &NodeRunLogHandler{
		service: service,
	}
}

// CreateNodeRunLog handles creating a new node run log
// POST /api/node-run-logs
func (h *NodeRunLogHandler) CreateNodeRunLog(c *fiber.Ctx) error {
	var req domain.CreateNodeRunLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	log, err := h.service.CreateNodeRunLog(c.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrForeignKeyViolation) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "invalid_reference",
				Message: "Invalid run_id or node_id",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create node run log",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(log)
}

// GetNodeRunLog handles retrieving a node run log by ID
// GET /api/node-run-logs/:id
func (h *NodeRunLogHandler) GetNodeRunLog(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid node run log ID",
		})
	}

	log, err := h.service.GetNodeRunLog(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNodeRunLogNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Node run log not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve node run log",
		})
	}

	return c.JSON(log)
}

// GetNodeRunLogsByRunID handles retrieving node run logs for a workflow run
// GET /api/workflow-runs/:run_id/logs
func (h *NodeRunLogHandler) GetNodeRunLogsByRunID(c *fiber.Ctx) error {
	runIDParam := c.Params("run_id")
	runID, err := uuid.Parse(runIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_run_id",
			Message: "Invalid workflow run ID",
		})
	}

	logs, err := h.service.GetNodeRunLogsByRunID(c.Context(), runID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve node run logs",
		})
	}

	return c.JSON(fiber.Map{
		"logs":  logs,
		"count": len(logs),
	})
}

// UpdateNodeRunLog handles updating a node run log
// PATCH /api/node-run-logs/:id
func (h *NodeRunLogHandler) UpdateNodeRunLog(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid node run log ID",
		})
	}

	var req domain.UpdateNodeRunLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.service.UpdateNodeRunLog(c.Context(), id, &req); err != nil {
		if errors.Is(err, domain.ErrNodeRunLogNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Node run log not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update node run log",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
