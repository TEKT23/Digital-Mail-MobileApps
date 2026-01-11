package models

import "gorm.io/gorm"

type Role string

const (
	RoleAdmin        Role = "admin"
	RoleDirektur     Role = "direktur"
	RoleStafProgram  Role = "staf_program" // Pengganti ADC
	RoleStafLembaga  Role = "staf_lembaga" // Pengganti Bagian Umum
	RoleManajerKPP   Role = "manajer_kpp"
	RoleManajerPemas Role = "manajer_pemas"
	RoleManajerPKL   Role = "manajer_pkl"
)

type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	FirstName    string `gorm:"type:varchar(100)"`
	LastName     string `gorm:"type:varchar(100)"`
	Email        string `gorm:"type:varchar(191);uniqueIndex;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	Role         Role   `gorm:"type:enum('admin','direktur','staf_program','staf_lembaga','manajer_kpp','manajer_pemas','manajer_pkl');not null;index"`
	Jabatan      string `gorm:"type:varchar(150)"`
	Atribut      string `gorm:"type:text"`
}

func (User) TableName() string {
	return "users"
}

// --- Helper Methods ---

func (u *User) IsStaf() bool {
	return u.Role == RoleStafProgram || u.Role == RoleStafLembaga
}

func (u *User) IsManajer() bool {
	return u.Role == RoleManajerKPP || u.Role == RoleManajerPemas || u.Role == RoleManajerPKL
}

func (u *User) CanVerifyScope(scope string) bool {
	if scope == "Eksternal" {
		return u.Role == RoleManajerKPP || u.Role == RoleManajerPemas
	}
	if scope == "Internal" {
		return u.Role == RoleManajerPKL
	}
	return false
}

func (u *User) IsDirektur() bool { return u.Role == RoleDirektur }
func (u *User) IsAdmin() bool    { return u.Role == RoleAdmin }

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleDirektur, RoleStafProgram, RoleStafLembaga, RoleManajerKPP, RoleManajerPemas, RoleManajerPKL:
		return true
	default:
		return false
	}
}
