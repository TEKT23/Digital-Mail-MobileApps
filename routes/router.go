package routes

import (
	"TugasAkhir/handlers"
	"TugasAkhir/middleware"
	"TugasAkhir/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Perubahan: Kita butuh parameter 'db' di sini untuk menginisialisasi handler baru
func SetupRoutes(app *fiber.App, db *gorm.DB) {

	// 1. INISIALISASI HANDLER BARU (Struct Based)
	// Handler ini butuh koneksi DB agar bisa jalan
	lkHandler := handlers.NewLetterKeluarHandler(db)
	lmHandler := handlers.NewLetterMasukHandler(db)

	api := app.Group("/api")

	// ---------------------------------------------------------
	// A. AUTHENTICATION (Mempertahankan Handler Lama)
	// ---------------------------------------------------------
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Post("/register", handlers.Register)
	auth.Post("/refresh", handlers.RefreshToken)
	auth.Post("/logout", handlers.Logout)
	auth.Post("/forgot-password", handlers.RequestPasswordReset)
	auth.Post("/reset-password", handlers.ResetPassword)

	// ---------------------------------------------------------
	// B. MIDDLEWARE GLOBAL (Require Login)
	// ---------------------------------------------------------
	// Semua route di bawah baris ini butuh Token JWT
	api.Use(middleware.RequireAuth())

	// ---------------------------------------------------------
	// C. PROFILE & SETTINGS (Mempertahankan Handler Lama)
	// ---------------------------------------------------------
	settings := api.Group("/settings")
	settings.Get("/profile", handlers.GetMyProfile)
	settings.Put("/profile", handlers.UpdateMyProfile)
	settings.Put("/change-password", handlers.ChangePassword)

	// ---------------------------------------------------------
	// D. WORKFLOW SURAT KELUAR (NEW)
	// Menggantikan logic Create/Update lama yang generic
	// ---------------------------------------------------------
	// Group: /api/letters
	letters := api.Group("/letters")

	// 1. Helper: Dropdown Verifier (Untuk Staf)
	letters.Get("/verifiers", lkHandler.GetAvailableVerifiers)

	// 2. Akses Staf (Create, Edit Draft, Arsip)
	// Note: RequireStaf() mencakup StafProgram & StafLembaga
	letters.Post("/keluar", middleware.RequireStaf(), lkHandler.CreateSuratKeluar)
	letters.Put("/keluar/:id", middleware.RequireStaf(), lkHandler.UpdateSurat)
	letters.Post("/keluar/:id/archive", middleware.RequireStaf(), lkHandler.ArchiveLetter)

	// 3. Akses Manajer (Verifikasi)
	// Note: RequireManajer() mencakup Manajer KPP, Pemas, & PKL
	letters.Post("/keluar/:id/verify/approve", middleware.RequireManajer(), lkHandler.VerifyLetterApprove)
	letters.Post("/keluar/:id/verify/reject", middleware.RequireManajer(), lkHandler.VerifyLetterReject)

	// 4. Akses Direktur (Approval Tanda Tangan)
	letters.Post("/keluar/:id/approve", middleware.RequireDirektur(), lkHandler.ApproveLetterByDirektur)
	letters.Post("/keluar/:id/reject", middleware.RequireDirektur(), lkHandler.RejectLetterByDirektur)

	// ---------------------------------------------------------
	// E. WORKFLOW SURAT MASUK (NEW)
	// ---------------------------------------------------------

	// 1. Akses Staf (Input Surat Masuk)
	letters.Post("/masuk", middleware.RequireStaf(), lmHandler.CreateSuratMasuk)
	letters.Post("/masuk/:id/archive", middleware.RequireStaf(), lmHandler.ArchiveSuratMasuk)

	// 2. Akses Direktur (Disposisi)
	letters.Get("/masuk/for-disposition", middleware.RequireDirektur(), lmHandler.GetLettersMasukForDisposition)
	letters.Post("/masuk/:id/dispose", middleware.RequireDirektur(), lmHandler.DisposeSuratMasuk)

	// ---------------------------------------------------------
	// F. VIEW / LIST DATA (GABUNGAN LAMA & BARU)
	// Handler ListLetters lama mungkin perlu diupdate query-nya nanti
	// agar support filter scope/status baru.
	// ---------------------------------------------------------

	// Get All Letters (Dashboard) - Updated Roles
	letters.Get("/",
		middleware.RequireRole(
			models.RoleStafProgram, models.RoleStafLembaga, // Staf
			models.RoleManajerKPP, models.RoleManajerPemas, models.RoleManajerPKL, // Manajer
			models.RoleDirektur, // Direktur
			models.RoleAdmin,    // Admin
		),
		handlers.ListLetters, // Pastikan handler lama ini masih ada
	)

	// Get Detail Letter
	letters.Get("/:id",
		middleware.RequireRole(
			models.RoleStafProgram, models.RoleStafLembaga,
			models.RoleManajerKPP, models.RoleManajerPemas, models.RoleManajerPKL,
			models.RoleDirektur, models.RoleAdmin,
		),
		handlers.GetLetterByID, // Pastikan handler lama ini masih ada
	)

	// Delete Letter (Hanya Admin atau Staf Pembuat - logic di handler)
	letters.Delete("/:id",
		middleware.RequireRole(models.RoleStafProgram, models.RoleStafLembaga, models.RoleAdmin),
		handlers.DeleteLetter,
	)

	// ---------------------------------------------------------
	// G. ADMIN ZONE (Mempertahankan Handler Lama)
	// ---------------------------------------------------------
	admin := api.Group("/admin",
		middleware.RequireRole(models.RoleAdmin), // Pakai helper baru RequireRole
	)
	admin.Post("/users", handlers.AdminCreateUser)
	admin.Get("/users", handlers.AdminListUsers)
	admin.Get("/users/:id", handlers.AdminGetUserByID)
	admin.Put("/users/:id", handlers.AdminUpdateUser)
	admin.Delete("/users/:id", handlers.AdminDeleteUser)
}
