package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type WorkflowNodeHandler struct {
	service domain.WorkflowNodeService
}

// NewWorkflowNodeHandler creates a new workflow node handler
func NewWorkflowNodeHandler(service domain.WorkflowNodeService) *WorkflowNodeHandler {
	return &WorkflowNodeHandler{
		service: service,
	}
}

// CreateWorkflowNode handles workflow node creation
// @Summary Create workflow node
// @Description Create a new node in a workflow
// @Tags Workflow Nodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateWorkflowNodeRequest true "Node details"
// @Success 200 {object} domain.WorkflowNodeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-nodes [post]
func (h *WorkflowNodeHandler) CreateWorkflowNode(c *fiber.Ctx) error {
	var req domain.CreateWorkflowNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	workflowNode, err := h.service.CreateWorkflowNode(c.Context(), &req)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create workflow node",
		})
	}

	return c.JSON(workflowNode)
}

// GetWorkflowNode handles retrieving a workflow node by ID
// @Summary Get workflow node by ID
// @Description Retrieve workflow node information by ID
// @Tags Workflow Nodes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow Node ID (UUID)"
// @Success 200 {object} domain.WorkflowNodeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-nodes/{id} [get]
func (h *WorkflowNodeHandler) GetWorkflowNode(c *fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid workflow node ID",
		})
	}

	workflowNode, err := h.service.GetWorkflowNode(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get workflow node",
		})
	}

	return c.JSON(workflowNode)
}

// UpdateWorkflowNode handles updating a workflow node by ID
// @Summary Update workflow node
// @Description Update workflow node information
// @Tags Workflow Nodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow Node ID (UUID)"
// @Param request body domain.UpdateWorkflowNodeRequest true "Node update details"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-nodes/{id} [put]
func (h *WorkflowNodeHandler) UpdateWorkflowNode(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid workflow node ID",
		})
	}
	var req domain.UpdateWorkflowNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}
	if err := h.service.UpdateWorkflowNode(c.Context(), id, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update workflow node",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteWorkflowNode handles deleting a workflow node by ID
// @Summary Delete workflow node
// @Description Delete a workflow node
// @Tags Workflow Nodes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow Node ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflow-nodes/{id} [delete]
func (h *WorkflowNodeHandler) DeleteWorkflowNode(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid workflow node ID",
		})
	}

	if err := h.service.DeleteWorkflowNode(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete workflow node",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetWorkflowNodes handles retrieving workflow nodes by workflow ID
// @Summary Get workflow nodes
// @Description Retrieve all nodes in a workflow
// @Tags Workflow Nodes
// @Produce json
// @Security BearerAuth
// @Param workflow_id path string true "Workflow ID (UUID)"
// @Success 200 {object} []domain.WorkflowNode "Returns nodes array"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /workflows/{workflow_id}/nodes [get]
func (h *WorkflowNodeHandler) GetWorkflowNodes(c *fiber.Ctx) error {
	workflowIDParam := c.Params("workflow_id")
	workflowID, err := uuid.Parse(workflowIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid workflow ID",
		})
	}

	workflowNodes, err := h.service.GetWorkflowNodesByWorkflowID(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get workflow nodes",
		})
	}

	return c.JSON(fiber.Map{
		"nodes": workflowNodes,
		"count": len(workflowNodes),
	})
}
