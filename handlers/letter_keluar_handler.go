package handlers

import (
	"TugasAkhir/middleware"
	"TugasAkhir/models"
	"TugasAkhir/services"
	"TugasAkhir/utils" // Imported for response helpers
	"TugasAkhir/utils/events"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

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
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return utils.Unauthorized(c, "Unauthorized")
	}

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
		return utils.BadRequest(c, "Invalid request body", nil)
	}

	canCreate, _ := h.permService.CanUserCreateLetter(user, req.Scope)
	if !canCreate {
		return utils.Forbidden(c, "Anda tidak memiliki izin membuat surat dengan scope ini")
	}

	if req.AssignedVerifierID == nil {
		return utils.UnprocessableEntity(c, "Surat keluar wajib memilih verifikator (Manajer)", nil)
	}

	letter := models.Letter{
		NomorSurat:         req.NomorSurat,
		JudulSurat:         req.Judul,
		BidangTujuan:       req.Tujuan,
		IsiSurat:           req.IsiSurat,
		JenisSurat:         models.LetterKeluar,
		Scope:              req.Scope,
		Status:             models.StatusDraft, // Default Draft
		CreatedByID:        user.ID,
		AssignedVerifierID: req.AssignedVerifierID,
		FilePath:           req.FilePath,
		Prioritas:          models.PriorityBiasa,
	}

	if err := h.db.Create(&letter).Error; err != nil {
		return utils.InternalServerError(c, "Gagal membuat surat")
	}

	events.LetterEventBus <- events.LetterEvent{
		Type:   events.LetterCreated,
		Letter: letter,
	}

	return utils.Created(c, "Surat keluar berhasil dibuat", letter)
}

// UpdateDraftLetter - Handler untuk edit dan submit draft
func (h *LetterKeluarHandler) UpdateDraftLetter(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	// 1. Ambil Data Awal
	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return utils.NotFound(c, "Letter not found")
	}

	if letter.CreatedByID != user.ID {
		return utils.Forbidden(c, "Bukan surat Anda")
	}

	if letter.Status != models.StatusDraft && letter.Status != models.StatusPerluRevisi {
		return utils.Conflict(c, "Hanya surat Draft atau Revisi yang bisa diedit")
	}

	oldStatus := letter.Status

	var req struct {
		Judul      string `json:"judul_surat"`
		IsiSurat   string `json:"isi_surat"`
		FilePath   string `json:"file_path"`
		VerifierID *uint  `json:"assigned_verifier_id"`
	}
	c.BodyParser(&req)

	if req.Judul != "" {
		letter.JudulSurat = req.Judul
	}
	if req.IsiSurat != "" {
		letter.IsiSurat = req.IsiSurat
	}
	if req.FilePath != "" {
		letter.FilePath = req.FilePath
	}

	statusChanged := false

	// 2. Logic Submit (Draft -> Perlu Verifikasi)
	if req.VerifierID != nil {
		if letter.Status == models.StatusDraft || letter.Status == models.StatusPerluRevisi {
			letter.Status = models.StatusPerluVerifikasi
			letter.AssignedVerifierID = req.VerifierID
			statusChanged = true
		}
	}

	// 3. Simpan ke DB
	if err := h.db.Save(letter).Error; err != nil {
		return utils.InternalServerError(c, "Gagal menyimpan perubahan")
	}

	// 4. Kirim Event (JIKA STATUS BERUBAH)
	if statusChanged {
		var freshLetter models.Letter
		if err := h.db.Preload("AssignedVerifier").First(&freshLetter, letter.ID).Error; err == nil {
			// Kirim data yang sudah lengkap (freshLetter) ke event bus
			events.LetterEventBus <- events.LetterEvent{
				Type:      events.LetterStatusMoved,
				Letter:    freshLetter,
				OldStatus: oldStatus,
			}
		}
	}

	return utils.OK(c, "Draft berhasil diperbarui", letter)
}

