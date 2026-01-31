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

type LetterKeluarHandler struct {
	db          *gorm.DB
	permService *services.PermissionService
}
type VerifierResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Jabatan  string `json:"jabatan"`
}

func NewLetterKeluarHandler(db *gorm.DB) *LetterKeluarHandler {
	return &LetterKeluarHandler{
		db:          db,
		permService: services.NewPermissionService(db),
	}
}

// CreateSuratKeluar - Handler untuk membuat surat keluar baru
func (h *LetterKeluarHandler) CreateSuratKeluar(c *fiber.Ctx) error {
	// 1. Ambil User dari Context (Token JWT)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return utils.Unauthorized(c, "Unauthorized")
	}

	// 2. Parsing Form Data (Text Fields)
	// Fiber akan mencocokkan field form-data dengan tag `form:"..."` di struct DTO
	var req letters.CreateLetterKeluarRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format data tidak valid", err.Error())
	}

	// 3. Validasi Input (Empty fields dll)
	if errMap := req.Validate(); len(errMap) > 0 {
		return utils.BadRequest(c, "Validasi gagal", errMap)
	}

	// 4. Cek Permission
	// Memastikan Staf Program boleh buat Eksternal, Staf Lembaga boleh buat Internal
	canCreate, _ := h.permService.CanUserCreateLetter(user, req.Scope, models.LetterKeluar)
	if !canCreate {
		return utils.Forbidden(c, "Anda tidak memiliki izin membuat surat keluar dengan scope ini")
	}

	// Tentukan mode: Draft atau Submit berdasarkan field Status
	// Jika Status kosong atau "draft" → mode draft
	// Jika Status "perlu_verifikasi" → mode submit
	isDraftMode := req.Status == "" || req.Status == models.StatusDraft

	// 5. Handle File Upload (Wajib HANYA jika bukan draft)
	var uploadedPath string
	fileHeader, err := c.FormFile("file") // "file" adalah key di form-data

	if isDraftMode {
		// MODE DRAFT: File opsional
		if err == nil {
			// User upload file meskipun draft
			ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
			if ext != ".pdf" && ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
				return utils.BadRequest(c, "Format file harus PDF atau Gambar", nil)
			}
			filename := fmt.Sprintf("surat/draft_%d%s", time.Now().UnixNano(), ext)
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
		filename := fmt.Sprintf("surat/keluar_%d%s", time.Now().UnixNano(), ext)
		uploadedPath, err = storage.UploadFile(c.Context(), fileHeader, filename)
		if err != nil {
			return utils.InternalServerError(c, "Gagal mengupload file ke server")
		}
	}

	// 6. Logic Penentuan Verifikator (Auto-Assign vs Manual) - HANYA jika bukan draft
	var verifierID *uint

	if !isDraftMode {
		if strings.EqualFold(req.Scope, models.ScopeInternal) {
			// === LOGIC OTOMATIS (Internal) ===
			// Sistem mencari sendiri siapa Manajer PKL-nya
			var manajerPKL models.User
			if err := h.db.Where("role = ?", models.RoleManajerPKL).First(&manajerPKL).Error; err != nil {
				return utils.InternalServerError(c, "Sistem Gagal: Manajer PKL belum terdaftar di sistem")
			}
			verifierID = &manajerPKL.ID

		} else {
			// === LOGIC MANUAL (Eksternal) ===
			// User (Staf Program) wajib memilih verifikator dari dropdown
			if req.AssignedVerifierID == nil {
				return utils.UnprocessableEntity(c, "Untuk surat Eksternal, Anda wajib memilih Verifikator (Manajer)", nil)
			}
			verifierID = req.AssignedVerifierID
		}
	}

	// 6.5 Validasi Reply Linking (Opsional) + STATUS UPDATE
	// Jika user menyertakan InReplyToID, validasi bahwa surat induk ada dan perlu balasan
	if req.InReplyToID != nil {
		var parentLetter models.Letter
		// Load parent letter with transactional lock if possible, or just standard load
		if err := h.db.First(&parentLetter, *req.InReplyToID).Error; err != nil {
			return utils.BadRequest(c, "Surat yang akan dibalas tidak ditemukan", nil)
		}
		if !parentLetter.IsSuratMasuk() {
			return utils.BadRequest(c, "Hanya surat masuk yang dapat dibalas", nil)
		}
		// Cek apakah status sudah diarsipkan (artinya sudah dibalas)
		// User report: unlimited reply bug. Fix: check status.
		if parentLetter.Status == models.StatusDiarsipkan {
			return utils.BadRequest(c, "Surat ini sudah dibalas / diarsipkan", nil)
		}
		if !parentLetter.NeedsReply {
			return utils.BadRequest(c, "Surat masuk ini tidak ditandai perlu balasan", nil)
		}

		// [FIX] Update Status Surat Induk menjadi 'diarsipkan' agar tidak muncul lagi di list 'needs-reply'
		// Kita lakukan update ini nanti di dalam transaction creation
	}

	// 7. Mapping ke Model
	letter := req.ToModel()
	letter.JenisSurat = models.LetterKeluar

	// Set status berdasarkan mode
	if isDraftMode {
		letter.Status = models.StatusDraft
	} else {
		// Status langsung ke 'perlu_verifikasi' karena file sudah ada & verifikator sudah ditentukan
		letter.Status = models.StatusPerluVerifikasi
	}

	letter.CreatedByID = user.ID
	letter.AssignedVerifierID = verifierID
	letter.FilePath = uploadedPath

	// Default Prioritas jika kosong
	if letter.Prioritas == "" {
		letter.Prioritas = models.PriorityBiasa
	}

	// 8. Simpan ke Database (Transactional with Auto-Increment Nomor Agenda)
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Only generate nomor agenda if NOT draft mode
		if !isDraftMode {
			nomorAgenda, err := utils.GenerateNomorAgenda(tx, models.LetterKeluar)
			if err != nil {
				return err
			}
			letter.NomorAgenda = nomorAgenda
		}
		// Draft letters will have empty nomor_agenda

		if err := tx.Create(&letter).Error; err != nil {
			return err
		}

		// [FIX] Jika ini adalah balasan (InReplyToID != nil), update status surat induk
		if letter.InReplyToID != nil {
			if err := tx.Model(&models.Letter{}).
				Where("id = ?", *letter.InReplyToID).
				Update("status", models.StatusDiarsipkan).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return utils.InternalServerError(c, "Gagal menyimpan data surat: "+err.Error())
	}

	// 9. Kirim Event Notifikasi (HANYA jika bukan draft)
	if !isDraftMode {
		// Kita perlu preload data verifier (nama/role) agar notifikasi di consumer lengkap
		h.db.Preload("AssignedVerifier").First(&letter, letter.ID)

		events.LetterEventBus <- events.LetterEvent{
			Type:   events.LetterCreated,
			Letter: letter,
		}
		AddPresignedURLToLetter(&letter)
		return utils.Created(c, "Surat keluar berhasil dibuat dan diteruskan ke Verifikator", letter)
	}

	AddPresignedURLToLetter(&letter)
	return utils.Created(c, "Draft surat berhasil disimpan", letter)
}

