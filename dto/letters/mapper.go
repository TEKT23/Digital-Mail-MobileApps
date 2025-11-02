package letters

import (
	"strings"

	"TugasAkhir/models"
)

func (r *CreateLetterRequest) ToModel() models.Letter {
	letter := models.Letter{
		Pengirim:     strings.TrimSpace(r.Pengirim),
		NomorSurat:   strings.TrimSpace(r.NomorSurat),
		NomorAgenda:  strings.TrimSpace(r.NomorAgenda),
		Disposisi:    strings.TrimSpace(r.Disposisi),
		BidangTujuan: strings.TrimSpace(r.BidangTujuan),
		JenisSurat:   r.JenisSurat,
		Prioritas:    r.Prioritas,
		IsiSurat:     r.IsiSurat,
		TanggalSurat: r.TanggalSurat,
		TanggalMasuk: r.TanggalMasuk,
		JudulSurat:   strings.TrimSpace(r.JudulSurat),
		Kesimpulan:   strings.TrimSpace(r.Kesimpulan),
		FilePath:     strings.TrimSpace(r.FilePath),
		Status:       r.Status,
		CreatedByID:  r.CreatedByID,
		VerifiedByID: r.VerifiedByID,
		DisposedByID: r.DisposedByID,
	}

	if r.Prioritas == "" {
		letter.Prioritas = models.PriorityBiasa
	}
	if r.Status == "" {
		letter.Status = models.StatusDraft
	}
	letter.TanggalDisposisi = r.TanggalDisposisi

	return letter
}

func ApplyUpdate(letter *models.Letter, req *UpdateLetterRequest) {
	if letter == nil || req == nil {
		return
	}

	if req.Pengirim != nil {
		letter.Pengirim = strings.TrimSpace(*req.Pengirim)
	}
	if req.NomorSurat != nil {
		letter.NomorSurat = strings.TrimSpace(*req.NomorSurat)
	}
	if req.NomorAgenda != nil {
		letter.NomorAgenda = strings.TrimSpace(*req.NomorAgenda)
	}
	if req.Disposisi != nil {
		letter.Disposisi = strings.TrimSpace(*req.Disposisi)
	}
	if req.TanggalDisposisi != nil {
		letter.TanggalDisposisi = req.TanggalDisposisi
	}
	if req.BidangTujuan != nil {
		letter.BidangTujuan = strings.TrimSpace(*req.BidangTujuan)
	}
	if req.JenisSurat != nil {
		letter.JenisSurat = *req.JenisSurat
	}
	if req.Prioritas != nil {
		letter.Prioritas = *req.Prioritas
	}
	if req.IsiSurat != nil {
		letter.IsiSurat = *req.IsiSurat
	}
	if req.TanggalSurat != nil {
		letter.TanggalSurat = req.TanggalSurat
	}
	if req.TanggalMasuk != nil {
		letter.TanggalMasuk = req.TanggalMasuk
	}
	if req.JudulSurat != nil {
		letter.JudulSurat = strings.TrimSpace(*req.JudulSurat)
	}
	if req.Kesimpulan != nil {
		letter.Kesimpulan = strings.TrimSpace(*req.Kesimpulan)
	}
	if req.FilePath != nil {
		letter.FilePath = strings.TrimSpace(*req.FilePath)
	}
	if req.Status != nil {
		letter.Status = *req.Status
	}
	if req.CreatedByID != nil {
		letter.CreatedByID = req.CreatedByID
	}
	if req.VerifiedByID != nil {
		letter.VerifiedByID = req.VerifiedByID
	}
	if req.DisposedByID != nil {
		letter.DisposedByID = req.DisposedByID
	}
}
