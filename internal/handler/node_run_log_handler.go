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
// @Summary Create node run log
// @Description Create a new execution log for a workflow node
// @Tags Node Run Logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateNodeRunLogRequest true "Node run log details"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /node-run-logs [post]
func (h *NodeRunLogHandler) CreateNodeRunLog(c *fiber.Ctx) error {
	var req domain.CreateNodeRunLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	err := h.service.CreateNodeRunLog(c.Context(), &req)
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

	return c.SendStatus(fiber.StatusNoContent)
}

// GetNodeRunLog handles retrieving a node run log by ID
// @Summary Get node run log by ID
// @Description Retrieve node run log information by ID
// @Tags Node Run Logs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Node Run Log ID (UUID)"
// @Success 200 {object} domain.NodeRunLogResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /node-run-logs/{id} [get]
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
// @Summary Get node run logs by workflow run ID
// @Description Retrieve all node execution logs for a specific workflow run with pagination
// @Tags Node Run Logs
// @Produce json
// @Security BearerAuth
// @Param run_id path string true "Workflow Run ID (UUID)"
// @Param page query int false "Page number (1-based)" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} domain.PaginatedResponse "Returns paginated logs"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-runs/{run_id}/logs [get]
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

	// Pagination için manuel olarak yapalım (service'de değişiklik yapmadan)
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Calculate offset and limit
	total := len(logs)
	offset := (page - 1) * pageSize
	end := offset + pageSize

	// Adjust bounds
	if offset >= total {
		offset = 0
		end = 0
		logs = []*domain.NodeRunLogResponse{}
	} else {
		if end > total {
			end = total
		}
		logs = logs[offset:end]
	}

	response := domain.NewPaginatedResponse(logs, total, page, pageSize)
	return c.JSON(response)
}

// UpdateNodeRunLog handles updating a node run log
// @Summary Update node run log
// @Description Update node run log status, output, or error message
// @Tags Node Run Logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Node Run Log ID (UUID)"
// @Param request body domain.UpdateNodeRunLogRequest true "Log update details"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /node-run-logs/{id} [patch]
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
