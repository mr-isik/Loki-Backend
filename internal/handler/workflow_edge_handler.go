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
// @Summary Create workflow edge
// @Description Create a connection between two nodes in a workflow
// @Tags Workflow Edges
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateWorkflowEdgeRequest true "Edge details"
// @Success 200 {object} domain.WorkflowEdgeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-edges [post]
func (h *WorkflowEdgeHandler) CreateWorkflowEdge(c *fiber.Ctx) error {
	var req domain.CreateWorkflowEdgeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	edge, err := h.service.CreateWorkflowEdge(c.Context(), &req)
	if err != nil {
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

	return c.JSON(edge)
}

// GetWorkflowEdge handles retrieving a workflow edge by ID
// @Summary Get workflow edge by ID
// @Description Retrieve workflow edge information by ID
// @Tags Workflow Edges
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow Edge ID (UUID)"
// @Success 200 {object} domain.WorkflowEdgeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-edges/{id} [get]
func (h *WorkflowEdgeHandler) GetWorkflowEdge(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid workflow edge ID",
		})
	}

	edge, err := h.service.GetWorkflowEdgeByID(c.Context(), id)
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

	return c.JSON(edge)
}

// GetWorkflowEdgesByWorkflow handles retrieving all edges for a workflow
// @Summary Get workflow edges
// @Description Retrieve all edges (connections) in a workflow
// @Tags Workflow Edges
// @Produce json
// @Security BearerAuth
// @Param workflow_id path string true "Workflow ID (UUID)"
// @Success 200 {object} []domain.WorkflowEdge "Returns edges array"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{workflow_id}/edges [get]
func (h *WorkflowEdgeHandler) GetWorkflowEdgesByWorkflow(c *fiber.Ctx) error {
	workflowIDParam := c.Params("workflow_id")
	workflowID, err := uuid.Parse(workflowIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_workflow_id",
			Message: "Invalid workflow ID",
		})
	}

	edges, err := h.service.GetWorkflowEdgesByWorkflowID(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get workflow edges",
		})
	}

	return c.JSON(fiber.Map{
		"edges": edges,
		"count": len(edges),
	})
}

// UpdateWorkflowEdge handles updating a workflow edge
// @Summary Update workflow edge
// @Description Update workflow edge information
// @Tags Workflow Edges
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow Edge ID (UUID)"
// @Param request body domain.UpdateWorkflowEdgeRequest true "Edge update details"
// @Success 200 {object} domain.WorkflowEdgeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-edges/{id} [put]
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

	edge, err := h.service.UpdateWorkflowEdge(c.Context(), id, &req)
	if err != nil {
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

	return c.JSON(edge)
}

// DeleteWorkflowEdge handles deleting a workflow edge
// @Summary Delete workflow edge
// @Description Delete a workflow edge (connection)
// @Tags Workflow Edges
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow Edge ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-edges/{id} [delete]
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
