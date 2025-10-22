package models

import (
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
	StatusDraft          LetterStatus = "draft"
	StatusPerluDisposisi LetterStatus = "perlu_disposisi"
	StatusBelumDisposisi LetterStatus = "belum_disposisi"
	StatusSudahDisposisi LetterStatus = "sudah_disposisi"
)

type Letter struct {
	gorm.Model
	IDSurat          int        `gorm:"primaryKey;autoIncrement:true"`
	Pengirim         string     `gorm:"type:varchar(200);index"`
	NomorSurat       string     `gorm:"type:varchar(100);index"`
	NomorAgenda      string     `gorm:"type:varchar(100);index"`
	Disposisi        string     `gorm:"type:text"`
	TanggalDisposisi *time.Time `gorm:"type:date"`
	BidangTujuan     string     `gorm:"type:varchar(150);index"`

	JenisSurat LetterType `gorm:"type:ENUM('masuk','keluar','internal');not null;index"`
	Prioritas  Priority   `gorm:"type:ENUM('biasa','segera','penting');default:'biasa';not null;index"`

	IsiSurat     string     `gorm:"type:longtext"`
	TanggalSurat *time.Time `gorm:"type:date"`
	TanggalMasuk *time.Time `gorm:"type:date;index"`
	JudulSurat   string     `gorm:"type:varchar(255);index"`
	Kesimpulan   string     `gorm:"type:text"`
	FilePath     string     `gorm:"type:varchar(255)"`

	Status LetterStatus `gorm:"type:ENUM('draft','perlu_disposisi','belum_disposisi','sudah_disposisi');default:'draft';not null;index"`

	//relation
	CreatedByID  *uint `gorm:"index"` // Bagian Umum
	VerifiedByID *uint `gorm:"index"` // ADC
	DisposedByID *uint `gorm:"index"` // Direktur
}

func (Letter) TableName() string {
	return "surat"
}