// UpdateDraftLetter - Handler untuk edit dan submit draft
func (h *LetterKeluarHandler) UpdateDraftLetter(c *fiber.Ctx) error {
	// 1. Ambil User
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		return utils.Unauthorized(c, "Unauthorized")
	}

	letterID, _ := c.ParamsInt("id")

	// 2. Ambil Data Surat Lama dari DB
	letter, err := h.permService.GetLetterByID(uint(letterID))
	if err != nil {
		return utils.NotFound(c, "Surat tidak ditemukan")
	}

	// 3. Validasi Kepemilikan & Status
	if letter.CreatedByID != user.ID {
		return utils.Forbidden(c, "Anda tidak berhak mengedit surat ini")
	}
	// Hanya boleh edit jika masih Draft atau sedang Diminta Revisi
	if letter.Status != models.StatusDraft && letter.Status != models.StatusPerluRevisi {
		return utils.Conflict(c, "Hanya surat status Draft atau Perlu Revisi yang bisa diedit")
	}

	oldStatus := letter.Status

	// 4. Parsing Form Data (Gunakan UpdateLetterRequest dari DTO)
	var req letters.UpdateLetterKeluarRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Format data tidak valid", err.Error())
	}

	// Validasi input
	if errMap := req.Validate(); len(errMap) > 0 {
		return utils.BadRequest(c, "Validasi gagal", errMap)
	}

	// 5. Update Field Teks (Judul, Nomor, dll)
	// Fungsi ApplyUpdate ini ada di dto/letters/mapper.go
	letters.ApplyUpdate(letter, &req)

	// 6. Handle File Upload (OPSIONAL untuk Edit)
	// Jika user mengupload file baru, kita ganti. Jika tidak, pakai file lama.
	fileHeader, err := c.FormFile("file")
	if err == nil {
		// Validasi Ekstensi
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext != ".pdf" && ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			return utils.BadRequest(c, "Format file revisi harus PDF atau Gambar", nil)
		}

		// Upload File Baru
		filename := fmt.Sprintf("surat/keluar_%d%s", time.Now().UnixNano(), ext)
		uploadedPath, err := storage.UploadFile(c.Context(), fileHeader, filename)
		if err != nil {
			return utils.InternalServerError(c, "Gagal mengupload file revisi")
		}

		// Hapus file lama dari storage jika perlu (Optional, untuk hemat storage)
		// storage.DeleteFile(c.Context(), letter.FilePath)

		// Update path di database
		letter.FilePath = uploadedPath
	}

	// 7. Logic Auto-Assign Manajer (Sama seperti Create)
	// Kita cek Scope surat saat ini (apakah berubah atau tetap)
	scopeToCheck := letter.Scope

	statusChanged := false

	if strings.EqualFold(scopeToCheck, models.ScopeInternal) {
		// === LOGIC OTOMATIS (Internal) ===
		// Cari Manajer PKL & Auto Assign
		var manajerPKL models.User
		if err := h.db.Where("role = ?", models.RoleManajerPKL).First(&manajerPKL).Error; err == nil {
			letter.AssignedVerifierID = &manajerPKL.ID

			// Jika user menekan tombol "Simpan & Ajukan" (biasanya ditandai dengan perubahan status/intent)
			// Di sini kita asumsikan setiap edit pada status 'Revisi' akan mengajukan ulang ke 'Perlu Verifikasi'
			if letter.Status == models.StatusDraft || letter.Status == models.StatusPerluRevisi {
				letter.Status = models.StatusPerluVerifikasi
				statusChanged = true
			}
		}
	} else {
		// === LOGIC MANUAL (Eksternal) ===
		// Jika user memilih verifikator baru di dropdown
		if req.AssignedVerifierID != nil {
			letter.AssignedVerifierID = req.AssignedVerifierID
			// Auto ajukan ulang
			if letter.Status == models.StatusDraft || letter.Status == models.StatusPerluRevisi {
				letter.Status = models.StatusPerluVerifikasi
				statusChanged = true
			}
		} else if letter.AssignedVerifierID != nil {
			// Jika user TIDAK ganti verifikator, tapi surat sedang Revisi,
			// Kita anggap dia mengajukan ulang ke orang yang sama.
			if letter.Status == models.StatusPerluRevisi {
				letter.Status = models.StatusPerluVerifikasi
				statusChanged = true
			}
		}
	}

	// 8. Simpan Perubahan ke DB (with nomor_agenda generation if needed)
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Generate nomor_agenda ONLY when transitioning draft→publish AND nomor_agenda is empty
		if oldStatus == models.StatusDraft && letter.Status == models.StatusPerluVerifikasi && letter.NomorAgenda == "" {
			nomorAgenda, err := utils.GenerateNomorAgenda(tx, models.LetterKeluar)
			if err != nil {
				return err
			}
			letter.NomorAgenda = nomorAgenda
		}

		return tx.Save(letter).Error
	})

	if err != nil {
		return utils.InternalServerError(c, "Gagal menyimpan revisi surat: "+err.Error())
	}

	// 9. Kirim Notifikasi (Hanya jika status berubah, misal: Revisi -> Perlu Verifikasi)
	if statusChanged {
		var freshLetter models.Letter
		if err := h.db.Preload("AssignedVerifier").First(&freshLetter, letter.ID).Error; err == nil {
			events.LetterEventBus <- events.LetterEvent{
				Type:      events.LetterStatusMoved,
				Letter:    freshLetter,
				OldStatus: oldStatus,
			}
		}
	}

	AddPresignedURLToLetter(letter)
	return utils.OK(c, "Surat berhasil diperbarui dan diajukan kembali", letter)
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
		return utils.NotFound(c, "Surat tidak ditemukan")
	}

	canApprove, _ := h.permService.CanUserApproveLetter(user, letter)
	if !canApprove {
		return utils.Forbidden(c, "Anda tidak memiliki izin menyetujui surat ini")
	}

	oldStatus := letter.Status

	// [PERUBAHAN DISINI]
	// Real Case: Langsung "Diarsipkan", bukan "Disetujui" dulu.
	letter.Status = models.StatusDiarsipkan
	letter.DisposedByID = &user.ID // DisposedBy diisi Direktur sebagai tanda approval

	if err := h.db.Save(letter).Error; err != nil {
		return utils.InternalServerError(c, "Gagal memproses persetujuan surat")
	}

	// Kirim Event Notifikasi
	events.LetterEventBus <- events.LetterEvent{
		Type:      events.LetterStatusMoved,
		Letter:    *letter,
		OldStatus: oldStatus,
	}

	return utils.OK(c, "Surat berhasil disetujui dan otomatis diarsipkan", nil)
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

	var response []VerifierResponse
	for _, v := range verifiers {
		response = append(response, VerifierResponse{
			ID:       v.ID,
			Username: v.Username,
			Email:    v.Email,
			Role:     string(v.Role),
			Jabatan:  v.Jabatan,
		})
	}
	return utils.OK(c, "Data verifikator berhasil diambil", verifiers)
}

