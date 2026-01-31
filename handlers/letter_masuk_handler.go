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

	// Tentukan mode: Draft atau Submit berdasarkan field Status
	// Jika Status kosong atau "draft" → mode draft
	// Jika Status "belum_disposisi" → mode submit (kirim ke Direktur)
	isDraftMode := req.Status == "" || req.Status == models.StatusDraft

	// 4. Handle File Upload (Wajib HANYA jika bukan draft)
	var uploadedPath string
	fileHeader, err := c.FormFile("file")

	if isDraftMode {
		// MODE DRAFT: File opsional
		if err == nil {
			// User upload file meskipun draft
			ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
			if ext != ".pdf" && ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
				return utils.BadRequest(c, "Format file harus PDF atau Gambar", nil)
			}
			filename := fmt.Sprintf("surat/draft_masuk_%d%s", time.Now().UnixNano(), ext)
			uploadedPath, err = storage.UploadFile(c.Context(), fileHeader, filename)
			if err != nil {
				return utils.InternalServerError(c, "Gagal mengupload file ke server")
			}
		}
	} else {
		// MODE SUBMIT: File wajib
		if err != nil {
			return utils.BadRequest(c, "File surat wajib diunggah untuk mengirim surat", nil)
		}
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext != ".pdf" && ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			return utils.BadRequest(c, "Format file harus PDF atau Gambar", nil)
		}
		filename := fmt.Sprintf("surat/masuk_%d%s", time.Now().UnixNano(), ext)
		uploadedPath, err = storage.UploadFile(c.Context(), fileHeader, filename)
		if err != nil {
			return utils.InternalServerError(c, "Gagal mengupload file ke server")
		}
	}

	// 5. Mapping ke Model
	letter := req.ToModel(user.ID, uploadedPath)
	letter.JenisSurat = models.LetterMasuk

	// Set status berdasarkan mode
	if isDraftMode {
		letter.Status = models.StatusDraft
	} else {
		// Status langsung ke 'belum_disposisi' untuk dikirim ke Direktur
		letter.Status = models.StatusBelumDisposisi
	}

	// Auto-Generate Nomor Agenda (Transactional)
	// Only generate for non-draft letters (published/submitted)
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Only generate nomor agenda if NOT draft mode
		if !isDraftMode {
			nomorAgenda, err := utils.GenerateNomorAgenda(tx, models.LetterMasuk)
			if err != nil {
				return err
			}
			letter.NomorAgenda = nomorAgenda
		}
		// Draft letters will have empty nomor_agenda

		if err := tx.Create(&letter).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return utils.InternalServerError(c, "Gagal mencatat surat masuk: "+err.Error())
	}

	// 6. Kirim Notifikasi (HANYA jika bukan draft)
	if !isDraftMode {
		events.LetterEventBus <- events.LetterEvent{
			Type:   events.LetterCreated,
			Letter: letter,
		}
		AddPresignedURLToLetter(&letter)
		return utils.Created(c, "Surat masuk berhasil dicatat dan dikirim ke Direktur", letter)
	}

	AddPresignedURLToLetter(&letter)
	return utils.Created(c, "Draft surat masuk berhasil disimpan", letter)
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

	// Cek Status (Boleh edit jika Draft atau Belum Disposisi)
	if letter.Status != models.StatusDraft && letter.Status != models.StatusBelumDisposisi {
		return utils.Conflict(c, "Surat yang sudah didisposisi tidak dapat diedit")
	}

	oldStatus := letter.Status

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

	// Handle File Upload (Optional Replace, WAJIB jika submit draft)
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

	// Jika status dikirim "belum_disposisi" DAN surat masih draft, maka ini adalah submission
	isSubmitting := req.Status != nil && *req.Status == models.StatusBelumDisposisi

	// [FIX] Gunakan oldStatus untuk cek kondisi draft
	if isSubmitting && oldStatus == models.StatusDraft {
		// File wajib ada untuk submit
		if letter.FilePath == "" {
			return utils.BadRequest(c, "File surat wajib diunggah untuk mengirim surat", nil)
		}
		letter.Status = models.StatusBelumDisposisi
	}

	// Use transaction for save (with nomor_agenda generation if needed)
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Generate nomor_agenda ONLY when transitioning draft→publish AND nomor_agenda is empty
		if oldStatus == models.StatusDraft && letter.Status == models.StatusBelumDisposisi && letter.NomorAgenda == "" {
			nomorAgenda, err := utils.GenerateNomorAgenda(tx, models.LetterMasuk)
			if err != nil {
				return err
			}
			letter.NomorAgenda = nomorAgenda
		}

		return tx.Save(letter).Error
	})

	if err != nil {
		return utils.InternalServerError(c, "Gagal menyimpan surat: "+err.Error())
	}

	// Kirim notifikasi jika status berubah dari draft ke belum_disposisi
	if oldStatus == models.StatusDraft && letter.Status == models.StatusBelumDisposisi {
		events.LetterEventBus <- events.LetterEvent{
			Type:      events.LetterStatusMoved,
			Letter:    *letter,
			OldStatus: oldStatus,
		}
		AddPresignedURLToLetter(letter)
		return utils.OK(c, "Draft surat berhasil dikirim ke Direktur", letter)
	}

	AddPresignedURLToLetter(letter)
	return utils.OK(c, "Surat masuk berhasil diperbarui", letter)
}

