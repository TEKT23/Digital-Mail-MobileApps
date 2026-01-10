package letters

import (
	"TugasAkhir/models"
	"time"
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

	// UPDATE: Ubah jadi uint (bukan pointer) sesuai Models
	CreatedByID uint                `json:"created_by_id"`
	CreatedBy   *LetterUserResponse `json:"created_by"`

	VerifiedByID *uint               `json:"verified_by_id"`
	VerifiedBy   *LetterUserResponse `json:"verified_by"`
	DisposedByID *uint               `json:"disposed_by_id"`
	DisposedBy   *LetterUserResponse `json:"disposed_by"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

type LetterUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Jabatan  string `json:"jabatan"`
}

func toLetterUserResponse(user *models.User) *LetterUserResponse {
	if user == nil {
		return nil
	}
	return &LetterUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		Jabatan:  user.Jabatan,
	}
}

func NewLetterResponse(letter *models.Letter) LetterResponse {
	if letter == nil {
		return LetterResponse{}
	}

	return LetterResponse{
		IDSurat:          letter.ID,
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

		// Assignment langsung (uint ke uint)
		CreatedByID: letter.CreatedByID,
		CreatedBy:   toLetterUserResponse(letter.CreatedBy),

		VerifiedByID: letter.VerifiedByID,
		VerifiedBy:   toLetterUserResponse(letter.VerifiedBy),
		DisposedByID: letter.DisposedByID,
		DisposedBy:   toLetterUserResponse(letter.DisposedBy),
		CreatedAt:    letter.CreatedAt,
		UpdatedAt:    letter.UpdatedAt,
	}
}
