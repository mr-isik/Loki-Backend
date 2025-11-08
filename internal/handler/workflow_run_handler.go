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
// @Summary Start workflow run
// @Description Start a new execution of a workflow
// @Tags Workflow Runs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workflow_id path string true "Workflow ID (UUID)"
// @Param request body domain.CreateWorkflowRunRequest true "Run configuration"
// @Success 201 {object} domain.WorkflowRunResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{workflow_id}/runs [post]
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
// @Summary Get workflow run by ID
// @Description Retrieve workflow run information by ID
// @Tags Workflow Runs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow Run ID (UUID)"
// @Success 200 {object} domain.WorkflowRunResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-runs/{id} [get]
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
// @Summary List workflow runs
// @Description Retrieve all runs for a specific workflow with pagination
// @Tags Workflow Runs
// @Produce json
// @Security BearerAuth
// @Param workflow_id path string true "Workflow ID (UUID)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "Returns runs array and total count"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{workflow_id}/runs [get]
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
// @Summary Update workflow run status
// @Description Update the status of a workflow run
// @Tags Workflow Runs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow Run ID (UUID)"
// @Param request body domain.UpdateWorkflowRunStatusRequest true "Status update"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-runs/{id}/status [patch]
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
