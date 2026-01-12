package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

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

func IsDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") || strings.Contains(msg, "unique constraint")
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
	return SuccessResponse(c, fiber.StatusCreated, message, data)
}

func OK(c *fiber.Ctx, message string, data interface{}) error {
	return SuccessResponse(c, fiber.StatusOK, message, data)
}

func NotFound(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusNotFound, message, nil)
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusUnauthorized, message, nil)
}

func Forbidden(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusForbidden, message, nil)
}

func BadRequest(c *fiber.Ctx, message string, errors interface{}) error {
	return ErrorResponse(c, fiber.StatusBadRequest, message, errors)
}

func UnprocessableEntity(c *fiber.Ctx, message string, errors interface{}) error {
	return ErrorResponse(c, fiber.StatusUnprocessableEntity, message, errors)
}

func InternalServerError(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, message, nil)
}

func Conflict(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusConflict, message, nil)
}
