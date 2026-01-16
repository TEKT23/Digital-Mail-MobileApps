package routes

import (
	"TugasAkhir/handlers"
	"TugasAkhir/middleware"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {

	// 1. INISIALISASI HANDLER
	lkHandler := handlers.NewLetterKeluarHandler(db)
	lmHandler := handlers.NewLetterMasukHandler(db)
	commonHandler := handlers.NewLetterCommonHandler(db) //

	api := app.Group("/api")

	// 2. AUTH & PUBLIC ROUTES
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Post("/register", handlers.Register)
	auth.Post("/refresh", handlers.RefreshToken)
	auth.Post("/logout", handlers.Logout)
	auth.Post("/forgot-password", handlers.RequestPasswordReset)
	auth.Post("/reset-password", handlers.ResetPassword)

	// 3. MIDDLEWARE & UTILITY
	api.Use(middleware.RequireAuth())

	// Route Upload File (PDF/Gambar)
	api.Post("/upload", handlers.UploadFileHandler)

	// 4. PROFILE & SETTINGS
	settings := api.Group("/settings")
	settings.Get("/profile", handlers.GetMyProfile)
	settings.Put("/profile", handlers.UpdateMyProfile)
	settings.Put("/change-password", handlers.ChangePassword)

	// 5. MANAJEMEN SURAT (Group: /api/letters)
	letters := api.Group("/letters")

	// --- A. COMMON ROUTES
	// Melihat Detail Surat

	// 1. Helper untuk Form (List Manajer)
	letters.Get("/verifiers", lkHandler.GetAvailableVerifiers)

	letters.Get("/:id", commonHandler.GetLetterByID)

	// Menghapus/Membatalkan Surat (Soft Delete / Cancel)
	letters.Delete("/:id", commonHandler.DeleteLetter)

	// --- B. WORKFLOW SURAT KELUAR ---

	// 2. Dashboard & Aksi STAF
	letters.Get("/keluar/my", middleware.RequireStaf(), lkHandler.GetMyLetters)
	// Buat Surat Baru
	letters.Post("/keluar", middleware.RequireStaf(), lkHandler.CreateSuratKeluar)
	// Edit Draft / Revisi (Handler Baru: UpdateDraftLetter)
	letters.Put("/keluar/:id", middleware.RequireStaf(), lkHandler.UpdateDraftLetter)
	// Arsipkan Surat (Finalisasi)
	letters.Post("/keluar/:id/archive", middleware.RequireStaf(), lkHandler.ArchiveLetter)

	// 3. Dashboard & Aksi MANAJER
	letters.Get("/keluar/need-verification", middleware.RequireManajer(), lkHandler.GetLettersNeedVerification)
	// Eksekusi Verifikasi
	letters.Post("/keluar/:id/verify/approve", middleware.RequireManajer(), lkHandler.VerifyLetterApprove)
	letters.Post("/keluar/:id/verify/reject", middleware.RequireManajer(), lkHandler.VerifyLetterReject)

	// 4. Dashboard & Aksi DIREKTUR
	letters.Get("/keluar/need-approval", middleware.RequireDirektur(), lkHandler.GetLettersNeedApproval)
	// Eksekusi Approval
	letters.Post("/keluar/:id/approve", middleware.RequireDirektur(), lkHandler.ApproveLetterByDirektur)
	letters.Post("/keluar/:id/reject", middleware.RequireDirektur(), lkHandler.RejectLetterByDirektur)

	// --- C. WORKFLOW SURAT MASUK ---

	// 1. Aksi STAF
	letters.Post("/masuk", middleware.RequireStaf(), lmHandler.CreateSuratMasuk)
	letters.Post("/masuk/:id/archive", middleware.RequireStaf(), lmHandler.ArchiveSuratMasuk)

	// 2. Dashboard & Aksi DIREKTUR (Disposisi)
	// Get surat masuk yang BELUM DISPOSISI
	letters.Get("/masuk/need-disposition", middleware.RequireDirektur(), lmHandler.GetLettersMasukForDisposition)
	// Eksekusi Disposisi
	letters.Post("/masuk/:id/dispose", middleware.RequireDirektur(), lmHandler.DisposeSuratMasuk)

	// 6. ADMIN ZONE
	admin := api.Group("/admin", middleware.RequireAdmin())
	admin.Post("/users", handlers.AdminCreateUser)
	admin.Get("/users", handlers.AdminListUsers)
	admin.Get("/users/:id", handlers.AdminGetUserByID)
	admin.Put("/users/:id", handlers.AdminUpdateUser)
	admin.Delete("/users/:id", handlers.AdminDeleteUser)
}
