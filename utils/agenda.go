package utils

import (
	"TugasAkhir/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// GenerateNomorAgenda generates the next agenda number for a given letter type.
// It uses row-level locking (FOR UPDATE) to prevent race conditions when
// multiple users create letters simultaneously.
//
// Features:
// - Independent sequences for masuk/keluar (filtered by jenis_surat)
// - Yearly reset (filtered by YEAR(created_at))
// - Race condition protection via FOR UPDATE
//
// Returns empty string if there's an error, caller should check error.
func GenerateNomorAgenda(tx *gorm.DB, jenisSurat models.LetterType) (string, error) {
	var lastSeq int
	currentYear := time.Now().Year()

	// Use raw SQL with FOR UPDATE to lock rows and prevent race conditions
	// COALESCE handles the case when table is empty (returns 0)
	// Filter nomor_agenda != '' to exclude drafts that never got a number
	err := tx.Raw(`
		SELECT COALESCE(MAX(CAST(nomor_agenda AS UNSIGNED)), 0) 
		FROM surat 
		WHERE jenis_surat = ? AND YEAR(created_at) = ? AND nomor_agenda != ''
		FOR UPDATE
	`, jenisSurat, currentYear).Scan(&lastSeq).Error

	if err != nil {
		return "", err
	}

	return strconv.Itoa(lastSeq + 1), nil
}
