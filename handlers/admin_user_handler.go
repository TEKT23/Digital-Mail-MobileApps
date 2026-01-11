package handlers

import (
	"strconv"
	"strings"

	"TugasAkhir/config"
	userdto "TugasAkhir/dto/users"
	"TugasAkhir/models"
	"TugasAkhir/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func bcryptHash(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), 12)
	return string(b), err
}

func isDupErr(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}

// Create API
func AdminCreateUser(c *fiber.Ctx) error {
	var req userdto.AdminUserCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "invalid request body", err.Error())
	}

	if validationErrors := req.Validate(); len(validationErrors) > 0 {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "validation error", validationErrors)
	}

	passwordHash, err := utils.HashPassword(strings.TrimSpace(req.Password))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to hash password", nil)
	}

	user := models.User{
		Username:     strings.TrimSpace(req.Username),
		FirstName:    strings.TrimSpace(req.FirstName),
		LastName:     strings.TrimSpace(req.LastName),
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: passwordHash,
		Role:         req.Role,
		Jabatan:      strings.TrimSpace(req.Jabatan),
		Atribut:      req.Atribut,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		if utils.IsDuplicateError(err) {
			return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "username or email already exists", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to create user", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "user created successfully", userdto.NewAdminUserResponse(user))
}

// READ ONE
func AdminGetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User
	if err := config.DB.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "user not found", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve user", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "user retrieved successfully", userdto.NewAdminUserResponse(user))
}

// LIST + FILTER
func AdminListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	role := strings.TrimSpace(c.Query("role", ""))
	q := strings.TrimSpace(c.Query("q", ""))

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
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to count users", err.Error())
	}

	var users []models.User
	if err := tx.Order("id DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve users", err.Error())
	}

	responses := make([]userdto.AdminUserResponse, 0, len(users))
	for i := range users {
		responses = append(responses, userdto.NewAdminUserResponse(users[i]))
	}

	meta := utils.PaginationMeta{Page: page, Limit: limit, Total: total}
	return utils.PaginatedResponse(c, fiber.StatusOK, "users retrieved successfully", responses, meta)
}

// Update API(partial)
func AdminUpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User
	if err := config.DB.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "user not found", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve user", err.Error())
	}

	var req userdto.AdminUserUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "invalid request body", err.Error())
	}

	if validationErrors := req.Validate(); len(validationErrors) > 0 {
		return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "validation error", validationErrors)
	}

	if req.Username != nil {
		user.Username = strings.TrimSpace(*req.Username)
	}
	if req.FirstName != nil {
		user.FirstName = strings.TrimSpace(*req.FirstName)
	}
	if req.LastName != nil {
		user.LastName = strings.TrimSpace(*req.LastName)
	}
	if req.Email != nil {
		user.Email = strings.TrimSpace(*req.Email)
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Jabatan != nil {
		user.Jabatan = strings.TrimSpace(*req.Jabatan)
	}
	if req.Atribut != nil {
		user.Atribut = *req.Atribut
	}
	if req.Password != nil {
		pwd := strings.TrimSpace(*req.Password)
		if pwd != "" {
			hash, err := bcryptHash(pwd)
			if err != nil {
				return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to hash password", nil)
			}
			user.PasswordHash = hash
		}
	}

	if err := config.DB.Save(&user).Error; err != nil {
		if isDupErr(err) {
			return utils.ErrorResponse(c, fiber.ErrBadRequest.Code, "username or email already exists", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to update user", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "user updated successfully", userdto.NewAdminUserResponse(user))
}

// Delete User API
func AdminDeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	result := config.DB.Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to delete user", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "user not found", nil)
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "user deleted successfully", nil)
}
