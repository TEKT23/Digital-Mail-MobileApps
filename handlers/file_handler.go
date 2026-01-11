package handlers

import (
	"TugasAkhir/utils/storage"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// UploadFileHandler - Menangani upload PDF/Gambar
func UploadFileHandler(c *fiber.Ctx) error {
	// 1. Ambil file header dari form-data
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File upload error"})
	}

	// 2. Validasi Ekstensi (PDF/JPG/PNG)
	ext := filepath.Ext(fileHeader.Filename)
	if ext != ".pdf" && ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		return c.Status(400).JSON(fiber.Map{"error": "Hanya file PDF dan Gambar yang diperbolehkan"})
	}

	// 3. Generate Nama Unik (Key)
	filename := fmt.Sprintf("surat_%d%s", time.Now().UnixNano(), ext)

	// 4. Upload ke Storage
	// Kita passing c.Context(), fileHeader langsung, dan filename sebagai key
	uploadedPath, err := storage.UploadFile(c.Context(), fileHeader, filename)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengupload ke storage"})
	}

	// 5. Return Path ke Frontend
	return c.JSON(fiber.Map{
		"success":   true,
		"file_path": uploadedPath,
		"message":   "File uploaded successfully",
	})
}
