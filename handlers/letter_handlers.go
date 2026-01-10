package handlers

import (
	"context"
	"log"
	"strconv"

	"TugasAkhir/config"
	letterdto "TugasAkhir/dto/letters"
	"TugasAkhir/models"
	"TugasAkhir/utils"
	"TugasAkhir/utils/storage"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// NOTE: CreateLetter dan UpdateLetter SUDAH DIHAPUS dari sini
// Karena sudah digantikan oleh letter_keluar_handlers.go dan letter_masuk_handlers.go

// GET /api/letters/:id
func GetLetterByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var letter models.Letter

	if err := config.DB.
		Preload("CreatedBy").
		Preload("VerifiedBy").
		Preload("DisposedBy").
		First(&letter, "id = ?", id).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "letter not found", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve letter", err.Error())
	}

	if letter.FilePath != "" {
		predesignURL, err := storage.GetPresignedURL(letter.FilePath)
		if err != nil {
			log.Printf("Failed to get presigned URL for key %s : %sv", letter.FilePath, err)
		} else {
			letter.FilePath = predesignURL
		}
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
	if err := config.DB.
		Preload("CreatedBy").
		Preload("VerifiedBy").
		Preload("DisposedBy").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&letters).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve letters", err.Error())
	}

	for i := range letters {
		letter := &letters[i]

		if letter.FilePath != "" {
			presignedURL, err := storage.GetPresignedURL(letter.FilePath)
			if err != nil {
				log.Printf("Failed to presign URL for key %s in list: %v", letter.FilePath, err)
				letter.FilePath = ""
			} else {
				letter.FilePath = presignedURL
			}
		}
	}

	responses := make([]letterdto.LetterResponse, 0, len(letters))
	for i := range letters {
		responses = append(responses, letterdto.NewLetterResponse(&letters[i]))
	}

	meta := utils.PaginationMeta{Page: page, Limit: limit, Total: total}
	return utils.PaginatedResponse(c, fiber.StatusOK, "letters retrieved successfully", responses, meta)
}

// DELETE /api/letters/:id
func DeleteLetter(c *fiber.Ctx) error {
	id := c.Params("id")

	var letter models.Letter
	if err := config.DB.First(&letter, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "letter not found", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve letter for deletion", err.Error())
	}

	// Permission: Hanya Admin yang boleh delete
	// (Jika ingin Staf juga boleh, tambahkan cek CreatedByID)

	result := config.DB.Delete(&models.Letter{}, "id = ?", id)
	if result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete letter", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "letter not found", nil)
	}

	if letter.FilePath != "" {
		go func(key string) {
			if err := storage.DeleteFile(context.Background(), key); err != nil {
				log.Printf("Failed to delete s3 object %s during letter deletion: %v", key, err)
			} else {
				log.Printf("Successfully deleted S3 object %s", key)
			}
		}(letter.FilePath)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "letter deleted successfully", nil)
}
