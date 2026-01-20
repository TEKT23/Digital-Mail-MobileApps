package letters

import (
	"TugasAkhir/models"
	"strings"
	"time"
)

func (r *CreateLetterMasukRequest) ToModel(userID uint, filePath string) models.Letter {
	// Parse Dates
	var tglSurat, tglMasuk *time.Time
	if r.TanggalSurat != "" {
		if t, err := time.Parse("2006-01-02", r.TanggalSurat); err == nil {
			tglSurat = &t
		}
	}
	if r.TanggalMasuk != "" {
		if t, err := time.Parse("2006-01-02", r.TanggalMasuk); err == nil {
			tglMasuk = &t
		}
	}

	priority := models.PriorityBiasa
	if r.Prioritas != "" {
		priority = models.Priority(r.Prioritas)
	}

	return models.Letter{
		NomorSurat:   strings.TrimSpace(r.NomorSurat),
		Pengirim:     strings.TrimSpace(r.Pengirim),
		JudulSurat:   strings.TrimSpace(r.JudulSurat),
		JenisSurat:   models.LetterMasuk,
		Scope:        strings.TrimSpace(r.Scope),
		Status:       models.StatusBelumDisposisi,
		CreatedByID:  userID,
		FilePath:     filePath,
		Prioritas:    priority,
		IsiSurat:     strings.TrimSpace(r.IsiSurat),
		TanggalSurat: tglSurat,
		TanggalMasuk: tglMasuk,
	}
}

func ApplyUpdateMasuk(letter *models.Letter, req *UpdateLetterMasukRequest) {
	if letter == nil || req == nil {
		return
	}

	if req.NomorSurat != nil {
		letter.NomorSurat = strings.TrimSpace(*req.NomorSurat)
	}
	if req.Pengirim != nil {
		letter.Pengirim = strings.TrimSpace(*req.Pengirim)
	}
	if req.JudulSurat != nil {
		letter.JudulSurat = strings.TrimSpace(*req.JudulSurat)
	}
	if req.IsiSurat != nil {
		letter.IsiSurat = strings.TrimSpace(*req.IsiSurat)
	}
	if req.Scope != nil {
		letter.Scope = strings.TrimSpace(*req.Scope)
	}
	if req.Prioritas != nil {
		letter.Prioritas = models.Priority(*req.Prioritas)
	}

	if req.TanggalSurat != nil && *req.TanggalSurat != "" {
		if t, err := time.Parse("2006-01-02", *req.TanggalSurat); err == nil {
			letter.TanggalSurat = &t
		}
	}
	if req.TanggalMasuk != nil && *req.TanggalMasuk != "" {
		if t, err := time.Parse("2006-01-02", *req.TanggalMasuk); err == nil {
			letter.TanggalMasuk = &t
		}
	}
}