// GetMySuratMasuk - List surat masuk buatan saya
func (h *LetterMasukHandler) GetMySuratMasuk(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return utils.Unauthorized(c, "Unauthorized")
	}

	var letters []models.Letter

	// Staf Lembaga bisa melihat SEMUA surat masuk (sebagai arsiparis)
	// Staf lain hanya melihat surat buatannya sendiri
	if user.Role == models.RoleStafLembaga {
		h.db.Where("jenis_surat = ?", models.LetterMasuk).
			Preload("CreatedBy").
			Order("created_at DESC").
			Find(&letters)
	} else {
		h.db.Where("created_by_id = ? AND jenis_surat = ?", user.ID, models.LetterMasuk).
			Preload("CreatedBy").
			Order("created_at DESC").
			Find(&letters)
	}

	AddPresignedURLsToLetters(letters)
	return utils.OK(c, "List surat masuk berhasil diambil", letters)
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
		NeedsReply         bool   `json:"needs_reply"` // Flag: apakah surat ini butuh balasan?
	}
	c.BodyParser(&req)

	oldStatus := letter.Status
	now := time.Now()

	// Update Data Disposisi
	letter.Disposisi = req.InstruksiDisposisi
	letter.BidangTujuan = req.TujuanDisposisi
	letter.DisposedByID = &user.ID
	letter.TanggalDisposisi = &now
	letter.NeedsReply = req.NeedsReply // Set flag needs_reply

	if req.Catatan != "" {
		letter.Disposisi += " | Catatan: " + req.Catatan
	}

	// Tentukan status berdasarkan needs_reply:
	// - needs_reply = true  → sudah_disposisi (menunggu surat keluar balasan)
	// - needs_reply = false → diarsipkan (tidak perlu tindakan lanjutan)
	if req.NeedsReply {
		letter.Status = models.StatusSudahDisposisi
	} else {
		letter.Status = models.StatusDiarsipkan
	}

	h.db.Save(letter)

	// Notif ke Staf Pembuat
	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	AddPresignedURLToLetter(letter)
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

	AddPresignedURLsToLetters(letters)
	return utils.OK(c, "List disposisi berhasil diambil", letters)
}

// GetMyDispositions - Direktur melihat riwayat surat masuk yang sudah didisposisi
func (h *LetterMasukHandler) GetMyDispositions(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	var letters []models.Letter

	// Surat masuk yang sudah didisposisi oleh direktur ini (disposed_by_id = user.ID)
	h.db.Where("disposed_by_id = ? AND jenis_surat = ?", user.ID, models.LetterMasuk).
		Preload("CreatedBy").
		Order("updated_at DESC").
		Find(&letters)

	AddPresignedURLsToLetters(letters)
	return utils.OK(c, "Riwayat surat masuk yang sudah didisposisi", letters)
}

// GetLettersNeedingReply - List surat masuk yang butuh balasan (needs_reply = true)
// Filter by scope: Staf Program → Eksternal, Staf Lembaga → Internal
func (h *LetterMasukHandler) GetLettersNeedingReply(c *fiber.Ctx) error {
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return utils.Unauthorized(c, "Unauthorized")
	}

	// Tentukan scope berdasarkan role staf
	var scopeFilter string
	switch user.Role {
	case models.RoleStafProgram:
		// [CHANGE] Staf Program should see ALL letters needing reply, not just External
		// scopeFilter = models.ScopeEksternal
		scopeFilter = ""
	case models.RoleStafLembaga:
		scopeFilter = models.ScopeInternal
	default:
		// Role lain (admin, direktur, manajer) bisa lihat semua
		scopeFilter = ""
	}

	var letters []models.Letter

	// Build query
	// [FIX] Exclude letter that already archived (already replied)
	query := h.db.Where("jenis_surat = ? AND needs_reply = ? AND status != ?", models.LetterMasuk, true, models.StatusDiarsipkan)

	// Apply scope filter jika staf
	if scopeFilter != "" {
		query = query.Where("scope = ?", scopeFilter)
	}

	query.Preload("CreatedBy").
		Preload("DisposedBy").
		Preload("Replies"). // Preload surat balasan jika ada
		Order("updated_at DESC").
		Find(&letters)

	// Filter hanya yang belum punya balasan
	var needsReply []models.Letter
	for _, l := range letters {
		if len(l.Replies) == 0 {
			needsReply = append(needsReply, l)
		}
	}

	AddPresignedURLsToLetters(needsReply)
	return utils.OK(c, "List surat masuk yang butuh balasan", needsReply)
}
