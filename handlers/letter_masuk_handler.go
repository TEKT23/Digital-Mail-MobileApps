package handlers

import (
	"TugasAkhir/middleware"
	"TugasAkhir/models"
	"TugasAkhir/services"
	"TugasAkhir/utils/events"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LetterMasukHandler struct {
	db          *gorm.DB
	permService *services.PermissionService
}

func NewLetterMasukHandler(db *gorm.DB) *LetterMasukHandler {
	return &LetterMasukHandler{
		db:          db,
		permService: services.NewPermissionService(db),
	}
}

// CreateSuratMasuk - Staf input surat masuk (BYPASS manajer, langsung ke Direktur)
func (h *LetterMasukHandler) CreateSuratMasuk(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Request Body
	var req struct {
		NomorSurat   string `json:"nomor_surat"`
		Pengirim     string `json:"pengirim"`
		Judul        string `json:"judul_surat"`
		TanggalSurat string `json:"tanggal_surat"` // Format YYYY-MM-DD
		TanggalMasuk string `json:"tanggal_masuk"` // Format YYYY-MM-DD
		Scope        string `json:"scope"`
		FileScanPath string `json:"file_scan_path"`
		Prioritas    string `json:"prioritas"`
		IsiRingkas   string `json:"isi_surat"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 1. Cek Permission (Internal vs Eksternal)
	canCreate, _ := h.permService.CanUserCreateLetter(user, req.Scope)
	if !canCreate {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Anda tidak memiliki izin membuat surat masuk dengan scope ini"})
	}

	// Helper parsing tanggal
	var tglSurat, tglMasuk *time.Time
	if req.TanggalSurat != "" {
		if t, err := time.Parse("2006-01-02", req.TanggalSurat); err == nil {
			tglSurat = &t
		}
	}
	if req.TanggalMasuk != "" {
		if t, err := time.Parse("2006-01-02", req.TanggalMasuk); err == nil {
			tglMasuk = &t
		}
	}

	// 2. Buat Object Surat
	letter := models.Letter{
		NomorSurat:   req.NomorSurat,
		Pengirim:     req.Pengirim,
		JudulSurat:   req.Judul,
		JenisSurat:   models.LetterMasuk,
		Scope:        req.Scope,
		Status:       models.StatusBelumDisposisi, // Langsung ke status ini (Siap Disposisi)
		CreatedByID:  user.ID,
		FilePath:     req.FileScanPath,
		Prioritas:    models.Priority(req.Prioritas),
		IsiSurat:     req.IsiRingkas,
		TanggalSurat: tglSurat,
		TanggalMasuk: tglMasuk,
		// AssignedVerifierID: nil (Surat Masuk tidak butuh verifikator)
	}

	// Default prioritas
	if letter.Prioritas == "" {
		letter.Prioritas = models.PriorityBiasa
	}

	if err := h.db.Create(&letter).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mencatat surat masuk"})
	}

	// 3. Kirim Notifikasi (Event Bus)
	// utils/fcm akan menangkap ini dan mengirim notif ke DIREKTUR
	events.LetterEventBus <- events.LetterEvent{
		Type:   events.LetterCreated,
		Letter: letter,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Surat masuk berhasil dicatat dan dikirim ke Direktur",
		"data":    letter,
	})
}

// DisposeSuratMasuk - Direktur memberikan instruksi disposisi
func (h *LetterMasukHandler) DisposeSuratMasuk(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	letterID, _ := c.ParamsInt("id")

	// Validasi Permission
	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Letter not found"})
	}

	canDispose, _ := h.permService.CanUserDisposeLetter(user, letter)
	if !canDispose {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Anda tidak berhak mendisposisi surat ini"})
	}

	var req struct {
		InstruksiDisposisi string `json:"instruksi_disposisi"`
		TujuanDisposisi    string `json:"tujuan_disposisi"` // Bidang Tujuan
		Catatan            string `json:"catatan"`
	}
	c.BodyParser(&req)

	oldStatus := letter.Status
	now := time.Now()

	// Update Data Disposisi
	letter.Status = models.StatusSudahDisposisi
	letter.Disposisi = req.InstruksiDisposisi
	letter.BidangTujuan = req.TujuanDisposisi
	letter.DisposedByID = &user.ID
	letter.TanggalDisposisi = &now

	if req.Catatan != "" {
		letter.Disposisi += " | Catatan: " + req.Catatan
	}

	h.db.Save(letter)

	// Notif ke Staf Pembuat
	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return c.JSON(fiber.Map{"success": true, "message": "Disposisi berhasil disimpan", "data": letter})
}

// ArchiveSuratMasuk - Staf arsip surat yang sudah didisposisi
func (h *LetterMasukHandler) ArchiveSuratMasuk(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	letterID, _ := c.ParamsInt("id")
	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}

	canArchive, _ := h.permService.CanUserArchiveLetter(user, letter)
	if !canArchive {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	oldStatus := letter.Status
	letter.Status = models.StatusDiarsipkan
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return c.JSON(fiber.Map{"success": true, "message": "Surat masuk diarsipkan"})
}

// GetLettersMasukForDisposition - Helper List untuk Direktur
func (h *LetterMasukHandler) GetLettersMasukForDisposition(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	if !user.IsDirektur() {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	var letters []models.Letter
	h.db.Where("jenis_surat = ? AND status = ?", models.LetterMasuk, models.StatusBelumDisposisi).
		Preload("CreatedBy").
		Order("created_at DESC").
		Find(&letters)

	return c.JSON(fiber.Map{"success": true, "data": letters})
}
