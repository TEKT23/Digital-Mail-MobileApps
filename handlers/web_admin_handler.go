package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"TugasAkhir/config"
	"TugasAkhir/middleware"
	"TugasAkhir/models"
	"TugasAkhir/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// WebAdminHandler - Handler untuk admin web panel
type WebAdminHandler struct {
	templates map[string]*template.Template
}

// PageData - Data untuk template
type PageData struct {
	Title      string
	Active     string
	User       *models.User
	Error      string
	Success    string
	Email      string
	Form       UserFormData
	Errors     map[string]string
	EditUser   *models.User
	Users      []models.User
	Query      string
	RoleFilter string
	Page       int
	TotalPages int
	Pages      []int
	Stats      DashboardStats
}

type UserFormData struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
	Role      string
	Jabatan   string
	Atribut   string
}

type DashboardStats struct {
	TotalUsers   int64
	TotalStaf    int64
	TotalManajer int64
	TotalAdmin   int64
}

// NewWebAdminHandler - Constructor
func NewWebAdminHandler() *WebAdminHandler {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
	}

	templates := make(map[string]*template.Template)
	layoutFile := "templates/layouts/base.html"

	// Parse each page template with the base layout
	pages := map[string]string{
		"login":        "templates/admin/login.html",
		"dashboard":    "templates/admin/dashboard.html",
		"users_list":   "templates/admin/users/list.html",
		"users_create": "templates/admin/users/create.html",
		"users_edit":   "templates/admin/users/edit.html",
		"settings":     "templates/admin/settings.html",
	}

	for name, pageFile := range pages {
		t := template.Must(template.New(filepath.Base(layoutFile)).Funcs(funcMap).ParseFiles(layoutFile, pageFile))
		templates[name] = t
	}

	return &WebAdminHandler{
		templates: templates,
	}
}

// Helper untuk render template
func (h *WebAdminHandler) render(c *fiber.Ctx, templateName string, data PageData) error {
	t, ok := h.templates[templateName]
	if !ok {
		log.Printf("Template not found: %s", templateName)
		return c.Status(500).SendString("Template not found: " + templateName)
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		log.Printf("Template error: %v", err)
		return c.Status(500).SendString("Template error: " + err.Error())
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(buf.Bytes())
}

// =====================
// AUTH HANDLERS
// =====================

// ShowLoginPage - GET /admin/login
func (h *WebAdminHandler) ShowLoginPage(c *fiber.Ctx) error {
	// Cek jika sudah login
	sess, _ := middleware.AdminSessionStore.Get(c)
	if sess.Get(middleware.SessionAdminIDKey) != nil {
		return c.Redirect("/admin")
	}

	return h.render(c, "login", PageData{
		Title:  "Login",
		Error:  c.Query("error"),
		Active: "login",
	})
}

// HandleLogin - POST /admin/login
func (h *WebAdminHandler) HandleLogin(c *fiber.Ctx) error {
	email := strings.TrimSpace(c.FormValue("email"))
	password := c.FormValue("password")

	// Validasi
	if email == "" || password == "" {
		return h.render(c, "login", PageData{
			Title:  "Login",
			Error:  "Email dan password harus diisi",
			Email:  email,
			Active: "login",
		})
	}

	// Cari user
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return h.render(c, "login", PageData{
			Title:  "Login",
			Error:  "Email atau password salah",
			Email:  email,
			Active: "login",
		})
	}

	// Cek password
	if !utils.CheckPassword(user.PasswordHash, password) {
		return h.render(c, "login", PageData{
			Title:  "Login",
			Error:  "Email atau password salah",
			Email:  email,
			Active: "login",
		})
	}

	// Cek role admin
	if user.Role != models.RoleAdmin {
		return h.render(c, "login", PageData{
			Title:  "Login",
			Error:  "Akses ditolak. Hanya admin yang dapat login.",
			Email:  email,
			Active: "login",
		})
	}

	// Buat session
	sess, err := middleware.AdminSessionStore.Get(c)
	if err != nil {
		return h.render(c, "login", PageData{
			Title:  "Login",
			Error:  "Gagal membuat session",
			Email:  email,
			Active: "login",
		})
	}

	sess.Set(middleware.SessionAdminIDKey, user.ID)
	sess.Set(middleware.SessionAdminRoleKey, string(user.Role))
	if err := sess.Save(); err != nil {
		return h.render(c, "login", PageData{
			Title:  "Login",
			Error:  "Gagal menyimpan session",
			Email:  email,
			Active: "login",
		})
	}

	return c.Redirect("/admin")
}

