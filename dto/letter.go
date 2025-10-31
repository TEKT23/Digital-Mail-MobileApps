package dto

import (
	"time"

	"TugasAkhir/models"
)

type LetterCreateRequest struct {
	Pengirim         string              `json:"pengirim" binding:"required,max=200"`
	NomorSurat       string              `json:"nomor_surat" binding:"required,max=100"`
	NomorAgenda      string              `json:"nomor_agenda" binding:"omitempty,max=100"`
	Disposisi        string              `json:"disposisi" binding:"omitempty"`
	TanggalDisposisi *time.Time          `json:"tanggal_disposisi" binding:"omitempty"`
	BidangTujuan     string              `json:"bidang_tujuan" binding:"omitempty,max=150"`
	JenisSurat       models.LetterType   `json:"jenis_surat" binding:"required,oneof=masuk keluar internal"`
	Prioritas        models.Priority     `json:"prioritas" binding:"omitempty,oneof=biasa segera penting"`
	IsiSurat         string              `json:"isi_surat" binding:"omitempty"`
	TanggalSurat     *time.Time          `json:"tanggal_surat" binding:"omitempty"`
	TanggalMasuk     *time.Time          `json:"tanggal_masuk" binding:"omitempty"`
	JudulSurat       string              `json:"judul_surat" binding:"required,max=255"`
	Kesimpulan       string              `json:"kesimpulan" binding:"omitempty"`
	FilePath         string              `json:"file_path" binding:"omitempty"`
	Status           models.LetterStatus `json:"status" binding:"omitempty,oneof=draft perlu_disposisi belum_disposisi sudah_disposisi"`
	CreatedByID      *uint               `json:"created_by_id" binding:"omitempty"`
	VerifiedByID     *uint               `json:"verified_by_id" binding:"omitempty"`
	DisposedByID     *uint               `json:"disposed_by_id" binding:"omitempty"`
}

type LetterUpdateRequest struct {
	Pengirim         *string              `json:"pengirim" binding:"omitempty,max=200"`
	NomorSurat       *string              `json:"nomor_surat" binding:"omitempty,max=100"`
	NomorAgenda      *string              `json:"nomor_agenda" binding:"omitempty,max=100"`
	Disposisi        *string              `json:"disposisi" binding:"omitempty"`
	TanggalDisposisi *time.Time           `json:"tanggal_disposisi" binding:"omitempty"`
	BidangTujuan     *string              `json:"bidang_tujuan" binding:"omitempty,max=150"`
	JenisSurat       *models.LetterType   `json:"jenis_surat" binding:"omitempty,oneof=masuk keluar internal"`
	Prioritas        *models.Priority     `json:"prioritas" binding:"omitempty,oneof=biasa segera penting"`
	IsiSurat         *string              `json:"isi_surat" binding:"omitempty"`
	TanggalSurat     *time.Time           `json:"tanggal_surat" binding:"omitempty"`
	TanggalMasuk     *time.Time           `json:"tanggal_masuk" binding:"omitempty"`
	JudulSurat       *string              `json:"judul_surat" binding:"omitempty,max=255"`
	Kesimpulan       *string              `json:"kesimpulan" binding:"omitempty"`
	FilePath         *string              `json:"file_path" binding:"omitempty"`
	Status           *models.LetterStatus `json:"status" binding:"omitempty,oneof=draft perlu_disposisi belum_disposisi sudah_disposisi"`
	CreatedByID      *uint                `json:"created_by_id" binding:"omitempty"`
	VerifiedByID     *uint                `json:"verified_by_id" binding:"omitempty"`
	DisposedByID     *uint                `json:"disposed_by_id" binding:"omitempty"`
}

type LetterResponse struct {
	IDSurat          int                 `json:"id_surat"`
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
	VerifiedByID     *uint               `json:"verified_by_id"`
	DisposedByID     *uint               `json:"disposed_by_id"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}
