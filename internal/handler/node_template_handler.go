package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type NodeTemplateHandler struct {
	service domain.NodeTemplateService
}

func NewNodeTemplateHandler(service domain.NodeTemplateService) *NodeTemplateHandler {
	return &NodeTemplateHandler{
		service: service,
	}
}

// ListNodeTemplates handles retrieving all node templates
// @Summary List node templates
// @Description Retrieve all available node templates
// @Tags Node Templates
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []domain.NodeTemplateResponse "Returns templates array and count"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /node-templates [get]
func (h *NodeTemplateHandler) ListNodeTemplates(c *fiber.Ctx) error {
	templates, err := h.service.ListNodeTemplates(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve node templates",
		})
	}

	return c.JSON(fiber.Map{
		"templates": templates,
		"count":     len(templates),
	})
}

// GetNodeTemplate handles retrieving a node template by ID
// @Summary Get node template by ID
// @Description Retrieve node template information by ID
// @Tags Node Templates
// @Produce json
// @Security BearerAuth
// @Param id path string true "Node Template ID (UUID)"
// @Success 200 {object} domain.NodeTemplateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /node-templates/{id} [get]
func (h *NodeTemplateHandler) GetNodeTemplate(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid node template ID",
		})
	}

	template, err := h.service.GetNodeTemplate(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNodeTemplateNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "Node template not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve node template",
		})
	}

	return c.JSON(template)
}
