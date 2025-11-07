package handler

import (
	"errors"

	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/mr-isik/loki-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WorkflowEdgeHandler struct {
	service *service.WorkflowEdgeService
}

func NewWorkflowEdgeHandler(service *service.WorkflowEdgeService) *WorkflowEdgeHandler {
	return &WorkflowEdgeHandler{service: service}
}

// CreateWorkflowEdge handles creating a new workflow edge
// POST /api/workflow-edges
func (h *WorkflowEdgeHandler) CreateWorkflowEdge(c *fiber.Ctx) error {
	var req domain.CreateWorkflowEdgeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.service.CreateWorkflowEdge(c.Context(), &req); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "invalid_input",
				Message: "Invalid node IDs provided",
			})
		}
		if errors.Is(err, domain.ErrForeignKeyViolation) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "invalid_nodes",
				Message: "Source or target node does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create workflow edge",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetWorkflowEdge handles retrieving a workflow edge by ID
// GET /api/workflow-edges/:id
func (h *WorkflowEdgeHandler) GetWorkflowEdge(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow edge ID",
		})
	}

	edge, err := h.service.GetWorkflowEdge(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowEdgeNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow edge not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get workflow edge",
		})
	}

	return c.JSON(edge.ToResponse())
}

// GetWorkflowEdgesByWorkflow handles retrieving all edges for a workflow
// GET /api/workflows/:workflow_id/edges
func (h *WorkflowEdgeHandler) GetWorkflowEdgesByWorkflow(c *fiber.Ctx) error {
	workflowIDParam := c.Params("workflow_id")
	workflowID, err := uuid.Parse(workflowIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_workflow_id",
			Message: "Invalid workflow ID",
		})
	}

	edges, err := h.service.GetWorkflowEdgesByWorkflow(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get workflow edges",
		})
	}

	response := make([]*domain.WorkflowEdgeResponse, len(edges))
	for i, edge := range edges {
		response[i] = edge.ToResponse()
	}

	return c.JSON(fiber.Map{
		"edges": response,
		"count": len(response),
	})
}

// UpdateWorkflowEdge handles updating a workflow edge
// PUT /api/workflow-edges/:id
func (h *WorkflowEdgeHandler) UpdateWorkflowEdge(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow edge ID",
		})
	}

	var req domain.UpdateWorkflowEdgeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.service.UpdateWorkflowEdge(c.Context(), id, &req); err != nil {
		if errors.Is(err, domain.ErrWorkflowEdgeNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow edge not found",
			})
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "invalid_input",
				Message: "Invalid node IDs provided",
			})
		}
		if errors.Is(err, domain.ErrForeignKeyViolation) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "invalid_nodes",
				Message: "Source or target node does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update workflow edge",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteWorkflowEdge handles deleting a workflow edge
// DELETE /api/workflow-edges/:id
func (h *WorkflowEdgeHandler) DeleteWorkflowEdge(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow edge ID",
		})
	}

	if err := h.service.DeleteWorkflowEdge(c.Context(), id); err != nil {
		if errors.Is(err, domain.ErrWorkflowEdgeNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Workflow edge not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete workflow edge",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
