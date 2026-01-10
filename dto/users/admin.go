package users

import (
	"strings"
	"time"

	"TugasAkhir/models"
)

type AdminUserCreateRequest struct {
	Username  string      `json:"username"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
	Password  string      `json:"password"`
	Role      models.Role `json:"role"`
	Jabatan   string      `json:"jabatan"`
	Atribut   string      `json:"atribut"`
}

type AdminUserUpdateRequest struct {
	Username  *string      `json:"username"`
	FirstName *string      `json:"first_name"`
	LastName  *string      `json:"last_name"`
	Email     *string      `json:"email"`
	Password  *string      `json:"password"`
	Role      *models.Role `json:"role"`
	Jabatan   *string      `json:"jabatan"`
	Atribut   *string      `json:"atribut"`
}

type AdminUserResponse struct {
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

func (r *AdminUserCreateRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(r.Username) == "" {
		errors["username"] = "username is required"
	}
	if strings.TrimSpace(r.Email) == "" {
		errors["email"] = "email is required"
	}
	if strings.TrimSpace(r.Password) == "" {
		errors["password"] = "password is required"
	} else if len(r.Password) < 8 {
		errors["password"] = "password must be at least 8 characters"
	}
	if !isValidRole(r.Role) {
		errors["role"] = "role must be bagian_umum, adc, direktur, or admin"
	}

	return errors
}

func (r *AdminUserUpdateRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if r.Password != nil {
		pwd := strings.TrimSpace(*r.Password)
		if pwd != "" && len(pwd) < 8 {
			errors["password"] = "password must be at least 8 characters"
		}
	}
	if r.Role != nil && !isValidRole(*r.Role) {
		errors["role"] = "role must be bagian_umum, adc, direktur, or admin"
	}

	return errors
}

func isValidRole(role models.Role) bool {
	switch role {
	// Daftar Role Lengkap (Baru)
	case models.RoleAdmin,
		models.RoleDirektur,
		models.RoleStafProgram,
		models.RoleStafLembaga,
		models.RoleManajerKPP,
		models.RoleManajerPemas,
		models.RoleManajerPKL:
		return true
	default:
		return false
	}
}

func NewAdminUserResponse(user models.User) AdminUserResponse {
	return AdminUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
		Jabatan:   user.Jabatan,
		Atribut:   user.Atribut,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
