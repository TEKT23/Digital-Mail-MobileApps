package letters

import (
	"time"

	"TugasAkhir/models"
)

type LetterResponse struct {
	IDSurat          uint                `json:"id_surat"`
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
	CreatedByID      *uint               `json:"created_by_id"`
	CreatedBy        *models.User        `json:"created_by"`
	VerifiedByID     *uint               `json:"verified_by_id"`
	VerifiedBy       *models.User        `json:"verified_by"`
	DisposedByID     *uint               `json:"disposed_by_id"`
	DisposedBy       *models.User        `json:"disposed_by"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

func NewLetterResponse(letter *models.Letter) LetterResponse {
	if letter == nil {
		return LetterResponse{}
	}

	return LetterResponse{
		IDSurat:          letter.IDSurat,
		Pengirim:         letter.Pengirim,
		NomorSurat:       letter.NomorSurat,
		NomorAgenda:      letter.NomorAgenda,
		Disposisi:        letter.Disposisi,
		TanggalDisposisi: letter.TanggalDisposisi,
		BidangTujuan:     letter.BidangTujuan,
		JenisSurat:       letter.JenisSurat,
		Prioritas:        letter.Prioritas,
		IsiSurat:         letter.IsiSurat,
		TanggalSurat:     letter.TanggalSurat,
		TanggalMasuk:     letter.TanggalMasuk,
		JudulSurat:       letter.JudulSurat,
		Kesimpulan:       letter.Kesimpulan,
		FilePath:         letter.FilePath,
		Status:           letter.Status,
		CreatedByID:      letter.CreatedByID,
		CreatedBy:        letter.CreatedBy,
		VerifiedByID:     letter.VerifiedByID,
		VerifiedBy:       letter.VerifiedBy,
		DisposedByID:     letter.DisposedByID,
		DisposedBy:       letter.DisposedBy,
		CreatedAt:        letter.CreatedAt,
		UpdatedAt:        letter.UpdatedAt,
	}
}
