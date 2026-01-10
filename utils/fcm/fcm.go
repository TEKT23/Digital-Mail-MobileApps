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

// Prefix untuk nama topic di Firebase
const FCMTopicPrefix = "topic_"

var (
	fcmClient *messaging.Client
)

func init() {
	log.Println("Initializing Firebase Admin SDK...")
	ctx := context.Background()
	config := &firebase.Config{ProjectID: "digimail-mobile"}

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

// ---------------------------------------------------------
// HELPER: Map Role ke Topic (INI YANG SEBELUMNYA ERROR)
// ---------------------------------------------------------
// Fungsi ini sekarang DIPANGGIL di dalam StartNotifierConsumer
// untuk menentukan tujuan notifikasi secara dinamis.
func mapRoleToTopic(role models.Role) string {
	// Contoh: role "direktur" -> "topic_direktur"
	return FCMTopicPrefix + string(role)
}

// Helper kirim notifikasi
func SendNotificationToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	if fcmClient == nil {
		return fmt.Errorf("FCM client not initialized")
	}

	msg := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority:     "high",
			Notification: &messaging.AndroidNotification{ChannelID: "default_channel"},
		},
	}

	_, err := fcmClient.Send(ctx, msg)
	return err
}

func StartNotifierConsumer(ctx context.Context) {
	log.Println("✅ FCM Notifier Consumer started")

	for {
		select {
		case <-ctx.Done():
			return
		case e := <-events.LetterEventBus:

			// Goroutine agar tidak blocking
			go func(event events.LetterEvent) {
				sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// Data standar untuk payload notifikasi (dikirim ke HP)
				data := map[string]string{
					"letter_id": strconv.FormatUint(uint64(event.Letter.ID), 10),
					"status":    string(event.Letter.Status),
					"type":      string(event.Letter.JenisSurat),
				}

				// === LOGIC PENGIRIMAN NOTIFIKASI ===

				switch event.Type {

				// KASUS 1: Surat Baru Dibuat
				case events.LetterCreated:
					if event.Letter.IsSuratMasuk() {
						// Surat Masuk -> Kirim ke DIREKTUR
						// PENGGUNAAN mapRoleToTopic DI SINI:
						topic := mapRoleToTopic(models.RoleDirektur)

						title := "Surat Masuk Baru"
						body := fmt.Sprintf("Surat dari %s menunggu disposisi Anda.", event.Letter.Pengirim)
						SendNotificationToTopic(sendCtx, topic, title, body, data)

					} else {
						// Surat Keluar -> Kirim ke MANAJER (Sesuai Scope)
						var targetRole models.Role
						if event.Letter.Scope == models.ScopeInternal {
							targetRole = models.RoleManajerPKL
						} else {
							// Eksternal: Kirim ke KPP & Pemas (Broadcast sementara)
							t1 := mapRoleToTopic(models.RoleManajerKPP)
							t2 := mapRoleToTopic(models.RoleManajerPemas)
							msg := fmt.Sprintf("Surat keluar #%s menunggu verifikasi.", event.Letter.NomorSurat)
							SendNotificationToTopic(sendCtx, t1, "Verifikasi Surat", msg, data)
							SendNotificationToTopic(sendCtx, t2, "Verifikasi Surat", msg, data)
							return
						}

						// Internal -> Kirim ke Manajer PKL
						topic := mapRoleToTopic(targetRole)
						title := "Permintaan Verifikasi"
						body := fmt.Sprintf("Surat keluar #%s menunggu verifikasi Anda.", event.Letter.NomorSurat)
						SendNotificationToTopic(sendCtx, topic, title, body, data)
					}

				// KASUS 2: Status Berubah (Disetujui, Revisi, Disposisi, dll)
				case events.LetterStatusMoved:

					// A. Jika Status jadi 'Perlu Persetujuan' -> Ke DIREKTUR
					if event.Letter.Status == models.StatusPerluPersetujuan {
						topic := mapRoleToTopic(models.RoleDirektur)
						title := "Persetujuan Diperlukan"
						body := fmt.Sprintf("Surat #%s menunggu tanda tangan Anda.", event.Letter.NomorSurat)
						SendNotificationToTopic(sendCtx, topic, title, body, data)
					}

					// B. Jika Status jadi 'Sudah Disposisi' -> Balik ke STAF
					if event.Letter.Status == models.StatusSudahDisposisi {
						// Tentukan staf mana yg dapat notif
						var targetRole models.Role
						if event.Letter.Scope == models.ScopeInternal {
							targetRole = models.RoleStafLembaga
						} else {
							targetRole = models.RoleStafProgram
						}

						topic := mapRoleToTopic(targetRole)
						title := "Disposisi Turun"
						body := fmt.Sprintf("Surat #%s telah didisposisi Direktur.", event.Letter.NomorSurat)
						SendNotificationToTopic(sendCtx, topic, title, body, data)
					}

					// C. Jika Status jadi 'Perlu Revisi' -> Balik ke STAF
					if event.Letter.Status == models.StatusPerluRevisi {
						var targetRole models.Role
						if event.Letter.Scope == models.ScopeInternal {
							targetRole = models.RoleStafLembaga
						} else {
							targetRole = models.RoleStafProgram
						}

						topic := mapRoleToTopic(targetRole)
						title := "Revisi Diperlukan"
						body := fmt.Sprintf("Surat #%s dikembalikan untuk revisi.", event.Letter.NomorSurat)
						SendNotificationToTopic(sendCtx, topic, title, body, data)
					}

					// D. Jika Status jadi 'Disetujui' -> Balik ke STAF (Siap Arsip)
					if event.Letter.Status == models.StatusDisetujui {
						var targetRole models.Role
						if event.Letter.Scope == models.ScopeInternal {
							targetRole = models.RoleStafLembaga
						} else {
							targetRole = models.RoleStafProgram
						}

						topic := mapRoleToTopic(targetRole)
						title := "Surat Disetujui"
						body := fmt.Sprintf("Surat #%s telah disetujui Direktur.", event.Letter.NomorSurat)
						SendNotificationToTopic(sendCtx, topic, title, body, data)
					}
				}
			}(e)
		}
	}
}
