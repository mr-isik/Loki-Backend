package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type WorkflowRunHandler struct {
	service domain.WorkflowRunService
}

func NewWorkflowRunHandler(service domain.WorkflowRunService) *WorkflowRunHandler {
	return &WorkflowRunHandler{
		service: service,
	}
}

// StartWorkflowRun handles starting a new workflow run
// POST /api/workflows/:workflow_id/runs
func (h *WorkflowRunHandler) StartWorkflowRun(c *fiber.Ctx) error {
	workflowIDParam := c.Params("workflow_id")
	workflowID, err := uuid.Parse(workflowIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_workflow_id",
			Message: "Invalid workflow ID",
		})
	}

	run, err := h.service.StartWorkflowRun(c.Context(), workflowID)
	if err != nil {
		if errors.Is(err, domain.ErrForeignKeyViolation) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "workflow_not_found",
				Message: "Workflow not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to start workflow run",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(run)
}

// GetWorkflowRun handles retrieving a workflow run by ID
// GET /api/workflow-runs/:id
func (h *WorkflowRunHandler) GetWorkflowRun(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow run ID",
		})
	}

	run, err := h.service.GetWorkflowRun(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowRunNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow run not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve workflow run",
		})
	}

	return c.JSON(run)
}

// ListWorkflowRuns handles retrieving workflow runs for a workflow
// GET /api/workflows/:workflow_id/runs
func (h *WorkflowRunHandler) ListWorkflowRuns(c *fiber.Ctx) error {
	workflowIDParam := c.Params("workflow_id")
	workflowID, err := uuid.Parse(workflowIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_workflow_id",
			Message: "Invalid workflow ID",
		})
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	runs, total, err := h.service.ListWorkflowRuns(c.Context(), workflowID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve workflow runs",
		})
	}

	return c.JSON(fiber.Map{
		"runs":   runs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateWorkflowRunStatus handles updating the status of a workflow run
// PATCH /api/workflow-runs/:id/status
func (h *WorkflowRunHandler) UpdateWorkflowRunStatus(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow run ID",
		})
	}

	var req struct {
		Status domain.WorkflowRunStatus `json:"status" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate status
	validStatuses := map[domain.WorkflowRunStatus]bool{
		domain.WorkflowRunStatusPending:   true,
		domain.WorkflowRunStatusRunning:   true,
		domain.WorkflowRunStatusCompleted: true,
		domain.WorkflowRunStatusFailed:    true,
		domain.WorkflowRunStatusCancelled: true,
	}

	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_status",
			Message: "Invalid workflow run status",
		})
	}

	if err := h.service.UpdateRunStatus(c.Context(), id, req.Status); err != nil {
		if errors.Is(err, domain.ErrWorkflowRunNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow run not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update workflow run status",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
