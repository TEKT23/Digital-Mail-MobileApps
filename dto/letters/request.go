package letters

import (
	"strings"
	"time"

	"TugasAkhir/models"
)

type CreateLetterRequest struct {
	Pengirim         string              `json:"pengirim"`
	NomorSurat       string              `json:"nomor_surat"`
	NomorAgenda      string              `json:"nomor_agenda"`
	Disposisi        string              `json:"disposisi"`
	TanggalDisposisi *time.Time          `json:"tanggal_disposisi"`
	BidangTujuan     string              `json:"bidang_tujuan"`
	JenisSurat       models.LetterType   `json:"jenis_surat"`
	Prioritas        models.Priority     `json:"prioritas"`
	IsiSurat         string              `json:"isi_surat"`
	TanggalSurat     *time.Time          `json:"tanggal_surat"`
	TanggalMasuk     *time.Time          `json:"tanggal_masuk"`
	JudulSurat       string              `json:"judul_surat"`
	Kesimpulan       string              `json:"kesimpulan"`
	FilePath         string              `json:"file_path"`
	Status           models.LetterStatus `json:"status"`
	// CreatedByID tetap pointer di request agar opsional (bisa nil) saat parsing JSON
	CreatedByID  *uint `json:"created_by_id"`
	VerifiedByID *uint `json:"verified_by_id"`
	DisposedByID *uint `json:"disposed_by_id"`
}

type UpdateLetterRequest struct {
	Pengirim         *string              `json:"pengirim"`
	NomorSurat       *string              `json:"nomor_surat"`
	NomorAgenda      *string              `json:"nomor_agenda"`
	Disposisi        *string              `json:"disposisi"`
	TanggalDisposisi *time.Time           `json:"tanggal_disposisi"`
	BidangTujuan     *string              `json:"bidang_tujuan"`
	JenisSurat       *models.LetterType   `json:"jenis_surat"`
	Prioritas        *models.Priority     `json:"prioritas"`
	IsiSurat         *string              `json:"isi_surat"`
	TanggalSurat     *time.Time           `json:"tanggal_surat"`
	TanggalMasuk     *time.Time           `json:"tanggal_masuk"`
	JudulSurat       *string              `json:"judul_surat"`
	Kesimpulan       *string              `json:"kesimpulan"`
	FilePath         *string              `json:"file_path"`
	Status           *models.LetterStatus `json:"status"`
	CreatedByID      *uint                `json:"created_by_id"`
	VerifiedByID     *uint                `json:"verified_by_id"`
	DisposedByID     *uint                `json:"disposed_by_id"`
}

func (r *CreateLetterRequest) Validate() map[string]string {
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
	if r.Prioritas != "" && !isValidPriority(r.Prioritas) {
		errors["prioritas"] = "prioritas must be biasa, segera, or penting"
	}
	if r.Status != "" && !isValidLetterStatus(r.Status) {
		errors["status"] = "status invalid"
	}

	return errors
}

func (r *UpdateLetterRequest) Validate() map[string]string {
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
	// UPDATE: Sesuaikan dengan status baru di models/letters.entity.go
	case models.StatusDraft,
		models.StatusPerluVerifikasi,
		models.StatusBelumDisposisi,
		models.StatusSudahDisposisi,
		models.StatusPerluPersetujuan,
		models.StatusPerluRevisi,
		models.StatusDisetujui,
		models.StatusDiarsipkan: // Ganti StatusTerkirim jadi StatusDiarsipkan
		return true
	default:
		return false
	}
}