// HandleLogout - POST /admin/logout
func (h *WebAdminHandler) HandleLogout(c *fiber.Ctx) error {
	sess, err := middleware.AdminSessionStore.Get(c)
	if err == nil {
		sess.Destroy()
	}
	return c.Redirect("/admin/login")
}

// =====================
// DASHBOARD HANDLERS
// =====================

// ShowDashboard - GET /admin
func (h *WebAdminHandler) ShowDashboard(c *fiber.Ctx) error {
	user, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	// Get stats
	var stats DashboardStats
	config.DB.Model(&models.User{}).Count(&stats.TotalUsers)
	config.DB.Model(&models.User{}).Where("role IN ?", []string{"staf_program", "staf_lembaga"}).Count(&stats.TotalStaf)
	config.DB.Model(&models.User{}).Where("role IN ?", []string{"manajer_kpp", "manajer_pemas", "manajer_pkl"}).Count(&stats.TotalManajer)
	config.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&stats.TotalAdmin)

	return h.render(c, "dashboard", PageData{
		Title:  "Dashboard",
		Active: "dashboard",
		User:   user,
		Stats:  stats,
	})
}

// =====================
// USER CRUD HANDLERS
// =====================

// ShowUserList - GET /admin/users
func (h *WebAdminHandler) ShowUserList(c *fiber.Ctx) error {
	user, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	// Get flash messages from query
	success := c.Query("success")
	errorMsg := c.Query("error")

	// Pagination dan filter
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 15
	offset := (page - 1) * limit

	roleFilter := strings.TrimSpace(c.Query("role"))
	query := strings.TrimSpace(c.Query("q"))

	// Build query
	tx := config.DB.Model(&models.User{})
	if roleFilter != "" {
		tx = tx.Where("role = ?", roleFilter)
	}
	if query != "" {
		like := "%" + query + "%"
		tx = tx.Where(
			config.DB.Where("username LIKE ?", like).
				Or("email LIKE ?", like).
				Or("first_name LIKE ?", like).
				Or("last_name LIKE ?", like),
		)
	}

	// Count total
	var total int64
	tx.Count(&total)

	// Get users
	var users []models.User
	tx.Order("id DESC").Limit(limit).Offset(offset).Find(&users)

	// Calculate pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	var pages []int
	for i := 1; i <= totalPages; i++ {
		if i <= 5 || i > totalPages-2 || (i >= page-1 && i <= page+1) {
			pages = append(pages, i)
		}
	}

	return h.render(c, "users_list", PageData{
		Title:      "Manajemen User",
		Active:     "users",
		User:       user,
		Users:      users,
		Query:      query,
		RoleFilter: roleFilter,
		Page:       page,
		TotalPages: totalPages,
		Pages:      pages,
		Success:    success,
		Error:      errorMsg,
	})
}

// ShowCreateUserForm - GET /admin/users/create
func (h *WebAdminHandler) ShowCreateUserForm(c *fiber.Ctx) error {
	user, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	return h.render(c, "users_create", PageData{
		Title:  "Tambah User",
		Active: "users",
		User:   user,
		Errors: make(map[string]string),
	})
}

