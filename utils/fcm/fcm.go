package fcm

import (
	"TugasAkhir/models"
	"TugasAkhir/utils/events"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

const FCMTopicPrefix = "digitalmail_role_"

var fcmClient *messaging.Client

// InitializeFCM initializes the Firebase Admin SDK
// This must be called AFTER loading .env
func InitializeFCM() {
	log.Println("🔥 Initializing Firebase Admin SDK...")
	ctx := context.Background()
	config := &firebase.Config{ProjectID: "digimail-mobile"}

	// Gunakan GOOGLE_APPLICATION_CREDENTIALS dari environment
	// Pastikan variable ini sudah diset sebelum fungsi ini dipanggil
	app, err := firebase.NewApp(ctx, config)
	if err != nil {
		log.Printf("❌ Error initializing Firebase app: %v\n", err)
		return
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("❌ Error getting Firebase Messaging client: %v\n", err)
		return
	}

	fcmClient = client
	log.Println("✅ Firebase Admin SDK Initialized Successfully")
}

func mapRoleToTopic(role models.Role) string {
	return FCMTopicPrefix + string(role)
}

func SendNotificationToTopic(ctx context.Context, topic, title, body string, data map[string]string) {
	if fcmClient == nil {
		return
	}

	// Add title and body to data payload so Android always receives them via onMessageReceived
	data["title"] = title
	data["body"] = body

	// Use data-only message (no Notification payload) to ensure onMessageReceived is always called
	msg := &messaging.Message{
		Topic: topic,
		Data:  data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	}
	response, err := fcmClient.Send(ctx, msg)
	if err != nil {
		log.Printf("❌ GAGAL kirim notif ke [%s]: %v\n", topic, err)
	} else {
		log.Printf("✅ SUKSES kirim notif ke [%s]. ID: %s\n", topic, response)
	}
}

func StartNotifierConsumer(ctx context.Context) {
	log.Println("🚀 FCM Consumer Running...")
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-events.LetterEventBus:
			go handleEvent(event)
		}
	}
}

func handleEvent(event events.LetterEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	letter := event.Letter
	log.Printf("📨 Processing Event: %s | ID: %d | Status: %s\n", event.Type, letter.ID, letter.Status)

	data := map[string]string{
		"letter_id": strconv.FormatUint(uint64(letter.ID), 10),
		"status":    string(letter.Status),
		"type":      string(letter.JenisSurat),
	}

	switch event.Type {
	case events.LetterCreated:
		// 1. Surat Masuk Baru -> Direktur
		if letter.IsSuratMasuk() {
			title := fmt.Sprintf("Surat Masuk: %s", letter.Pengirim)
			body := fmt.Sprintf("Perihal: %s", truncateString(letter.JudulSurat, 50))
			notifyDirektur(ctx, title, body, data)
		}

		// Jika langsung verifikasi (Auto-Assign Manajer PKL)
		if letter.Status == models.StatusPerluVerifikasi {
			notifySpecificManajer(ctx, letter, data)
		}

	case events.LetterStatusMoved:
		switch letter.Status {
		case models.StatusPerluVerifikasi:
			// 2. Verifikasi -> Manajer
			notifySpecificManajer(ctx, letter, data)

		case models.StatusPerluPersetujuan:
			// 3. Butuh Persetujuan -> Direktur
			title := "Butuh Tanda Tangan"
			body := fmt.Sprintf("Surat Keluar #%s menunggu persetujuan Anda.", letter.NomorSurat)
			notifyDirektur(ctx, title, body, data)

		case models.StatusPerluRevisi:
			// 4. Perlu Revisi -> Staf/Pembuat
			// Infer role penolak berdasarkan previous status
			rolePenolak := "Pimpinan" // Default
			if event.OldStatus == models.StatusPerluVerifikasi {
				rolePenolak = "Manajer"
			} else if event.OldStatus == models.StatusPerluPersetujuan {
				rolePenolak = "Direktur"
			}

			title := "Revisi Diperlukan"
			body := fmt.Sprintf("Surat #%s dikembalikan oleh %s. Cek catatan revisi.", letter.NomorSurat, rolePenolak)
			notifyStaf(ctx, letter, title, body, data)

		case models.StatusDiarsipkan:
			// 6. Surat Final/Arsip -> Staf
			title := "Surat Selesai & Diarsipkan"
			body := fmt.Sprintf("Surat #%s telah selesai diproses dan diarsipkan.", letter.NomorSurat)

			// 1. Beritahu Pembuat (Staf Program / Staf Lembaga)
			notifyStaf(ctx, letter, title, body, data)

			// 2. Beritahu Bagian Arsip (Staf Lembaga) - Jika pembuatnya bukan Staf Lembaga
			// Logic notifyArchivist bisa kita panggil agar Staf Lembaga (Admin Arsip) tau ada surat baru masuk arsip
			notifyArchivist(ctx, letter, data)

		case models.StatusSudahDisposisi:
			// 5. Disposisi Turun -> Staf
			title := "Disposisi Baru"
			body := fmt.Sprintf("Direktur telah mendisposisikan surat dari %s. Segera tindak lanjuti.", letter.Pengirim)
			notifyStaf(ctx, letter, title, body, data)
		}
	}
}

// truncateString memotong string jika lebih panjang dari limit
func truncateString(str string, limit int) string {
	if len(str) <= limit {
		return str
	}
	return str[:limit] + "..."
}

func notifySpecificManajer(ctx context.Context, l models.Letter, data map[string]string) {
	// 2. Verifikasi (Target: Manajer)
	title := "Verifikasi Surat Keluar"
	body := fmt.Sprintf("Surat #%s perihal '%s' menunggu verifikasi Anda.", l.NomorSurat, truncateString(l.JudulSurat, 30))

	// 1. Cek Data Object Manajer (Hasil Preload)
	if l.AssignedVerifier != nil {
		topic := mapRoleToTopic(l.AssignedVerifier.Role)
		SendNotificationToTopic(ctx, topic, title, body, data)
		return
	}

	// 2. Cek ID (Safety Net)
	if l.AssignedVerifierID != nil {
		log.Printf("❌ ERROR CRITICAL: Surat ID %d punya VerifierID %d tapi Struct User Kosong/Nil. Notifikasi TIDAK dikirim untuk mencegah salah alamat. Cek Preload di Handler!", l.ID, *l.AssignedVerifierID)
		return
	}

	// 3. Fallback Terakhir (Hanya jika benar-benar tidak ada ID sama sekali)
	log.Println("⚠️ Warning: Surat tidak memiliki Verifier ID sama sekali.")
}

func notifyArchivist(ctx context.Context, l models.Letter, data map[string]string) {
	// 6. Surat Final/Arsip (Target: Archivist/Staf Lembaga)
	title := "Surat Selesai & Diarsipkan"
	body := fmt.Sprintf("Surat #%s telah selesai diproses dan diarsipkan.", l.NomorSurat)
	SendNotificationToTopic(ctx, mapRoleToTopic(models.RoleStafLembaga), title, body, data)
}

func notifyDirektur(ctx context.Context, title, body string, data map[string]string) {
	SendNotificationToTopic(ctx, mapRoleToTopic(models.RoleDirektur), title, body, data)
}

func notifyStaf(ctx context.Context, l models.Letter, title, body string, data map[string]string) {
	targetRole := models.RoleStafProgram
	if l.Scope == models.ScopeInternal {
		targetRole = models.RoleStafLembaga
	}
	SendNotificationToTopic(ctx, mapRoleToTopic(targetRole), title, body, data)
}
