package utils

import "github.com/gofiber/fiber/v2"

// APIResponse defines the common structure returned by the API.
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// JSONSuccess sends a successful JSON response with the provided status code, message and data.
func JSONSuccess(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	if statusCode == 0 {
		statusCode = fiber.StatusOK
	}

	response := APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}

	return c.Status(statusCode).JSON(response)
}

// JSONError sends an error JSON response with the provided status code, message and error details.
func JSONError(c *fiber.Ctx, statusCode int, message string, errDetail interface{}) error {
	if statusCode == 0 {
		statusCode = fiber.StatusInternalServerError
	}

	response := APIResponse{
		Status:  "error",
		Message: message,
		Errors:  errDetail,
	}

	return c.Status(statusCode).JSON(response)
}
