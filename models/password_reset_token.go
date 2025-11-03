package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrPasswordResetTokenExpired = errors.New("password reset token expired")
	ErrPasswordResetTokenUsed    = errors.New("password reset token already used")
)

const PasswordResetTokenTTL = time.Hour

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

func (t PasswordResetToken) IsExpired(reference time.Time) bool {
	if reference.IsZero() {
		reference = time.Now()
	}
	return !reference.Before(t.ExpiresAt)
}

func (t PasswordResetToken) Validate(reference time.Time) error {
	if reference.IsZero() {
		reference = time.Now()
	}
	if t.Used {
		return ErrPasswordResetTokenUsed
	}
	if t.IsExpired(reference) {
		return ErrPasswordResetTokenExpired
	}
	return nil
}

func (t *PasswordResetToken) Consume(tx *gorm.DB, reference time.Time) error {
	if reference.IsZero() {
		reference = time.Now()
	}

	if err := t.Validate(reference); err != nil {
		return err
	}

	usedAt := reference
	updates := map[string]any{
		"used":    true,
		"used_at": &usedAt,
	}

	res := tx.Model(&PasswordResetToken{}).
		Where("id = ? AND used = ? AND expires_at > ?", t.ID, false, reference).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		var latest PasswordResetToken
		if err := tx.First(&latest, t.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPasswordResetTokenUsed
			}
			return err
		}

		if latest.Used {
			return ErrPasswordResetTokenUsed
		}
		if latest.IsExpired(reference) {
			return ErrPasswordResetTokenExpired
		}
		return ErrPasswordResetTokenUsed
	}

	t.Used = true
	t.UsedAt = &usedAt
	return nil
}
