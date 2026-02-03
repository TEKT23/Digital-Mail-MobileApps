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
	auth.Get("/reset-password", handlers.ShowResetPasswordForm)
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

	// --- A. HELPER ROUTES (must be before :id routes) ---
	letters.Get("/verifiers", lkHandler.GetAvailableVerifiers)

	// --- B. WORKFLOW SURAT KELUAR ---

	// Dashboard & Aksi STAF
	letters.Get("/keluar/my", middleware.RequireStaf(), lkHandler.GetMyLetters)
	letters.Post("/keluar", middleware.RequireStaf(), lkHandler.CreateSuratKeluar)
	letters.Put("/keluar/:id", middleware.RequireStaf(), lkHandler.UpdateDraftLetter)
	letters.Post("/keluar/:id/archive", middleware.RequireStaf(), lkHandler.ArchiveLetter)

	// Dashboard & Aksi MANAJER
	letters.Get("/keluar/need-verification", middleware.RequireManajer(), lkHandler.GetLettersNeedVerification)
	letters.Post("/keluar/:id/verify/approve", middleware.RequireManajer(), lkHandler.VerifyLetterApprove)
	letters.Post("/keluar/:id/verify/reject", middleware.RequireManajer(), lkHandler.VerifyLetterReject)

	// Dashboard & Aksi DIREKTUR
	letters.Get("/keluar/need-approval", middleware.RequireDirektur(), lkHandler.GetLettersNeedApproval)
	letters.Get("/keluar/my-approvals", middleware.RequireDirektur(), lkHandler.GetMyApprovals)
	letters.Post("/keluar/:id/approve", middleware.RequireDirektur(), lkHandler.ApproveLetterByDirektur)
	letters.Post("/keluar/:id/reject", middleware.RequireDirektur(), lkHandler.RejectLetterByDirektur)

	// --- C. WORKFLOW SURAT MASUK ---

	// Aksi STAF
	letters.Get("/masuk/my", middleware.RequireStaf(), lmHandler.GetMySuratMasuk)
	letters.Post("/masuk", middleware.RequireStaf(), lmHandler.CreateSuratMasuk)
	letters.Put("/masuk/:id", middleware.RequireStaf(), lmHandler.UpdateSuratMasuk)
	letters.Post("/masuk/:id/archive", middleware.RequireStaf(), lmHandler.ArchiveSuratMasuk)

	// Dashboard & Aksi DIREKTUR (Disposisi)
	letters.Get("/masuk/need-disposition", middleware.RequireDirektur(), lmHandler.GetLettersMasukForDisposition)
	letters.Get("/masuk/my-dispositions", middleware.RequireDirektur(), lmHandler.GetMyDispositions)
	letters.Post("/masuk/:id/dispose", middleware.RequireDirektur(), lmHandler.DisposeSuratMasuk)

	// Reply Linking - Surat masuk yang butuh balasan
	letters.Get("/masuk/needs-reply", middleware.RequireStaf(), lmHandler.GetLettersNeedingReply)

	// --- D. GENERIC ROUTES (must be LAST to avoid catching specific routes) ---
	// Melihat Detail Surat (any letter by ID)
	letters.Get("/:id", commonHandler.GetLetterByID)
	// Menghapus/Membatalkan Surat (Soft Delete / Cancel)
	letters.Delete("/:id", commonHandler.DeleteLetter)

	// 6. ADMIN ZONE (API)
	admin := api.Group("/admin", middleware.RequireAdmin())
	admin.Post("/users", handlers.AdminCreateUser)
	admin.Get("/users", handlers.AdminListUsers)
	admin.Get("/users/:id", handlers.AdminGetUserByID)
	admin.Put("/users/:id", handlers.AdminUpdateUser)
	admin.Delete("/users/:id", handlers.AdminDeleteUser)

	// 7. ADMIN WEB PANEL (Session-based auth)
	webHandler := handlers.NewWebAdminHandler()

	// Static files
	app.Static("/static", "./static")

	// Public routes (login)
	adminWeb := app.Group("/admin")
	adminWeb.Get("/login", webHandler.ShowLoginPage)
	adminWeb.Post("/login", webHandler.HandleLogin)

	// Protected routes (require session)
	adminWebAuth := adminWeb.Group("", middleware.RequireAdminSession())
	adminWebAuth.Post("/logout", webHandler.HandleLogout)
	adminWebAuth.Get("/", webHandler.ShowDashboard)
	adminWebAuth.Get("/users", webHandler.ShowUserList)
	adminWebAuth.Get("/users/create", webHandler.ShowCreateUserForm)
	adminWebAuth.Post("/users", webHandler.HandleCreateUser)
	adminWebAuth.Get("/users/:id/edit", webHandler.ShowEditUserForm)
	adminWebAuth.Post("/users/:id", webHandler.HandleUpdateUser)
	adminWebAuth.Post("/users/:id/delete", webHandler.HandleDeleteUser)
	adminWebAuth.Get("/settings", webHandler.ShowSettings)
	adminWebAuth.Post("/settings/profile", webHandler.HandleUpdateProfile)
	adminWebAuth.Post("/settings/password", webHandler.HandleChangePassword)
}
