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

const FCMTopicPrefix = "topic_"

var fcmClient *messaging.Client

// Init Firebase
func init() {
	log.Println("🔥 Initializing Firebase Admin SDK...")
	ctx := context.Background()
	config := &firebase.Config{ProjectID: "digimail-mobile"}

	app, err := firebase.NewApp(ctx, config)
	if err != nil {
		log.Fatalf("❌ Error initializing Firebase app: %v\n", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("❌ Error getting Firebase Messaging client: %v\n", err)
	}

	fcmClient = client
}

func mapRoleToTopic(role models.Role) string {
	return FCMTopicPrefix + string(role)
}

func SendNotificationToTopic(ctx context.Context, topic, title, body string, data map[string]string) {
	if fcmClient == nil {
		return
	}
	msg := &messaging.Message{
		Topic:        topic,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data:         data,
		Android: &messaging.AndroidConfig{
			Priority:     "high",
			Notification: &messaging.AndroidNotification{ChannelID: "default_channel"},
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
		if letter.IsSuratMasuk() {
			notifyDirektur(ctx, letter, "Surat Masuk Baru", data)
		}
		// Jika langsung verifikasi
		if letter.Status == models.StatusPerluVerifikasi {
			notifySpecificManajer(ctx, letter, data)
		}

	case events.LetterStatusMoved:
		switch letter.Status {
		case models.StatusPerluVerifikasi:
			notifySpecificManajer(ctx, letter, data)

		case models.StatusPerluPersetujuan:
			notifyDirektur(ctx, letter, "Butuh Persetujuan", data)

		case models.StatusPerluRevisi:
			notifyStaf(ctx, letter, "Surat Perlu Revisi", "Mohon cek catatan revisi.", data)

		case models.StatusDisetujui:
			notifyStaf(ctx, letter, "Surat Disetujui", "Surat Anda telah disetujui Direktur.", data)
			notifyArchivist(ctx, letter, data) // CC ke Arsip

		case models.StatusSudahDisposisi:
			notifyStaf(ctx, letter, "Disposisi Turun", "Direktur telah memberikan disposisi.", data)
		}
	}
}

func notifySpecificManajer(ctx context.Context, l models.Letter, data map[string]string) {
	title := "Verifikasi Surat Baru"
	body := fmt.Sprintf("Surat #%s menunggu verifikasi Anda.", l.NomorSurat)

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
	title := "Arsip Surat Keluar"
	body := fmt.Sprintf("Surat #%s telah final. Silakan diarsipkan.", l.NomorSurat)
	SendNotificationToTopic(ctx, mapRoleToTopic(models.RoleStafLembaga), title, body, data)
}

func notifyDirektur(ctx context.Context, l models.Letter, title string, data map[string]string) {
	body := fmt.Sprintf("Surat dari %s menunggu disposisi.", l.Pengirim)
	if l.IsSuratKeluar() {
		body = fmt.Sprintf("Surat Keluar #%s menunggu tanda tangan.", l.NomorSurat)
	}
	SendNotificationToTopic(ctx, mapRoleToTopic(models.RoleDirektur), title, body, data)
}

func notifyStaf(ctx context.Context, l models.Letter, title, body string, data map[string]string) {
	targetRole := models.RoleStafProgram
	if l.Scope == models.ScopeInternal {
		targetRole = models.RoleStafLembaga
	}
	SendNotificationToTopic(ctx, mapRoleToTopic(targetRole), title, body, data)
}
