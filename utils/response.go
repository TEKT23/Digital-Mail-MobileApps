package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// PaginationMeta untuk data list dengan paging
type PaginationMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// SuccessResponseStruct untuk format 200/201
type SuccessResponseStruct struct {
	Success bool        `json:"success"` // true
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// ErrorResponseStruct untuk format 4xx/5xx
type ErrorResponseStruct struct {
	Success bool        `json:"success"`          // false
	Code    int         `json:"code"`             // HTTP Status Code
	Message string      `json:"message"`          // Pesan Error Utama
	Errors  interface{} `json:"errors,omitempty"` // Detail Error (misal validasi field)
}

// SuccessResponse mengirim response sukses standar
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	if statusCode == 0 {
		statusCode = fiber.StatusOK
	}

	return c.Status(statusCode).JSON(SuccessResponseStruct{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// PaginatedResponse mengirim response list dengan meta pagination
func PaginatedResponse(c *fiber.Ctx, statusCode int, message string, data interface{}, meta PaginationMeta) error {
	return c.Status(statusCode).JSON(SuccessResponseStruct{
		Success: true,
		Message: message,
		Data:    fiber.Map{"items": data},
		Meta:    meta,
	})
}

// ErrorResponse mengirim response error standar
func ErrorResponse(c *fiber.Ctx, statusCode int, message string, errDetail interface{}) error {
	if statusCode == 0 {
		statusCode = fiber.StatusInternalServerError
	}
	// Jika message kosong, gunakan default message dari Fiber
	if message == "" {
		message = fiber.ErrInternalServerError.Message
	}

	return c.Status(statusCode).JSON(ErrorResponseStruct{
		Success: false,
		Code:    statusCode,
		Message: message,
		Errors:  errDetail,
	})
}

// --- Helper Functions Singkat ---

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

// IsDuplicateError mendeteksi error unique constraint dari database
func IsDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") || strings.Contains(msg, "unique constraint")
}
