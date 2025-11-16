package handlers

import (
	"TugasAkhir/utils"
	"TugasAkhir/utils/events"
	"TugasAkhir/utils/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"TugasAkhir/config"
	letterdto "TugasAkhir/dto/letters"
	"TugasAkhir/middleware"
	"TugasAkhir/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// POST /api/letters
func CreateLetter(c *fiber.Ctx) error {
	jsonData := c.FormValue("data")
	if jsonData == "" {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "form field 'data' (JSON string) is required", nil)
	}

	var req letterdto.CreateLetterRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "invalid 'data' field (must be a valid JSON string)", err.Error())
	}

	if validationErrors := req.Validate(); len(validationErrors) > 0 {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "validation error", validationErrors)
	}

	fileHeader, err := c.FormFile("file_surat")
	if err != nil {
		if err == http.ErrMissingFile {
			return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "form field 'file_surat' (file upload) is required", nil)
		}
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "invalid file upload", err.Error())
	}

	claims, ok := middleware.GetJWTClaims(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "authorization context missing", nil)
	}

	ext := filepath.Ext(fileHeader.Filename)
	s3key := fmt.Sprintf("surat/%d%s%s", claims.UserID, uuid.NewString(), ext)

	if _, err := storage.UploadFile(c.Context(), fileHeader, s3key); err != nil {
		return utils.ErrorResponse(c, fiber.ErrInternalServerError.Code, "upload file error", err.Error())
	}

	letter := req.ToModel()
	letter.CreatedByID = &claims.UserID
	letter.FilePath = s3key

	if err := config.DB.Create(&letter).Error; err != nil {
		go storage.DeleteFile(context.Background(), s3key)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to create letter", err.Error())
	}

	events.LetterEventBus <- events.LetterEvent{
		Type:   events.LetterCreated,
		Letter: letter,
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
	if err := config.DB.Order("id DESC").Limit(limit).Offset(offset).Find(&letters).Error; err != nil {
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

	claims, ok := middleware.GetJWTClaims(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "authorization context missing", nil)
	}
	userRole := claims.Role

	//isAdmin?
	if userRole != models.RoleAdmin {
		if err := validateWorkflowUpdate(userRole, &letter, &req); err != nil {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
	}

	oldStatus := letter.Status

	letterdto.ApplyUpdate(&letter, &req)

	if userRole != models.RoleAdmin {
		if userRole == models.RoleADC && letter.Status == models.StatusBelumDisposisi && oldStatus == models.StatusPerluVerifikasi {
			letter.VerifiedByID = &claims.UserID
		}
		if userRole == models.RoleDirektur && letter.Status == models.StatusSudahDisposisi && oldStatus == models.StatusBelumDisposisi {
			letter.DisposedByID = &claims.UserID
		}
	}

	if err := config.DB.Save(&letter).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to update letter", err.Error())
	}

	if oldStatus != letter.Status {
		events.LetterEventBus <- events.LetterEvent{
			Type:      events.LetterStatusMoved,
			Letter:    letter,
			OldStatus: oldStatus,
		}
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "letter updated successfully", letterdto.NewLetterResponse(&letter))
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
				log.Println("Failed to delete s3 object %s during letter deletion: %v", key, err)
			} else {
				log.Printf("Successfully deleted S3 object %s", key)
			}
		}(letter.FilePath)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "letter deleted successfully", nil)
}

func validateWorkflowUpdate(role models.Role, letter *models.Letter, req *letterdto.UpdateLetterRequest) error {

	newStatus := letter.Status
	if req.Status != nil {
		newStatus = *req.Status
	}

	switch role {
	//Rule Bagian Umum
	case models.RoleBagianUmum:
		if letter.Status != models.StatusDraft && newStatus == models.StatusDraft {
		} else if letter.Status != models.StatusDraft {
			return errors.New("bagian umum hanya dapat mengedit surat dengan status 'draft'")
		}

		//Rule ADC
	case models.RoleADC:
		if letter.Status != models.StatusPerluVerifikasi {
			return errors.New("ADC hanya dapat memproses surat dengan status 'perlu_disposisi'")
		}

		if newStatus != letter.Status && newStatus != models.StatusBelumDisposisi {
			return errors.New("ADC hanya dapat mengubah status ke 'belum_disposisi' (verifikasi)")
		}

		if req.Disposisi != nil || req.TanggalDisposisi != nil || req.BidangTujuan != nil {
			return errors.New("ADC tidak memiliki izin untuk mengisi data disposisi")
		}
		//rule direktur
	case models.RoleDirektur:
		if req.Pengirim != nil || req.NomorSurat != nil || req.NomorAgenda != nil ||
			req.JenisSurat != nil || req.IsiSurat != nil ||
			req.JudulSurat != nil || req.Kesimpulan != nil || req.FilePath != nil {
			return errors.New("direktur tidak memiliki izin untuk mengubah konten utama surat, hanya disposisi")
		}

		if newStatus != letter.Status {
			if newStatus != models.StatusSudahDisposisi {
				return errors.New("direktur hanya dapat mengubah status ke 'sudah_disposisi'")
			}
			if letter.Status != models.StatusBelumDisposisi {
				return errors.New("surat ini belum siap untuk disposisi oleh Direktur (status saat ini: " + string(letter.Status) + ")")
			}
		}

	default:
		return errors.New("role Anda tidak memiliki izin untuk memperbarui surat")
	}

	return nil
}
