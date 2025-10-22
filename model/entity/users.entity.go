package entity

import "gorm.io/gorm"

type Role string

const (
	RoleBagianUmum Role = "bagian_umum"
	RoleADC        Role = "adc"
	RoleDirektur   Role = "direktur"
	RoleAdmin      Role = "admin"
)

type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	FirstName    string `gorm:"type:varchar(100)"`
	LastName     string `gorm:"type:varchar(100)"`
	Email        string `gorm:"type:varchar(191);uniqueIndex;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	Role         Role   `gorm:"type:ENUM('bagian_umum','adc','direktur','admin');not null;index"`
	Jabatan      string `gorm:"type:varchar(150)"`
	Atribut      string `gorm:"type:text"`
}

func (User) TableName() string {
	return "users"
}
