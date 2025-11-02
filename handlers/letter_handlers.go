package handlers

import (
	"TugasAkhir/utils"
	"strconv"

	"TugasAkhir/config"
	letterdto "TugasAkhir/dto/letters"
	"TugasAkhir/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// POST /api/letters
func CreateLetter(c *fiber.Ctx) error {
	var req letterdto.CreateLetterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "invalid request body", err.Error())
	}

	if validationErrors := req.Validate(); len(validationErrors) > 0 {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "validation error", validationErrors)
	}

	letter := req.ToModel()
	if err := config.DB.Create(&letter).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to create letter", err.Error())
	}
	response := letterdto.NewLetterResponse(&letter)
	return utils.SuccessResponse(c, fiber.StatusCreated, "letter created successfully", response)
}

// GET /api/letters/:id
func GetLetterByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var letter models.Letter
	if err := config.DB.First(&letter, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "letter not found", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve letter", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "letter retrieved successfully", letterdto.NewLetterResponse(&letter))
}

// GET /api/letters?page=&limit=
func ListLetters(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	if err := config.DB.Model(&models.Letter{}).Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to count letters", err.Error())
	}

	var letters []models.Letter
	if err := config.DB.Order("id DESC").Limit(limit).Offset(offset).Find(&letters).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve letters", err.Error())
	}

	responses := make([]letterdto.LetterResponse, 0, len(letters))
	for i := range letters {
		responses = append(responses, letterdto.NewLetterResponse(&letters[i]))
	}

	meta := utils.PaginationMeta{Page: page, Limit: limit, Total: total}
	return utils.PaginatedResponse(c, fiber.StatusOK, "letters retrieved successfully", responses, meta)
}

// PUT /api/letters/:id
func UpdateLetter(c *fiber.Ctx) error {
	id := c.Params("id")
	var letter models.Letter
	if err := config.DB.First(&letter, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "letter not found", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve letter", err.Error())
	}

	var req letterdto.UpdateLetterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "invalid request body", err.Error())
	}

	if validationErrors := req.Validate(); len(validationErrors) > 0 {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "validation error", validationErrors)
	}

	letterdto.ApplyUpdate(&letter, &req)
	if err := config.DB.Save(&letter).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to update letter", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "letter updated successfully", letterdto.NewLetterResponse(&letter))
}

// DELETE /api/letters/:id
func DeleteLetter(c *fiber.Ctx) error {
	id := c.Params("id")
	result := config.DB.Delete(&models.Letter{}, "id = ?", id)
	if result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete letter", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "letter not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "letter deleted successfully", nil)
}
