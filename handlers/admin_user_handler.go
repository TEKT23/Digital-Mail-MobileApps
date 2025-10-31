package handlers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"TugasAkhir/config"
	"TugasAkhir/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type adminUserReq struct {
	Username  *string      `json:"username"`
	FirstName *string      `json:"first_name"`
	LastName  *string      `json:"last_name"`
	Email     *string      `json:"email"`
	Password  *string      `json:"password"` // plain; di-hash saat create/update bila diisi
	Role      *models.Role `json:"role"`     // bagian_umum|adc|direktur|admin
	Jabatan   *string      `json:"jabatan"`
	Atribut   *string      `json:"atribut"`
}

type adminUserResp struct {
	ID        uint        `json:"id"`
	Username  string      `json:"username"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
	Role      models.Role `json:"role"`
	Jabatan   string      `json:"jabatan"`
	Atribut   string      `json:"atribut"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

func toAdminResp(u models.User) adminUserResp {
	return adminUserResp{
		ID:        u.ID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Role:      u.Role,
		Jabatan:   u.Jabatan,
		Atribut:   u.Atribut,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}

func requiredStr(field string, v *string) error {
	if v == nil || strings.TrimSpace(*v) == "" {
		return errors.New(field + " is required")
	}
	return nil
}

func validRole(r *models.Role) bool {
	if r == nil {
		return false
	}
	switch *r {
	case models.RoleBagianUmum, models.RoleADC, models.RoleDirektur, models.RoleAdmin:
		return true
	default:
		return false
	}
}

func bcryptHash(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), 12)
	return string(b), err
}

func isDupErr(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}

// Crate API
func AdminCreateUser(c *fiber.Ctx) error {
	var in adminUserReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}

	if err := requiredStr("username", in.Username); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := requiredStr("email", in.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if !validRole(in.Role) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role is required: bagian_umum|adc|direktur|admin"})
	}

	var pwdHash string
	if in.Password != nil && strings.TrimSpace(*in.Password) != "" {
		h, err := bcryptHash(*in.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
		}
		pwdHash = h
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password is required"})
	}

	u := models.User{
		Username:     strings.TrimSpace(*in.Username),
		Email:        strings.TrimSpace(*in.Email),
		PasswordHash: pwdHash,
		Role:         *in.Role,
	}

	if in.FirstName != nil {
		u.FirstName = strings.TrimSpace(*in.FirstName)
	}
	if in.LastName != nil {
		u.LastName = strings.TrimSpace(*in.LastName)
	}
	if in.Jabatan != nil {
		u.Jabatan = strings.TrimSpace(*in.Jabatan)
	}
	if in.Atribut != nil {
		u.Atribut = *in.Atribut
	}

	if err := config.DB.Create(&u).Error; err != nil {
		if isDupErr(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username or email already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(toAdminResp(u))
}

// READ ONE
func AdminGetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var u models.User
	if err := config.DB.First(&u, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(toAdminResp(u))
}

// LIST + FILTER
func AdminListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	role := strings.TrimSpace(c.Query("role", ""))
	q := strings.TrimSpace(c.Query("q", "")) // search username/email/first/last

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset := (page - 1) * limit

	tx := config.DB.Model(&models.User{})
	if role != "" {
		tx = tx.Where("role = ?", role)
	}
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where(
			config.DB.Where("username LIKE ?", like).
				Or("email LIKE ?", like).
				Or("first_name LIKE ?", like).
				Or("last_name LIKE ?", like),
		)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var users []models.User
	if err := tx.Order("id DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	resp := make([]adminUserResp, 0, len(users))
	for _, it := range users {
		resp = append(resp, toAdminResp(it))
	}

	return c.JSON(fiber.Map{
		"data":  resp,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

// Update API(partial)
func AdminUpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var u models.User
	if err := config.DB.First(&u, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var in adminUserReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}

	if in.Username != nil {
		u.Username = strings.TrimSpace(*in.Username)
	}
	if in.FirstName != nil {
		u.FirstName = strings.TrimSpace(*in.FirstName)
	}
	if in.LastName != nil {
		u.LastName = strings.TrimSpace(*in.LastName)
	}
	if in.Email != nil {
		u.Email = strings.TrimSpace(*in.Email)
	}
	if in.Role != nil {
		if !validRole(in.Role) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role"})
		}
		u.Role = *in.Role
	}
	if in.Jabatan != nil {
		u.Jabatan = strings.TrimSpace(*in.Jabatan)
	}
	if in.Atribut != nil {
		u.Atribut = *in.Atribut
	}
	if in.Password != nil && strings.TrimSpace(*in.Password) != "" {
		h, err := bcryptHash(*in.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
		}
		u.PasswordHash = h
	}

	if err := config.DB.Save(&u).Error; err != nil {
		if isDupErr(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username or email already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(toAdminResp(u))
}

// Delete User API
func AdminDeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
