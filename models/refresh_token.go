package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	gorm.Model
	Token     string    `gorm:"type:varchar(512);uniqueIndex;not null"`
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"constraint:OnDelete:CASCADE;"`
	ExpiresAt time.Time `gorm:"not null;index"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
