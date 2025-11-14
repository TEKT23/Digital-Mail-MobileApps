package events

import (
	"TugasAkhir/models" //
)

// LetterEventType mendefinisikan jenis event terkait siklus hidup surat
type LetterEventType string

const (
	// LetterCreated dipublikasikan saat surat baru berhasil dibuat
	LetterCreated LetterEventType = "LetterCreated"

	// LetterStatusMoved dipublikasikan saat status surat berubah
	// (misalnya, dari draft ke perlu_disposisi)
	LetterStatusMoved LetterEventType = "LetterStatusMoved"
)

// LetterEvent adalah payload untuk event surat
type LetterEvent struct {
	Type      LetterEventType
	Letter    models.Letter
	OldStatus models.LetterStatus // Status lama (hanya relevan untuk LetterStatusMoved)
}

// LetterEventBus adalah channel untuk menangani event surat.
// Channel ini di-buffer untuk mencegah blocking pada handler API
// saat mempublikasikan event.
var LetterEventBus = make(chan LetterEvent, 100)
