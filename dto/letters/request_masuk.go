package letters

import (
	"TugasAkhir/models"
	"strings"
)

// CreateLetterMasukRequest - Req khusus untuk surat masuk
type CreateLetterMasukRequest struct {
	NomorAgenda  string `json:"nomor_agenda" form:"nomor_agenda"`
	NomorSurat   string `json:"nomor_surat" form:"nomor_surat"`
	Pengirim     string `json:"pengirim" form:"pengirim"`
	JudulSurat   string `json:"judul_surat" form:"judul_surat"`
	TanggalSurat string `json:"tanggal_surat" form:"tanggal_surat"` // YYYY-MM-DD
	TanggalMasuk string `json:"tanggal_masuk" form:"tanggal_masuk"` // YYYY-MM-DD
	Scope        string `json:"scope" form:"scope"`
	Prioritas    string `json:"prioritas" form:"prioritas"`
	IsiSurat     string `json:"isi_surat" form:"isi_surat"`
	Kesimpulan   string `json:"kesimpulan" form:"kesimpulan"` // Opsional

	// Status: kosong/"draft" = simpan sebagai draft, "belum_disposisi" = kirim ke Direktur
	Status models.LetterStatus `json:"status" form:"status"`

	// Note: FilePath di-handle handler
}

// UpdateLetterMasukRequest - Req untuk edit (hanya field yang relevan)
type UpdateLetterMasukRequest struct {
	NomorAgenda  *string `json:"nomor_agenda" form:"nomor_agenda"`
	NomorSurat   *string `json:"nomor_surat" form:"nomor_surat"`
	Pengirim     *string `json:"pengirim" form:"pengirim"`
	JudulSurat   *string `json:"judul_surat" form:"judul_surat"`
	TanggalSurat *string `json:"tanggal_surat" form:"tanggal_surat"`
	TanggalMasuk *string `json:"tanggal_masuk" form:"tanggal_masuk"`
	Scope        *string `json:"scope" form:"scope"`
	Prioritas    *string `json:"prioritas" form:"prioritas"`
	IsiSurat     *string `json:"isi_surat" form:"isi_surat"`
	Kesimpulan   *string `json:"kesimpulan" form:"kesimpulan"`

	// Status: "belum_disposisi" = submit draft ke Direktur
	Status *models.LetterStatus `json:"status" form:"status"`
}

func (r *CreateLetterMasukRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(r.NomorSurat) == "" {
		errors["nomor_surat"] = "nomor_surat is required"
	}
	if strings.TrimSpace(r.Pengirim) == "" {
		errors["pengirim"] = "pengirim is required"
	}
	if strings.TrimSpace(r.JudulSurat) == "" {
		errors["judul_surat"] = "judul_surat is required"
	}
	if r.Prioritas != "" && !isValidPriorityString(r.Prioritas) {
		errors["prioritas"] = "prioritas must be biasa, segera, or penting"
	}
	if r.Scope != "" {
		// Validasi scope jika perlu
	}

	return errors
}

func (r *UpdateLetterMasukRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.Prioritas != nil && !isValidPriorityString(*r.Prioritas) {
		errors["prioritas"] = "prioritas must be biasa, segera, or penting"
	}
	return errors
}

func isValidPriorityString(p string) bool {
	switch models.Priority(p) {
	case models.PriorityBiasa, models.PrioritySegera, models.PriorityPenting:
		return true
	default:
		return false
	}
}
