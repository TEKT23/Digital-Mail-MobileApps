package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type LetterType string
type Priority string
type LetterStatus string

const (
	LetterMasuk    LetterType = "masuk"
	LetterKeluar   LetterType = "keluar"
	LetterInternal LetterType = "internal"
)

const (
	PriorityBiasa   Priority = "biasa"
	PrioritySegera  Priority = "segera"
	PriorityPenting Priority = "penting"
)

const (
	StatusDraft            LetterStatus = "draft"
	StatusPerluVerifikasi  LetterStatus = "perlu_verifikasi"
	StatusBelumDisposisi   LetterStatus = "belum_disposisi"
	StatusSudahDisposisi   LetterStatus = "sudah_disposisi"
	StatusPerluPersetujuan LetterStatus = "perlu_persetujuan"
	StatusPerluRevisi      LetterStatus = "perlu_revisi"
	StatusDisetujui        LetterStatus = "disetujui"
	StatusDiarsipkan       LetterStatus = "diarsipkan"
)

const (
	ScopeInternal  = "Internal"
	ScopeEksternal = "Eksternal"
)

type Letter struct {
	gorm.Model
	Pengirim         string     `gorm:"type:varchar(200);index"`
	NomorSurat       string     `gorm:"type:varchar(100);index"`
	NomorAgenda      string     `gorm:"type:varchar(100);index"`
	Disposisi        string     `gorm:"type:text"`
	TanggalDisposisi *time.Time `gorm:"type:datetime"`
	BidangTujuan     string     `gorm:"type:varchar(150);index"`

	JenisSurat LetterType `gorm:"type:enum('masuk','keluar','internal');not null;index"`
	Prioritas  Priority   `gorm:"type:enum('biasa','segera','penting');default:'biasa';not null;index"`

	// NEW FIELDS
	Scope              string `json:"scope" gorm:"type:enum('Internal','Eksternal');not null;index"`
	AssignedVerifierID *uint  `json:"assigned_verifier_id" gorm:"index"`
	AssignedVerifier   *User  `json:"assigned_verifier,omitempty" gorm:"foreignKey:AssignedVerifierID"`

	IsiSurat     string     `gorm:"type:longtext"`
	TanggalSurat *time.Time `gorm:"type:datetime"`
	TanggalMasuk *time.Time `gorm:"type:datetime;index"`
	JudulSurat   string     `gorm:"type:varchar(255);index"`
	Kesimpulan   string     `gorm:"type:text"`
	FilePath     string     `gorm:"type:varchar(255)"`

	Status LetterStatus `gorm:"type:enum('draft','perlu_verifikasi','belum_disposisi','sudah_disposisi','perlu_persetujuan','perlu_revisi','disetujui','diarsipkan');default:'draft';not null;index"`

	CreatedByID  uint  `gorm:"not null;index"`
	CreatedBy    *User `gorm:"foreignkey:CreatedByID"`
	VerifiedByID *uint `gorm:"index"`
	VerifiedBy   *User `gorm:"foreignkey:VerifiedByID"`
	DisposedByID *uint `gorm:"index"`
	DisposedBy   *User `gorm:"foreignkey:DisposedByID"`

	// Reply Linking Fields
	NeedsReply  bool     `gorm:"default:false;index" json:"needs_reply"` // Flag: surat masuk ini butuh balasan?
	InReplyToID *uint    `gorm:"index" json:"in_reply_to_id,omitempty"`  // FK: surat ini adalah balasan dari surat mana?
	InReplyTo   *Letter  `gorm:"foreignKey:InReplyToID" json:"in_reply_to,omitempty"`
	Replies     []Letter `gorm:"foreignKey:InReplyToID" json:"replies,omitempty"` // Surat-surat yang membalas surat ini
}

func (Letter) TableName() string { return "surat" }

func (l *Letter) IsSuratKeluar() bool { return l.JenisSurat == LetterKeluar }
func (l *Letter) IsSuratMasuk() bool  { return l.JenisSurat == LetterMasuk }

func (l *Letter) CanTransitionTo(newStatus LetterStatus) bool {
	validTransitions := map[LetterStatus][]LetterStatus{
		StatusDraft:            {StatusPerluVerifikasi, StatusBelumDisposisi},
		StatusPerluVerifikasi:  {StatusPerluPersetujuan, StatusPerluRevisi},
		StatusPerluPersetujuan: {StatusDisetujui, StatusPerluRevisi},
		StatusPerluRevisi:      {StatusPerluVerifikasi, StatusPerluPersetujuan, StatusDraft}, // Revisi boleh balik ke Draft
		StatusDisetujui:        {StatusDiarsipkan},
		StatusBelumDisposisi:   {StatusSudahDisposisi},
		StatusSudahDisposisi:   {StatusDiarsipkan},
	}

	allowed, exists := validTransitions[l.Status]
	if !exists {
		return false
	}
	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}

func (l *Letter) Validate() error {
	if l.IsSuratKeluar() && l.AssignedVerifierID == nil {
		return errors.New("surat keluar harus memiliki assigned verifier")
	}
	return nil
}