// HandleCreateUser - POST /admin/users
func (h *WebAdminHandler) HandleCreateUser(c *fiber.Ctx) error {
	user, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	form := UserFormData{
		Username:  strings.TrimSpace(c.FormValue("username")),
		Email:     strings.TrimSpace(c.FormValue("email")),
		FirstName: strings.TrimSpace(c.FormValue("first_name")),
		LastName:  strings.TrimSpace(c.FormValue("last_name")),
		Role:      c.FormValue("role"),
		Jabatan:   strings.TrimSpace(c.FormValue("jabatan")),
		Atribut:   strings.TrimSpace(c.FormValue("atribut")),
	}
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")

	errors := make(map[string]string)

	// Validasi
	if form.Username == "" {
		errors["username"] = "Username harus diisi"
	}
	if form.Email == "" {
		errors["email"] = "Email harus diisi"
	}
	if form.FirstName == "" {
		errors["first_name"] = "Nama depan harus diisi"
	}
	if form.Role == "" {
		errors["role"] = "Role harus dipilih"
	}
	if password == "" {
		errors["password"] = "Password harus diisi"
	} else if len(password) < 6 {
		errors["password"] = "Password minimal 6 karakter"
	} else if password != confirmPassword {
		errors["password"] = "Password tidak sama"
	}

	if len(errors) > 0 {
		return h.render(c, "users_create", PageData{
			Title:  "Tambah User",
			Active: "users",
			User:   user,
			Form:   form,
			Errors: errors,
		})
	}

	// Hash password
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		errors["password"] = "Gagal memproses password"
		return h.render(c, "users_create", PageData{
			Title:  "Tambah User",
			Active: "users",
			User:   user,
			Form:   form,
			Errors: errors,
		})
	}

	// Create user
	newUser := models.User{
		Username:     form.Username,
		Email:        form.Email,
		FirstName:    form.FirstName,
		LastName:     form.LastName,
		Role:         models.Role(form.Role),
		Jabatan:      form.Jabatan,
		Atribut:      form.Atribut,
		PasswordHash: passwordHash,
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		if utils.IsDuplicateError(err) {
			errors["username"] = "Username atau email sudah digunakan"
			return h.render(c, "users_create", PageData{
				Title:  "Tambah User",
				Active: "users",
				User:   user,
				Form:   form,
				Errors: errors,
			})
		}
		return h.render(c, "users_create", PageData{
			Title:  "Tambah User",
			Active: "users",
			User:   user,
			Form:   form,
			Error:  "Gagal membuat user: " + err.Error(),
			Errors: make(map[string]string),
		})
	}

	return c.Redirect("/admin/users?success=User berhasil dibuat")
}

// ShowEditUserForm - GET /admin/users/:id/edit
func (h *WebAdminHandler) ShowEditUserForm(c *fiber.Ctx) error {
	user, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")
	var editUser models.User
	if err := config.DB.First(&editUser, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Redirect("/admin/users?error=User tidak ditemukan")
		}
		return c.Redirect("/admin/users?error=Gagal mengambil data user")
	}

	return h.render(c, "users_edit", PageData{
		Title:    "Edit User",
		Active:   "users",
		User:     user,
		EditUser: &editUser,
		Errors:   make(map[string]string),
	})
}

// HandleUpdateUser - POST /admin/users/:id
func (h *WebAdminHandler) HandleUpdateUser(c *fiber.Ctx) error {
	user, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")
	var editUser models.User
	if err := config.DB.First(&editUser, "id = ?", id).Error; err != nil {
		return c.Redirect("/admin/users?error=User tidak ditemukan")
	}

	// Update fields
	editUser.Username = strings.TrimSpace(c.FormValue("username"))
	editUser.Email = strings.TrimSpace(c.FormValue("email"))
	editUser.FirstName = strings.TrimSpace(c.FormValue("first_name"))
	editUser.LastName = strings.TrimSpace(c.FormValue("last_name"))
	editUser.Role = models.Role(c.FormValue("role"))
	editUser.Jabatan = strings.TrimSpace(c.FormValue("jabatan"))
	editUser.Atribut = strings.TrimSpace(c.FormValue("atribut"))

	errors := make(map[string]string)

	// Validasi
	if editUser.Username == "" {
		errors["username"] = "Username harus diisi"
	}
	if editUser.Email == "" {
		errors["email"] = "Email harus diisi"
	}
	if editUser.FirstName == "" {
		errors["first_name"] = "Nama depan harus diisi"
	}

	// Update password jika diisi
	newPassword := c.FormValue("password")
	if newPassword != "" {
		if len(newPassword) < 6 {
			errors["password"] = "Password minimal 6 karakter"
		} else {
			hash, err := utils.HashPassword(newPassword)
			if err != nil {
				errors["password"] = "Gagal memproses password"
			} else {
				editUser.PasswordHash = hash
			}
		}
	}

	if len(errors) > 0 {
		return h.render(c, "users_edit", PageData{
			Title:    "Edit User",
			Active:   "users",
			User:     user,
			EditUser: &editUser,
			Errors:   errors,
		})
	}

	if err := config.DB.Save(&editUser).Error; err != nil {
		if utils.IsDuplicateError(err) {
			errors["username"] = "Username atau email sudah digunakan"
			return h.render(c, "users_edit", PageData{
				Title:    "Edit User",
				Active:   "users",
				User:     user,
				EditUser: &editUser,
				Errors:   errors,
			})
		}
		return h.render(c, "users_edit", PageData{
			Title:    "Edit User",
			Active:   "users",
			User:     user,
			EditUser: &editUser,
			Error:    "Gagal update user: " + err.Error(),
			Errors:   make(map[string]string),
		})
	}

	return c.Redirect("/admin/users?success=User berhasil diupdate")
}

