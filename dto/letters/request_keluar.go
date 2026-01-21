package letters

import (
	"strings"
	"time"

	"TugasAkhir/models"
)

// CreateLetterKeluarRequest - Req khusus untuk surat keluar
type CreateLetterKeluarRequest struct {
	// Tambahkan tag `form` di sebelah tag `json`
	Pengirim         string            `json:"pengirim" form:"pengirim"`
	NomorSurat       string            `json:"nomor_surat" form:"nomor_surat"`
	NomorAgenda      string            `json:"nomor_agenda" form:"nomor_agenda"`
	Disposisi        string            `json:"disposisi" form:"disposisi"`
	TanggalDisposisi *time.Time        `json:"tanggal_disposisi" form:"tanggal_disposisi"`
	BidangTujuan     string            `json:"bidang_tujuan" form:"bidang_tujuan"`
	JenisSurat       models.LetterType `json:"jenis_surat" form:"jenis_surat"`
	Prioritas        models.Priority   `json:"prioritas" form:"prioritas"`
	IsiSurat         string            `json:"isi_surat" form:"isi_surat"`
	TanggalSurat     *time.Time        `json:"tanggal_surat" form:"tanggal_surat"`
	TanggalMasuk     *time.Time        `json:"tanggal_masuk" form:"tanggal_masuk"`
	JudulSurat       string            `json:"judul_surat" form:"judul_surat"`
	Kesimpulan       string            `json:"kesimpulan" form:"kesimpulan"`

	// FilePath tidak perlu `form` karena akan diisi manual oleh handler setelah upload sukses
	FilePath string `json:"file_path"`

	// Status: kosong/"draft" = simpan sebagai draft, "perlu_verifikasi" = kirim ke Manajer
	Status models.LetterStatus `json:"status" form:"status"`

	// Gunakan pointer agar opsional, tambahkan form tag
	CreatedByID  *uint `json:"created_by_id" form:"created_by_id"`
	VerifiedByID *uint `json:"verified_by_id" form:"verified_by_id"`
	DisposedByID *uint `json:"disposed_by_id" form:"disposed_by_id"`

	// Field Scope Wajib Ada Form Tag
	Scope string `json:"scope" form:"scope"`

	// Khusus Verifier ID (User input dari dropdown)
	AssignedVerifierID *uint `json:"assigned_verifier_id" form:"assigned_verifier_id"`
}

type UpdateLetterKeluarRequest struct {
	Pengirim         *string              `json:"pengirim" form:"pengirim"`
	NomorSurat       *string              `json:"nomor_surat" form:"nomor_surat"`
	NomorAgenda      *string              `json:"nomor_agenda" form:"nomor_agenda"`
	Disposisi        *string              `json:"disposisi" form:"disposisi"`
	TanggalDisposisi *time.Time           `json:"tanggal_disposisi" form:"tanggal_disposisi"`
	BidangTujuan     *string              `json:"bidang_tujuan" form:"bidang_tujuan"`
	JenisSurat       *models.LetterType   `json:"jenis_surat" form:"jenis_surat"`
	Prioritas        *models.Priority     `json:"prioritas" form:"prioritas"`
	IsiSurat         *string              `json:"isi_surat" form:"isi_surat"`
	TanggalSurat     *time.Time           `json:"tanggal_surat" form:"tanggal_surat"`
	TanggalMasuk     *time.Time           `json:"tanggal_masuk" form:"tanggal_masuk"`
	JudulSurat       *string              `json:"judul_surat" form:"judul_surat"`
	Kesimpulan       *string              `json:"kesimpulan" form:"kesimpulan"`
	FilePath         *string              `json:"file_path"` // Diisi manual handler
	Status           *models.LetterStatus `json:"status" form:"status"`

	CreatedByID        *uint `json:"created_by_id" form:"created_by_id"`
	VerifiedByID       *uint `json:"verified_by_id" form:"verified_by_id"`
	DisposedByID       *uint `json:"disposed_by_id" form:"disposed_by_id"`
	AssignedVerifierID *uint `json:"assigned_verifier_id" form:"assigned_verifier_id"`
}

func (r *CreateLetterKeluarRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(r.Pengirim) == "" {
		errors["pengirim"] = "pengirim is required"
	}
	if strings.TrimSpace(r.NomorSurat) == "" {
		errors["nomor_surat"] = "nomor_surat is required"
	}
	if strings.TrimSpace(r.JudulSurat) == "" {
		errors["judul_surat"] = "judul_surat is required"
	}
	if !isValidLetterType(r.JenisSurat) {
		errors["jenis_surat"] = "jenis_surat must be masuk, keluar, or internal"
	}
	// Validasi Prioritas Opsional (default biasa)
	if r.Prioritas != "" && !isValidPriority(r.Prioritas) {
		errors["prioritas"] = "prioritas must be biasa, segera, or penting"
	}
	// Status Opsional
	if r.Status != "" && !isValidLetterStatus(r.Status) {
		errors["status"] = "status invalid"
	}

	return errors
}

func (r *UpdateLetterKeluarRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if r.JenisSurat != nil && !isValidLetterType(*r.JenisSurat) {
		errors["jenis_surat"] = "jenis_surat must be masuk, keluar, or internal"
	}
	if r.Prioritas != nil && !isValidPriority(*r.Prioritas) {
		errors["prioritas"] = "prioritas must be biasa, segera, or penting"
	}
	if r.Status != nil && !isValidLetterStatus(*r.Status) {
		errors["status"] = "status invalid"
	}

	return errors
}

func isValidLetterType(t models.LetterType) bool {
	switch t {
	case models.LetterMasuk, models.LetterKeluar, models.LetterInternal:
		return true
	default:
		return false
	}
}

func isValidPriority(p models.Priority) bool {
	switch p {
	case models.PriorityBiasa, models.PrioritySegera, models.PriorityPenting:
		return true
	default:
		return false
	}
}

func isValidLetterStatus(s models.LetterStatus) bool {
	switch s {
	case models.StatusDraft,
		models.StatusPerluVerifikasi,
		models.StatusBelumDisposisi,
		models.StatusSudahDisposisi,
		models.StatusPerluPersetujuan,
		models.StatusPerluRevisi,
		models.StatusDisetujui,
		models.StatusDiarsipkan:
		return true
	default:
		return false
	}
}
