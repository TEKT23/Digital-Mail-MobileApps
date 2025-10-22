package handlers

import (
	"strconv"
	"time"

	"TugasAkhir/config"
	"TugasAkhir/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LetterReq struct {
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

func notFound(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
}
func badReq(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func applyPatch(m *models.Letter, in *LetterReq) {
	if in.Pengirim != nil {
		m.Pengirim = *in.Pengirim
	}
	if in.NomorSurat != nil {
		m.NomorSurat = *in.NomorSurat
	}
	if in.NomorAgenda != nil {
		m.NomorAgenda = *in.NomorAgenda
	}
	if in.Disposisi != nil {
		m.Disposisi = *in.Disposisi
	}
	if in.TanggalDisposisi != nil {
		m.TanggalDisposisi = in.TanggalDisposisi
	}
	if in.BidangTujuan != nil {
		m.BidangTujuan = *in.BidangTujuan
	}
	if in.JenisSurat != nil {
		m.JenisSurat = *in.JenisSurat
	}
	if in.Prioritas != nil {
		m.Prioritas = *in.Prioritas
	}
	if in.IsiSurat != nil {
		m.IsiSurat = *in.IsiSurat
	}
	if in.TanggalSurat != nil {
		m.TanggalSurat = in.TanggalSurat
	}
	if in.TanggalMasuk != nil {
		m.TanggalMasuk = in.TanggalMasuk
	}
	if in.JudulSurat != nil {
		m.JudulSurat = *in.JudulSurat
	}
	if in.Kesimpulan != nil {
		m.Kesimpulan = *in.Kesimpulan
	}
	if in.FilePath != nil {
		m.FilePath = *in.FilePath
	}
	if in.Status != nil {
		m.Status = *in.Status
	}
	if in.CreatedByID != nil {
		m.CreatedByID = in.CreatedByID
	}
	if in.VerifiedByID != nil {
		m.VerifiedByID = in.VerifiedByID
	}
	if in.DisposedByID != nil {
		m.DisposedByID = in.DisposedByID
	}
}

// ---------- Handlers ----------

// POST /api/letters
func CreateLetter(c *fiber.Ctx) error {
	var in LetterReq
	if err := c.BodyParser(&in); err != nil {
		return badReq(c, "invalid JSON body")
	}

	// Validasi minimal
	if in.JenisSurat == nil {
		return badReq(c, "jenis_surat is required: masuk|keluar|internal")
	}
	if in.Status == nil {
		// default di model sudah "draft", tapi jika user kirim null, biarkan default
	}

	m := &models.Letter{}
	applyPatch(m, &in)

	if err := config.DB.Create(m).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(m)
}

// GET /api/letters/:id
func GetLetterByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var m models.Letter
	if err := config.DB.First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return notFound(c)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(m)
}

// GET /api/letters?page=&limit=
func ListLetters(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	if err := config.DB.Model(&models.Letter{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var items []models.Letter
	if err := config.DB.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":  items,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

// PUT /api/letters/:id
func UpdateLetter(c *fiber.Ctx) error {
	id := c.Params("id")
	var m models.Letter
	if err := config.DB.First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return notFound(c)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var in LetterReq
	if err := c.BodyParser(&in); err != nil {
		return badReq(c, "invalid JSON body")
	}

	applyPatch(&m, &in)

	if err := config.DB.Save(&m).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(m)
}

// DELETE /api/letters/:id
func DeleteLetter(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.Letter{}, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
