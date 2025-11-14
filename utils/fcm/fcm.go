package fcm

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"TugasAkhir/models" //
	"TugasAkhir/utils/events"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

const (
	// FCMTopicPrefix adalah prefix untuk semua topik berbasis role
	FCMTopicPrefix = "digitalmail_role_"

	// Path ke service account key Anda.
	// Asumsi file ini ada di root folder proyek Anda.
	ServiceAccountKeyPath = "D:\\Development\\Sistem Penyuratan Digital\\digimail-mobile-firebase-adminsdk-fbsvc-77c65bbf85.json"
)

var (
	// fcmClient adalah client Firebase Messaging yang diinisialisasi secara global
	fcmClient *messaging.Client
)

// init() adalah fungsi khusus Go yang berjalan sekali saat package ini di-load.
// Kita gunakan ini untuk inisialisasi Firebase Admin SDK secara global.
func init() {
	log.Println("Initializing Firebase Admin SDK...")

	opt := option.WithCredentialsFile(ServiceAccountKeyPath)

	// Gunakan context background untuk inisialisasi
	ctx := context.Background()

	config := &firebase.Config{
		ProjectID: "digimail-mobile",
	}

	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v\n", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Firebase Messaging client: %v\n", err)
	}

	fcmClient = client
	log.Println("✅ Firebase Admin SDK initialized successfully.")
}

// mapRoleToTopic mengembalikan nama topik FCM untuk role tertentu
func mapRoleToTopic(role models.Role) string {
	return FCMTopicPrefix + string(role)
}

// SendNotificationToTopic mengirimkan notifikasi ke topik FCM tertentu
func SendNotificationToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	if fcmClient == nil {
		return fmt.Errorf("FCM client not initialized")
	}

	// Membuat payload pesan
	msg := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		// Opsi Android untuk prioritas dan channel (jika diperlukan)
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID: "default_channel", // Pastikan channel ini ada di aplikasi Android
			},
		},
		// Opsi APNS (iOS)
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	// Mengirim pesan
	response, err := fcmClient.Send(ctx, msg)
	if err != nil {
		log.Printf("Error sending FCM to topic %s: %v", topic, err)
		return err
	}

	log.Printf("Successfully sent message to topic %s: %s", topic, response)
	return nil
}

// StartNotifierConsumer adalah goroutine yang mendengarkan event bus
func StartNotifierConsumer(ctx context.Context) {
	log.Println("✅ FCM Notifier Consumer started")

	for {
		select {
		case event := <-events.LetterEventBus:
			// Proses setiap event di goroutine baru agar tidak memblokir consumer
			go func(e events.LetterEvent) {

				// Buat context dengan timeout untuk operasi pengiriman
				sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				letterIDStr := strconv.FormatUint(uint64(e.Letter.ID), 10)

				switch e.Type {

				// --- KASUS: SURAT BARU DIBUAT ---
				case events.LetterCreated:
					log.Printf("Event: Letter Created (ID: %d)", e.Letter.ID)

					title := "Surat Baru Dibuat"
					data := map[string]string{
						"letter_id": letterIDStr,
						"action":    "letter_created",
						"status":    string(e.Letter.Status),
					}

					// Target 1: ADC (Perlu Verifikasi)
					//
					bodyADC := fmt.Sprintf("Surat masuk \"%s\" memerlukan verifikasi Anda.", e.Letter.JudulSurat)
					topicADC := mapRoleToTopic(models.RoleADC)
					if err := SendNotificationToTopic(sendCtx, topicADC, title, bodyADC, data); err != nil {
						log.Printf("Error sending FCM to ADC: %v", err)
					}

					// Target 2: Direktur (Hanya Pemberitahuan)
					//
					bodyDirektur := fmt.Sprintf("Telah masuk surat baru: \"%s\".", e.Letter.JudulSurat)
					topicDirektur := mapRoleToTopic(models.RoleDirektur)
					if err := SendNotificationToTopic(sendCtx, topicDirektur, title, bodyDirektur, data); err != nil {
						log.Printf("Error sending FCM to Direktur: %v", err)
					}

				// --- KASUS: STATUS SURAT BERUBAH ---
				case events.LetterStatusMoved:
					log.Printf("Event: Letter Status Moved (ID: %d, New Status: %s)", e.Letter.ID, e.Letter.Status)

					var targetRole models.Role
					var title, body string

					//
					switch e.Letter.Status {

					// Dari Draft -> Perlu Disposisi (Target: ADC)
					case models.StatusPerluDisposisi:
						targetRole = models.RoleADC //
						title = "Surat Perlu Verifikasi"
						body = fmt.Sprintf("Surat \"%s\" (No. %s) memerlukan verifikasi Anda.", e.Letter.JudulSurat, e.Letter.NomorSurat)

					// Dari Perlu Disposisi -> Belum Disposisi (Target: Direktur)
					case models.StatusBelumDisposisi:
						targetRole = models.RoleDirektur //
						title = "Surat Siap Disposisi"
						body = fmt.Sprintf("Surat \"%s\" (No. %s) siap untuk disposisi Anda.", e.Letter.JudulSurat, e.Letter.NomorSurat)

					default:
						// Abaikan perubahan status lainnya (misal: 'sudah_disposisi' atau 'draft')
						return
					}

					// Kirim notifikasi jika ada target
					if targetRole != "" {
						topic := mapRoleToTopic(targetRole)
						data := map[string]string{
							"letter_id": letterIDStr,
							"action":    "status_change",
							"status":    string(e.Letter.Status),
						}
						if err := SendNotificationToTopic(sendCtx, topic, title, body, data); err != nil {
							log.Printf("Error sending FCM to %s: %v", targetRole, err)
						}
					}
				}
			}(event)

		case <-ctx.Done():
			// Jika context dibatalkan (misal: server shutdown), hentikan consumer
			log.Println("FCM Notifier Consumer stopped.")
			return
		}
	}
}