// HandleDeleteUser - POST /admin/users/:id/delete
func (h *WebAdminHandler) HandleDeleteUser(c *fiber.Ctx) error {
	_, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")

	// Cegah hapus diri sendiri
	adminID := c.Locals(middleware.SessionAdminIDKey)
	if fmt.Sprintf("%v", adminID) == id {
		return c.Redirect("/admin/users?error=Tidak dapat menghapus akun sendiri")
	}

	result := config.DB.Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return c.Redirect("/admin/users?error=Gagal menghapus user")
	}
	if result.RowsAffected == 0 {
		return c.Redirect("/admin/users?error=User tidak ditemukan")
	}

	return c.Redirect("/admin/users?success=User berhasil dihapus")
}

// =====================
// SETTINGS HANDLERS
// =====================

// ShowSettings - GET /admin/settings
func (h *WebAdminHandler) ShowSettings(c *fiber.Ctx) error {
	user, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	success := c.Query("success")
	errorMsg := c.Query("error")

	return h.render(c, "settings", PageData{
		Title:   "Settings",
		Active:  "settings",
		User:    user,
		Success: success,
		Error:   errorMsg,
	})
}

// HandleUpdateProfile - POST /admin/settings/profile
func (h *WebAdminHandler) HandleUpdateProfile(c *fiber.Ctx) error {
	user, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	user.FirstName = strings.TrimSpace(c.FormValue("first_name"))
	user.LastName = strings.TrimSpace(c.FormValue("last_name"))
	user.Jabatan = strings.TrimSpace(c.FormValue("jabatan"))

	if err := config.DB.Save(user).Error; err != nil {
		return c.Redirect("/admin/settings?error=Gagal update profil")
	}

	return c.Redirect("/admin/settings?success=Profil berhasil diupdate")
}

// HandleChangePassword - POST /admin/settings/password
func (h *WebAdminHandler) HandleChangePassword(c *fiber.Ctx) error {
	user, err := middleware.GetAdminFromSession(c)
	if err != nil {
		return c.Redirect("/admin/login")
	}

	oldPassword := c.FormValue("old_password")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")

	// Validasi
	if !utils.CheckPassword(user.PasswordHash, oldPassword) {
		return c.Redirect("/admin/settings?error=Password lama salah")
	}

	if len(newPassword) < 6 {
		return c.Redirect("/admin/settings?error=Password baru minimal 6 karakter")
	}

	if newPassword != confirmPassword {
		return c.Redirect("/admin/settings?error=Konfirmasi password tidak sama")
	}

	// Hash password baru
	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return c.Redirect("/admin/settings?error=Gagal memproses password")
	}

	user.PasswordHash = hash
	if err := config.DB.Save(user).Error; err != nil {
		return c.Redirect("/admin/settings?error=Gagal update password")
	}

	return c.Redirect("/admin/settings?success=Password berhasil diubah")
}
