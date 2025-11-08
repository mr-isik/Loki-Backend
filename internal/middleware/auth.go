package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mr-isik/loki-backend/internal/util"
)

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(jwtManager *util.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authorization header required",
			})
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "invalid_token",
				"message": "Invalid authorization header format. Expected: Bearer <token>",
			})
		}

		token := parts[1]

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			if err == util.ErrExpiredToken {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "token_expired",
					"message": "Token has expired",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "invalid_token",
				"message": "Invalid or malformed token",
			})
		}

		// Set user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("name", claims.Name)

		return c.Next()
	}
}

// OptionalAuthMiddleware creates an optional JWT authentication middleware
// It will set user info in context if valid token is provided, but won't reject requests without tokens
func OptionalAuthMiddleware(jwtManager *util.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}

		token := parts[1]
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			return c.Next()
		}

		// Set user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("name", claims.Name)

		return c.Next()
	}
}
