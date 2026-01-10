package handlers

import (
	"TugasAkhir/middleware"
	"TugasAkhir/models"
	"TugasAkhir/services"
	"TugasAkhir/utils/events"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Gunakan nama struct sesuai request Anda
type LetterKeluarHandler struct {
	db          *gorm.DB
	permService *services.PermissionService
}

func NewLetterKeluarHandler(db *gorm.DB) *LetterKeluarHandler {
	return &LetterKeluarHandler{
		db:          db,
		permService: services.NewPermissionService(db),
	}
}

// CreateSuratKeluar - Handler untuk membuat surat keluar baru
func (h *LetterKeluarHandler) CreateSuratKeluar(c *fiber.Ctx) error {
	// 1. Ambil User dari Context (Middleware)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 2. Parse Request Body (Format JSON tetap Bahasa Indonesia agar frontend tidak error)
	var req struct {
		NomorSurat         string `json:"nomor_surat"`
		Judul              string `json:"judul_surat"`
		Tujuan             string `json:"tujuan"`
		IsiSurat           string `json:"isi_surat"`
		Scope              string `json:"scope"`
		AssignedVerifierID *uint  `json:"assigned_verifier_id"`
		FilePath           string `json:"file_path"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 3. Cek Permission: Apakah user boleh buat surat dengan Scope ini?
	canCreate, _ := h.permService.CanUserCreateLetter(user, req.Scope)
	if !canCreate {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Anda tidak memiliki izin membuat surat dengan scope ini"})
	}

	// 4. Validasi: Surat Keluar WAJIB punya verifikator (Manajer)
	if req.AssignedVerifierID == nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Surat keluar wajib memilih verifikator (Manajer)",
		})
	}

	// 5. Buat Object Surat
	letter := models.Letter{
		NomorSurat:         req.NomorSurat,
		JudulSurat:         req.Judul,
		BidangTujuan:       req.Tujuan, // Mapping field Tujuan -> BidangTujuan
		IsiSurat:           req.IsiSurat,
		JenisSurat:         models.LetterKeluar,          // Tipe Surat Keluar
		Scope:              req.Scope,                    // Internal / Eksternal
		Status:             models.StatusPerluVerifikasi, // Langsung masuk status verifikasi
		CreatedByID:        user.ID,
		AssignedVerifierID: req.AssignedVerifierID,
		FilePath:           req.FilePath,
		Prioritas:          models.PriorityBiasa, // Default
	}

	// 6. Simpan ke Database
	if err := h.db.Create(&letter).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat surat"})
	}

	// 7. Kirim Notifikasi (Pakai Event Bus yang sudah ada di Utils Anda)
	// Ini akan ditangkap oleh utils/fcm/fcm.go
	events.LetterEventBus <- events.LetterEvent{
		Type:   events.LetterCreated,
		Letter: letter,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Surat keluar berhasil dibuat",
		"data":    letter,
	})
}

// UpdateSurat - Handler untuk edit surat (Hanya jika Draft / Perlu Revisi)
func (h *LetterKeluarHandler) UpdateSurat(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	letterID, _ := c.ParamsInt("id")

	// Get Data Surat
	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Letter not found"})
	}

	// Permission: Hanya pembuat yang boleh edit
	if letter.CreatedByID != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Hanya pembuat surat yang bisa mengedit"})
	}

	// Validasi Status: Hanya boleh edit jika Draft atau Perlu Revisi
	if letter.Status != models.StatusDraft && letter.Status != models.StatusPerluRevisi {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Surat sedang diproses, tidak bisa diedit"})
	}

	var req struct {
		NomorSurat         string `json:"nomor_surat"`
		Judul              string `json:"judul_surat"`
		IsiSurat           string `json:"isi_surat"`
		Tujuan             string `json:"tujuan"`
		AssignedVerifierID *uint  `json:"assigned_verifier_id"`
		FilePath           string `json:"file_path"`
	}
	c.BodyParser(&req)

	// Update Field
	letter.NomorSurat = req.NomorSurat
	letter.JudulSurat = req.Judul
	letter.IsiSurat = req.IsiSurat
	letter.BidangTujuan = req.Tujuan
	letter.FilePath = req.FilePath

	if req.AssignedVerifierID != nil {
		letter.AssignedVerifierID = req.AssignedVerifierID
	}

	h.db.Save(letter)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Surat berhasil diupdate",
		"data":    letter,
	})
}

// VerifyLetterApprove - Manajer menyetujui verifikasi
func (h *LetterKeluarHandler) VerifyLetterApprove(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Letter not found"})
	}

	// Cek Permission Verifikasi (Manajer)
	canVerify, _ := h.permService.CanUserVerifyLetter(user, letter)
	if !canVerify {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Anda tidak memiliki izin memverifikasi surat ini"})
	}

	oldStatus := letter.Status

	// Update Status: Perlu Verifikasi -> Perlu Persetujuan
	letter.Status = models.StatusPerluPersetujuan
	letter.VerifiedByID = &user.ID
	h.db.Save(letter)

	// Kirim Notifikasi (Status Moved)
	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return c.JSON(fiber.Map{"success": true, "message": "Surat berhasil diverifikasi"})
}

// VerifyLetterReject - Manajer menolak (revisi)
func (h *LetterKeluarHandler) VerifyLetterReject(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Letter not found"})
	}

	canVerify, _ := h.permService.CanUserVerifyLetter(user, letter)
	if !canVerify {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	oldStatus := letter.Status

	// Update Status: Perlu Verifikasi -> Perlu Revisi
	letter.Status = models.StatusPerluRevisi
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return c.JSON(fiber.Map{"success": true, "message": "Surat dikembalikan untuk revisi"})
}

// ApproveLetterByDirektur - Direktur menyetujui surat
func (h *LetterKeluarHandler) ApproveLetterByDirektur(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not found"})
	}

	// Cek Permission Approval (Direktur)
	canApprove, _ := h.permService.CanUserApproveLetter(user, letter)
	if !canApprove {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	oldStatus := letter.Status

	// Update Status: Perlu Persetujuan -> Disetujui
	letter.Status = models.StatusDisetujui
	letter.DisposedByID = &user.ID // Direktur tercatat di disposed_by
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return c.JSON(fiber.Map{"success": true, "message": "Surat berhasil disetujui"})
}

// RejectLetterByDirektur - Direktur menolak surat
func (h *LetterKeluarHandler) RejectLetterByDirektur(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not found"})
	}

	canApprove, _ := h.permService.CanUserApproveLetter(user, letter)
	if !canApprove {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	oldStatus := letter.Status

	// Update Status: Perlu Persetujuan -> Perlu Revisi
	letter.Status = models.StatusPerluRevisi
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return c.JSON(fiber.Map{"success": true, "message": "Surat ditolak"})
}

// ArchiveLetter - Staf mengarsipkan surat final
func (h *LetterKeluarHandler) ArchiveLetter(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not found"})
	}

	canArchive, _ := h.permService.CanUserArchiveLetter(user, letter)
	if !canArchive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	oldStatus := letter.Status
	letter.Status = models.StatusDiarsipkan
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return c.JSON(fiber.Map{"success": true, "message": "Surat diarsipkan"})
}

// GetAvailableVerifiers - Helper untuk dropdown Manajer di Frontend
func (h *LetterKeluarHandler) GetAvailableVerifiers(c *fiber.Ctx) error {
	scope := c.Query("scope")
	var verifiers []models.User
	query := h.db.Model(&models.User{})

	// Filter berdasarkan Scope
	if strings.EqualFold(scope, models.ScopeEksternal) {
		query = query.Where("role IN ?", []models.Role{models.RoleManajerKPP, models.RoleManajerPemas})
	} else if strings.EqualFold(scope, models.ScopeInternal) {
		query = query.Where("role = ?", models.RoleManajerPKL)
	} else {
		query = query.Where("role IN ?", []models.Role{models.RoleManajerKPP, models.RoleManajerPemas, models.RoleManajerPKL})
	}

	// Hanya ambil field penting (jangan password hash)
	query.Select("id, username, email, role, jabatan").Find(&verifiers)

	return c.JSON(fiber.Map{"success": true, "data": verifiers})
}
