package utils

import "github.com/gofiber/fiber/v2"

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type PaginationMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
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

func ErrorResponse(c *fiber.Ctx, statusCode int, message string, errDetail interface{}) error {
	if statusCode == 0 {
		statusCode = fiber.StatusInternalServerError
	}
	if message == "" {
		message = fiber.ErrInternalServerError.Message
	}

	response := APIResponse{
		Status:  "error",
		Message: message,
		Errors:  errDetail,
	}

	return c.Status(statusCode).JSON(response)
}
func PaginatedResponse(c *fiber.Ctx, statusCode int, message string, data interface{}, meta PaginationMeta) error {
	payload := fiber.Map{
		"items": data,
		"meta":  meta,
	}

	return SuccessResponse(c, statusCode, message, payload)
}
