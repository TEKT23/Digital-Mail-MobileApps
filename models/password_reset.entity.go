package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordResetToken struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	TokenHash string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"not null;default:false"`
	UsedAt    *time.Time

	User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
