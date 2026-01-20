package handlers

import (
	"TugasAkhir/dto/letters"
	"TugasAkhir/middleware"
	"TugasAkhir/models"
	"TugasAkhir/services"
	"TugasAkhir/utils" // Imported for response helpers
	"TugasAkhir/utils/events"
	"TugasAkhir/utils/storage"
	"fmt"
	"path/filepath"
	"strings"
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
		return utils.Unauthorized(c, "Unauthorized")
	}

	// 1. Parsing Form Data (DTO)
	var req letters.CreateLetterMasukRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", err.Error())
	}

	// 2. Validasi Input
	if errMap := req.Validate(); len(errMap) > 0 {
		return utils.BadRequest(c, "Validasi gagal", errMap)
	}

	// 3. Cek Permission
	canCreate, _ := h.permService.CanUserCreateLetter(user, req.Scope, models.LetterMasuk)
	if !canCreate {
		return utils.Forbidden(c, "Anda tidak memiliki izin mencatat surat masuk")
	}

	// 4. Handle File Upload
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return utils.BadRequest(c, "File surat wajib diunggah", nil)
	}

	// Validasi Ekstensi
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".pdf" && ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		return utils.BadRequest(c, "Format file harus PDF atau Gambar", nil)
	}

	// Upload ke Storage
	filename := fmt.Sprintf("surat/masuk_%d%s", time.Now().UnixNano(), ext)
	uploadedPath, err := storage.UploadFile(c.Context(), fileHeader, filename)
	if err != nil {
		return utils.InternalServerError(c, "Gagal mengupload file ke server")
	}

	// 5. Mapping ke Model
	letter := req.ToModel(user.ID, uploadedPath)
	letter.JenisSurat = models.LetterMasuk
	// Status default 'belum_disposisi' sudah di-set di mapper, tapi force set disini agar jelas
	letter.Status = models.StatusBelumDisposisi

	if err := h.db.Create(&letter).Error; err != nil {
		return utils.InternalServerError(c, "Gagal mencatat surat masuk")
	}

	// 6. Kirim Notifikasi
	events.LetterEventBus <- events.LetterEvent{
		Type:   events.LetterCreated,
		Letter: letter,
	}

	return utils.Created(c, "Surat masuk berhasil dicatat dan dikirim ke Direktur", letter)
}

// UpdateSuratMasuk - Staf edit surat masuk (hanya jika belum diarsip/disposisi final)
func (h *LetterMasukHandler) UpdateSuratMasuk(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return utils.Unauthorized(c, "Unauthorized")
	}

	letterID, _ := c.ParamsInt("id")
	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return utils.NotFound(c, "Surat tidak ditemukan")
	}

	// Cek Kepemilikan (Hanya pembuat yang boleh edit)
	if letter.CreatedByID != user.ID {
		return utils.Forbidden(c, "Anda tidak berhak mengedit surat ini")
	}

	// Cek Status (Boleh edit selama belum Diarsipkan)
	// Jika sudah didisposisi, mungkin user ingin update kesalahan data minor?
	// Untuk aman: Hanya boleh edit jika StatusBelumDisposisi
	if letter.Status != models.StatusBelumDisposisi {
		return utils.Conflict(c, "Surat yang sudah didisposisi tidak dapat diedit")
	}

	// Parsing Request
	var req letters.UpdateLetterMasukRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body", err.Error())
	}
	if errMap := req.Validate(); len(errMap) > 0 {
		return utils.BadRequest(c, "Validasi gagal", errMap)
	}

	// Apply Update Metadata
	letters.ApplyUpdateMasuk(letter, &req)

	// Handle File Upload (Optional Replace)
	fileHeader, err := c.FormFile("file")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext != ".pdf" && ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			return utils.BadRequest(c, "Format file harus PDF atau Gambar", nil)
		}
		filename := fmt.Sprintf("surat/masuk_%d%s", time.Now().UnixNano(), ext)
		uploadedPath, err := storage.UploadFile(c.Context(), fileHeader, filename)
		if err != nil {
			return utils.InternalServerError(c, "Gagal mengupload file revisi")
		}
		letter.FilePath = uploadedPath
	}

	h.db.Save(letter)
	return utils.OK(c, "Surat masuk berhasil diperbarui", letter)
}

// GetMySuratMasuk - List surat masuk buatan saya
func (h *LetterMasukHandler) GetMySuratMasuk(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return utils.Unauthorized(c, "Unauthorized")
	}

	var letters []models.Letter
	h.db.Where("created_by_id = ? AND jenis_surat = ?", user.ID, models.LetterMasuk).
		Order("created_at DESC").
		Find(&letters)

	return utils.OK(c, "List surat masuk saya berhasil diambil", letters)
}

// DisposeSuratMasuk - Direktur memberikan instruksi disposisi
func (h *LetterMasukHandler) DisposeSuratMasuk(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return utils.Unauthorized(c, "Unauthorized")
	}

	letterID, _ := c.ParamsInt("id")

	// Validasi Permission
	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return utils.NotFound(c, "Letter not found")
	}

	canDispose, _ := h.permService.CanUserDisposeLetter(user, letter)
	if !canDispose {
		return utils.Forbidden(c, "Anda tidak berhak mendisposisi surat ini")
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

	return utils.OK(c, "Disposisi berhasil disimpan", letter)
}

// ArchiveSuratMasuk - Staf arsip surat yang sudah didisposisi
func (h *LetterMasukHandler) ArchiveSuratMasuk(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return utils.Unauthorized(c, "Unauthorized")
	}

	letterID, _ := c.ParamsInt("id")
	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return utils.NotFound(c, "Not found")
	}

	canArchive, _ := h.permService.CanUserArchiveLetter(user, letter)
	if !canArchive {
		return utils.Forbidden(c, "Forbidden")
	}

	oldStatus := letter.Status
	letter.Status = models.StatusDiarsipkan
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return utils.OK(c, "Surat masuk diarsipkan", nil)
}

// GetLettersMasukForDisposition - Helper List untuk Direktur
func (h *LetterMasukHandler) GetLettersMasukForDisposition(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	if !user.IsDirektur() {
		return utils.Forbidden(c, "Forbidden")
	}

	var letters []models.Letter
	h.db.Where("jenis_surat = ? AND status = ?", models.LetterMasuk, models.StatusBelumDisposisi).
		Preload("CreatedBy").
		Order("created_at DESC").
		Find(&letters)

	return utils.OK(c, "List disposisi berhasil diambil", letters)
}
