package handlers

import (
	"TugasAkhir/middleware"
	"TugasAkhir/models"
	"TugasAkhir/services"
	"TugasAkhir/utils/storage"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LetterCommonHandler struct {
	db          *gorm.DB
	permService *services.PermissionService
}

func NewLetterCommonHandler(db *gorm.DB) *LetterCommonHandler {
	return &LetterCommonHandler{
		db:          db,
		permService: services.NewPermissionService(db),
	}
}

// GetLetterByID - Melihat detail surat (Generic untuk Masuk & Keluar)
func (h *LetterCommonHandler) GetLetterByID(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	letterID, _ := c.ParamsInt("id")

	// Preload relasi lengkap agar frontend senang
	var letter models.Letter
	if err := h.db.Preload("CreatedBy").Preload("AssignedVerifier").Preload("VerifiedBy").Preload("DisposedBy").First(&letter, letterID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Letter not found"})
	}

	// Cek Permission via Service (Logic baru yang sudah kita buat)
	canView, _ := h.permService.CanUserViewLetter(user, &letter)
	if !canView {
		return c.Status(403).JSON(fiber.Map{"error": "Anda tidak memiliki akses melihat surat ini"})
	}

	// [FIX] Generate Presigned URL agar gambar bisa dibuka di frontend
	// Meskipun bucket public, URL lengkap tetap dibutuhkan
	if letter.FilePath != "" {
		url, err := storage.GetPresignedURL(letter.FilePath)
		if err == nil {
			letter.FilePath = url
		}
	}

	return c.JSON(fiber.Map{"success": true, "data": letter})
}

// DeleteLetter - Soft Delete / Cancel (Hanya Admin atau Pembuat saat Draft)
func (h *LetterCommonHandler) DeleteLetter(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	letterID, _ := c.ParamsInt("id")
	var letter models.Letter
	if err := h.db.First(&letter, letterID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Letter not found"})
	}

	// Logic: Hanya Admin, ATAU Pembuat surat jika status masih Draft
	isOwnerDraft := letter.CreatedByID == user.ID && letter.Status == models.StatusDraft
	if !user.IsAdmin() && !isOwnerDraft {
		return c.Status(403).JSON(fiber.Map{"error": "Dilarang menghapus surat ini"})
	}

	h.db.Delete(&letter)
	return c.JSON(fiber.Map{"success": true, "message": "Surat berhasil dihapus"})
}
