package services

import (
	"TugasAkhir/models"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrUnauthorized = errors.New("unauthorized: user not authenticated")
	ErrForbidden    = errors.New("forbidden: insufficient permissions")
	ErrNotFound     = errors.New("resource not found")
)

type PermissionService struct {
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// CanUserCreateLetter - Cek izin membuat surat
func (ps *PermissionService) CanUserCreateLetter(user *models.User, scope string) (bool, error) {
	if user == nil {
		return false, ErrUnauthorized
	}
	if scope == models.ScopeEksternal {
		return user.Role == models.RoleStafProgram, nil
	}
	if scope == models.ScopeInternal {
		return user.Role == models.RoleStafLembaga, nil
	}
	return false, ErrForbidden
}

// CanUserVerifyLetter - Cek apakah Manajer boleh verifikasi
func (ps *PermissionService) CanUserVerifyLetter(user *models.User, letter *models.Letter) (bool, error) {
	if user == nil {
		return false, ErrUnauthorized
	}
	if letter == nil {
		return false, ErrNotFound
	}

	// 1. User harus Manajer
	if !user.IsManajer() {
		return false, nil
	}

	// 2. Scope sesuai
	if !user.CanVerifyScope(letter.Scope) {
		return false, nil
	}

	// 3. Verifier ID cocok (jika ada)
	if letter.AssignedVerifierID != nil && *letter.AssignedVerifierID != user.ID {
		return false, nil
	}

	// 4. Status harus perlu_verifikasi
	if letter.Status != models.StatusPerluVerifikasi {
		return false, nil
	}

	return true, nil
}

// CanUserApproveLetter - Cek apakah Direktur boleh approve
func (ps *PermissionService) CanUserApproveLetter(user *models.User, letter *models.Letter) (bool, error) {
	if user == nil {
		return false, ErrUnauthorized
	}

	if !user.IsDirektur() {
		return false, nil
	}

	if letter.Status != models.StatusPerluPersetujuan {
		return false, nil
	}

	return true, nil
}

// === BAGIAN INI YANG TADI MUNGKIN HILANG/ERROR ===
// CanUserDisposeLetter - Cek apakah Direktur boleh disposisi surat masuk
func (ps *PermissionService) CanUserDisposeLetter(user *models.User, letter *models.Letter) (bool, error) {
	if user == nil {
		return false, ErrUnauthorized
	}

	// 1. Hanya Direktur
	if !user.IsDirektur() {
		return false, nil
	}

	// 2. Harus Surat Masuk (Helper method ini ada di models/letters.entity.go)
	if !letter.IsSuratMasuk() {
		return false, nil
	}

	// 3. Status harus 'belum_disposisi'
	if letter.Status != models.StatusBelumDisposisi {
		return false, nil
	}

	return true, nil
}

// =================================================

// CanUserArchiveLetter - Cek apakah Staf boleh arsip
func (ps *PermissionService) CanUserArchiveLetter(user *models.User, letter *models.Letter) (bool, error) {
	if user == nil {
		return false, ErrUnauthorized
	}

	if letter.CreatedByID != user.ID {
		return false, nil
	}

	if letter.Status != models.StatusDisetujui && letter.Status != models.StatusSudahDisposisi {
		return false, nil
	}

	return true, nil
}

// GetLetterByID - Helper fetch letter
func (ps *PermissionService) GetLetterByID(id uint) (*models.Letter, error) {
	var letter models.Letter
	err := ps.db.
		Preload("AssignedVerifier").
		Preload("CreatedBy").
		First(&letter, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &letter, nil
}
