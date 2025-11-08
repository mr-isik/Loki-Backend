package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type UserHandler struct {
	service domain.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service domain.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// CreateUser handles user creation
// @Summary Create a new user
// @Description Create a new user (admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateUserRequest true "User details"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req domain.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	_, err := h.service.CreateUser(c.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Error:   "user_exists",
				Message: "User with this email already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create user",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetUser handles retrieving a user by ID
// @Summary Get user by ID
// @Description Retrieve user information by ID
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} domain.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid user ID format",
		})
	}

	user, err := h.service.GetUser(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get user",
		})
	}

	return c.JSON(user)
}

// UpdateUser handles updating a user
// @Summary Update user
// @Description Update user information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body domain.UpdateUserRequest true "User update details"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [patch]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid user ID format",
		})
	}

	var req domain.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	_, err = h.service.UpdateUser(c.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "User not found",
			})
		}
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Error:   "email_taken",
				Message: "Email is already taken",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update user",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteUser handles deleting a user
// @Summary Delete user
// @Description Soft delete a user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid user ID format",
		})
	}

	if err := h.service.DeleteUser(c.Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete user",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
