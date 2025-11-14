package handlers

import (
	"TugasAkhir/utils"
	"TugasAkhir/utils/events"
	"errors"
	"strconv"

	"TugasAkhir/config"
	letterdto "TugasAkhir/dto/letters"
	"TugasAkhir/middleware"
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

	claims, ok := middleware.GetJWTClaims(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "authorization context missing", nil)
	}

	letter := req.ToModel()
	letter.CreatedByID = &claims.UserID
	if err := config.DB.Create(&letter).Error; err != nil {
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
		if userRole == models.RoleADC && letter.Status == models.StatusBelumDisposisi && oldStatus == models.StatusPerluDisposisi {
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
	result := config.DB.Delete(&models.Letter{}, "id = ?", id)
	if result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete letter", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "letter not found", nil)
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
		if letter.Status != models.StatusPerluDisposisi {
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
