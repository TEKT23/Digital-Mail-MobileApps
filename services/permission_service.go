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
func (ps *PermissionService) CanUserCreateLetter(user *models.User, scope string, letterType models.LetterType) (bool, error) {
	if user == nil {
		return false, ErrUnauthorized
	}

	// ATURAN 1: Staf Program
	if user.Role == models.RoleStafProgram {
		if letterType == models.LetterKeluar && scope == models.ScopeEksternal {
			return true, nil
		}
		return false, nil
	}

	// ATURAN 2: Staf Lembaga
	if user.Role == models.RoleStafLembaga {
		if letterType == models.LetterMasuk {
			return true, nil
		}

		if letterType == models.LetterKeluar && scope == models.ScopeInternal {
			return true, nil
		}

		return false, nil
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
	// Pointer harus dideference dengan hati-hati
	if letter.AssignedVerifierID != nil {
		if *letter.AssignedVerifierID != user.ID {
			return false, nil
		}
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

	isCreator := letter.CreatedByID == user.ID
	isArchivist := user.Role == models.RoleStafLembaga

	if !isCreator && !isArchivist {
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

func (ps *PermissionService) CanUserViewLetter(user *models.User, letter *models.Letter) (bool, error) {
	if user == nil {
		return false, ErrUnauthorized
	}

	// 1. Admin & Direktur bisa lihat semua
	if user.Role == models.RoleAdmin || user.Role == models.RoleDirektur {
		return true, nil
	}

	// 2. Staf Lembaga (Archiver) bisa lihat SEMUA surat (internal & eksternal, masuk & keluar)
	if user.Role == models.RoleStafLembaga {
		return true, nil
	}

	// 3. Pembuat surat (Staf) bisa lihat suratnya sendiri
	if letter.CreatedByID == user.ID {
		return true, nil
	}

	// 4. Verifier (Manajer) bisa lihat jika ditugaskan kepadanya
	if letter.AssignedVerifierID != nil && *letter.AssignedVerifierID == user.ID {
		return true, nil
	}

	// 5. Manajer bisa lihat surat di Scope-nya (meski bukan verifier langsung, opsional)
	if user.IsManajer() {
		return user.CanVerifyScope(letter.Scope), nil
	}

	// 6. Staf Program bisa lihat surat Eksternal (scope bidangnya)
	if user.Role == models.RoleStafProgram && letter.Scope == models.ScopeEksternal {
		return true, nil
	}

	return false, nil
}