// GetMyLetters
func (h *LetterKeluarHandler) GetMyLetters(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	var letters []models.Letter

	// Staf Lembaga bisa melihat SEMUA surat keluar (sebagai arsiparis)
	// Staf lain hanya melihat surat buatannya sendiri
	if user.Role == models.RoleStafLembaga {
		h.db.Where("jenis_surat = ?", models.LetterKeluar).
			Preload("CreatedBy").
			Order("updated_at DESC").
			Find(&letters)
	} else {
		h.db.Where("created_by_id = ? AND jenis_surat = ?", user.ID, models.LetterKeluar).
			Preload("CreatedBy").
			Order("updated_at DESC").
			Find(&letters)
	}

	AddPresignedURLsToLetters(letters)
	return utils.OK(c, "List surat keluar berhasil diambil", letters)
}

// GetLettersNeedVerification (FIXED)
func (h *LetterKeluarHandler) GetLettersNeedVerification(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	var letters []models.Letter

	h.db.Where("status = ? AND assigned_verifier_id = ?", models.StatusPerluVerifikasi, user.ID).
		Preload("CreatedBy").
		Order("created_at ASC").
		Find(&letters)

	AddPresignedURLsToLetters(letters)
	return utils.OK(c, "List surat perlu verifikasi berhasil diambil", letters)
}

// GetLettersNeedApproval
func (h *LetterKeluarHandler) GetLettersNeedApproval(c *fiber.Ctx) error {
	var letters []models.Letter
	h.db.Where("status = ? AND jenis_surat = ?", models.StatusPerluPersetujuan, models.LetterKeluar).Preload("CreatedBy").Preload("VerifiedBy").Order("created_at ASC").Find(&letters)
	AddPresignedURLsToLetters(letters)
	return utils.OK(c, "List surat perlu persetujuan berhasil diambil", letters)
}

// GetMyApprovals - Direktur melihat riwayat surat keluar yang sudah di-approve
func (h *LetterKeluarHandler) GetMyApprovals(c *fiber.Ctx) error {
	user, _ := middleware.GetUserFromContext(c)
	var letters []models.Letter

	// Surat keluar yang sudah di-approve oleh direktur ini (disposed_by_id = user.ID)
	h.db.Where("disposed_by_id = ? AND jenis_surat = ?", user.ID, models.LetterKeluar).
		Preload("CreatedBy").
		Preload("VerifiedBy").
		Order("updated_at DESC").
		Find(&letters)

	AddPresignedURLsToLetters(letters)
	return utils.OK(c, "Riwayat surat keluar yang sudah disetujui", letters)
}
