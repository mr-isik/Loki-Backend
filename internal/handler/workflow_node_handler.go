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
// POST /api/workflow-nodes
func (h *WorkflowNodeHandler) CreateWorkflowNode(c *fiber.Ctx) error {
    var req domain.CreateWorkflowNodeRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Error:   "invalid_request",
            Message: "Invalid request body",
        })
    }

    if err := h.service.CreateWorkflowNode(c.Context(), &req); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Error:   "internal_error",
            Message: "Failed to create workflow node",
        })
    }

    return c.SendStatus(fiber.StatusNoContent)
}

// GetWorkflowNode handles retrieving a workflow node by ID
// GET /api/workflow-nodes/:id
func (h *WorkflowNodeHandler) GetWorkflowNode(c *fiber.Ctx) (*domain.WorkflowNodeResponse, error) {
    idParam := c.Params("id")

    id, err := uuid.Parse(idParam)
    if err != nil {
        return nil, c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Error:   "invalid_request",
            Message: "Invalid workflow node ID",
        })
    }

    workflowNode, err := h.service.GetWorkflowNode(c.Context(), id)
    if err != nil {
        return nil, c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Error:   "internal_error",
            Message: "Failed to get workflow node",
        })
    }

    return workflowNode, nil
}

// UpdateWorkflowNode handles updating a workflow node by ID
// PUT /api/workflow-nodes/:id
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
// DELETE /api/workflow-nodes/:id
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

    return nil
}

// GetWorkflowNodesByWorkflowID handles retrieving workflow nodes by workflow ID
// GET /api/workflows/:workflow_id/nodes
func (h *WorkflowNodeHandler) GetWorkflowNodesByWorkflowID(c *fiber.Ctx) ([]*domain.WorkflowNodeResponse, error) {
    workflowIDParam := c.Params("workflow_id")
    workflowID, err := uuid.Parse(workflowIDParam)
    if err != nil {
        return nil, c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Error:   "invalid_request",
            Message: "Invalid workflow ID",
        })
    }
    workflowNodes, err := h.service.GetWorkflowNodesByWorkflowID(c.Context(), workflowID)
    if err != nil {
        return nil, c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Error:   "internal_error",
            Message: "Failed to get workflow nodes",
        })
    }

    return workflowNodes, c.Status(fiber.StatusNoContent).Send(nil)
}