// VerifyLetterApprove
func (h *LetterKeluarHandler) VerifyLetterApprove(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return utils.NotFound(c, "Letter not found")
	}

	canVerify, _ := h.permService.CanUserVerifyLetter(user, letter)
	if !canVerify {
		return utils.Forbidden(c, "Anda tidak memiliki izin memverifikasi surat ini")
	}

	oldStatus := letter.Status
	letter.Status = models.StatusPerluPersetujuan
	letter.VerifiedByID = &user.ID
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return utils.OK(c, "Surat berhasil diverifikasi", nil)
}

// VerifyLetterReject
func (h *LetterKeluarHandler) VerifyLetterReject(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return utils.NotFound(c, "Letter not found")
	}

	canVerify, _ := h.permService.CanUserVerifyLetter(user, letter)
	if !canVerify {
		return utils.Forbidden(c, "Forbidden")
	}

	oldStatus := letter.Status
	letter.Status = models.StatusPerluRevisi
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return utils.OK(c, "Surat dikembalikan untuk revisi", nil)
}

// ApproveLetterByDirektur
func (h *LetterKeluarHandler) ApproveLetterByDirektur(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return utils.NotFound(c, "Not found")
	}

	canApprove, _ := h.permService.CanUserApproveLetter(user, letter)
	if !canApprove {
		return utils.Forbidden(c, "Forbidden")
	}

	oldStatus := letter.Status
	letter.Status = models.StatusDisetujui
	letter.DisposedByID = &user.ID
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return utils.OK(c, "Surat berhasil disetujui", nil)
}

// RejectLetterByDirektur
func (h *LetterKeluarHandler) RejectLetterByDirektur(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	letterID, _ := c.ParamsInt("id")

	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return utils.NotFound(c, "Not found")
	}

	canApprove, _ := h.permService.CanUserApproveLetter(user, letter)
	if !canApprove {
		return utils.Forbidden(c, "Forbidden")
	}

	oldStatus := letter.Status
	letter.Status = models.StatusPerluRevisi
	h.db.Save(letter)

	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return utils.OK(c, "Surat ditolak", nil)
}

// ArchiveLetter
func (h *LetterKeluarHandler) ArchiveLetter(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
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

	return utils.OK(c, "Surat diarsipkan", nil)
}

// GetAvailableVerifiers
func (h *LetterKeluarHandler) GetAvailableVerifiers(c *fiber.Ctx) error {
	scope := c.Query("scope")
	var verifiers []models.User
	query := h.db.Model(&models.User{})

	if strings.EqualFold(scope, models.ScopeEksternal) {
		query = query.Where("role IN ?", []models.Role{models.RoleManajerKPP, models.RoleManajerPemas})
	} else if strings.EqualFold(scope, models.ScopeInternal) {
		query = query.Where("role = ?", models.RoleManajerPKL)
	} else {
		query = query.Where("role IN ?", []models.Role{models.RoleManajerKPP, models.RoleManajerPemas, models.RoleManajerPKL})
	}

	query.Select("id, username, email, role, jabatan").Find(&verifiers)
	return utils.OK(c, "Data verifikator berhasil diambil", verifiers)
}

// GetMyLetters
func (h *LetterKeluarHandler) GetMyLetters(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	var letters []models.Letter
	h.db.Where("created_by_id = ? AND jenis_surat = ?", user.ID, models.LetterKeluar).Order("updated_at DESC").Find(&letters)
	return utils.OK(c, "List surat saya berhasil diambil", letters)
}

// GetLettersNeedVerification (FIXED)
func (h *LetterKeluarHandler) GetLettersNeedVerification(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	var letters []models.Letter

	h.db.Where("status = ? AND assigned_verifier_id = ?", models.StatusPerluVerifikasi, user.ID).
		Preload("CreatedBy").
		Order("created_at ASC").
		Find(&letters)

	return utils.OK(c, "List surat perlu verifikasi berhasil diambil", letters)
}

// GetLettersNeedApproval
func (h *LetterKeluarHandler) GetLettersNeedApproval(c *fiber.Ctx) error {
	var letters []models.Letter
	h.db.Where("status = ? AND jenis_surat = ?", models.StatusPerluPersetujuan, models.LetterKeluar).Preload("CreatedBy").Preload("VerifiedBy").Order("created_at ASC").Find(&letters)
	return utils.OK(c, "List surat perlu persetujuan berhasil diambil", letters)
}
