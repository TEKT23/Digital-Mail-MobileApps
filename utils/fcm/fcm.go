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
)

const (
	FCMTopicPrefix = "digitalmail_role_"
)

var (
	fcmClient *messaging.Client
)

func init() {
	log.Println("Initializing Firebase Admin SDK...")

	ctx := context.Background()

	config := &firebase.Config{
		ProjectID: "digimail-mobile",
	}
	app, err := firebase.NewApp(ctx, config)
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

				// surat baru dibuat
				case events.LetterCreated:
					log.Printf("Event: Letter Created (ID: %d). No notification sent.", e.Letter.ID)

				// kalau status surat berubah
				case events.LetterStatusMoved:
					log.Printf("Event: Letter Status Moved (ID: %d, New Status: %s)", e.Letter.ID, e.Letter.Status)

					var targetRole models.Role
					var title, body string

					//
					switch e.Letter.Status {

					// Dari Draft -> Perlu Disposisi (Target: ADC)
					case models.StatusPerluVerifikasi: // <-- Nama status baru
						title = "Surat Perlu Verifikasi"
						data := map[string]string{
							"letter_id": letterIDStr,
							"action":    "status_change",
							"status":    string(e.Letter.Status),
						}

						bodyADC := fmt.Sprintf("Surat \"%s\" (No. %s) memerlukan verifikasi.", e.Letter.JudulSurat, e.Letter.NomorSurat)
						topicADC := mapRoleToTopic(models.RoleADC)
						if err := SendNotificationToTopic(sendCtx, topicADC, title, bodyADC, data); err != nil {
							log.Printf("Error sending FCM to ADC: %v", err)
						}

						bodyDirektur := fmt.Sprintf("Surat masuk \"%s\" (No. %s) sedang diverifikasi ADC.", e.Letter.JudulSurat, e.Letter.NomorSurat)
						topicDirektur := mapRoleToTopic(models.RoleDirektur)
						if err := SendNotificationToTopic(sendCtx, topicDirektur, title, bodyDirektur, data); err != nil {
							log.Printf("Error sending FCM to Direktur: %v", err)
						}

						return

					// Dari Perlu Disposisi -> Belum Disposisi (Target: Direktur)
					case models.StatusBelumDisposisi:
						targetRole = models.RoleDirektur //
						title = "Surat Siap Disposisi"
						body = fmt.Sprintf("Surat \"%s\" (No. %s) siap untuk disposisi Anda.", e.Letter.JudulSurat, e.Letter.NomorSurat)

					case models.StatusSudahDisposisi:
						title = "Surat Selesai Disposisi"
						body = fmt.Sprintf("Surat \"%s\" (No. %s) telah selesai didisposisi.", e.Letter.JudulSurat, e.Letter.NomorSurat)
						data := map[string]string{
							"letter_id": letterIDStr,
							"action":    "status_change",
							"status":    string(e.Letter.Status),
						}

						// notif ke adc
						topicADC := mapRoleToTopic(models.RoleADC)
						if err := SendNotificationToTopic(sendCtx, topicADC, title, body, data); err != nil {
							log.Printf("Error sending FCM to ADC: %v", err)
						}

						// notif ke bagian umum
						topicBU := mapRoleToTopic(models.RoleBagianUmum)
						if err := SendNotificationToTopic(sendCtx, topicBU, title, body, data); err != nil {
							log.Printf("Error sending FCM to Bagian Umum: %v", err)
						}

					case models.StatusPerluPersetujuan:
						targetRole = models.RoleDirektur //
						title = "Surat Keluar Perlu Persetujuan"
						body = fmt.Sprintf("Surat keluar \"%s\" memerlukan persetujuan Anda.", e.Letter.JudulSurat)

					case models.StatusPerluRevisi:
						targetRole = models.RoleADC //
						title = "Surat Keluar Perlu Revisi"
						body = fmt.Sprintf("Surat keluar \"%s\" perlu revisi dari Direktur.", e.Letter.JudulSurat)

					case models.StatusDisetujui:
						targetRole = models.RoleADC //
						title = "Surat Keluar Disetujui"
						body = fmt.Sprintf("Surat keluar \"%s\" telah disetujui.", e.Letter.JudulSurat)

					default:
						return
					}

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
			log.Println("FCM Notifier Consumer stopped.")
			return
		}
	}
}
