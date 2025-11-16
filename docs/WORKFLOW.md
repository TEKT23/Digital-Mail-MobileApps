# Dokumentasi Alur Kerja (Workflows)

Dokumen ini menjelaskan alur kerja (state machine) untuk siklus hidup surat, berdasarkan logika di `handlers/letter_handlers.go`.

## 1. Daftar Status Surat

Semua status ini didefinisikan dalam `models/letters.entity.go`.

### Status Alur Surat Masuk
* `draft`: Surat dibuat oleh Bagian Umum, belum diajukan.
* `perlu_verifikasi`: Diajukan oleh Bagian Umum, menunggu verifikasi ADC.
* `belum_disposisi`: Diverifikasi oleh ADC, menunggu disposisi Direktur.
* `sudah_disposisi`: Telah didisposisi oleh Direktur. Alur selesai.

### Status Alur Surat Keluar
* `draft`: Surat dibuat oleh ADC, belum diajukan.
* `perlu_persetujuan`: Diajukan oleh ADC, menunggu persetujuan Direktur.
* `perlu_revisi`: Ditolak oleh Direktur, dikembalikan ke ADC untuk revisi.
* `disetujui`: Disetujui oleh Direktur, dikembalikan ke ADC untuk finalisasi.
* `terkirim`: Telah dikirim/diarsipkan oleh ADC. Alur selesai.

---

## 2. Alur Surat Masuk (Disposisi)

Alur ini berlaku untuk `JenisSurat: "masuk"` atau `"internal"`.

| Langkah | Aktor | Status Awal | Tindakan | Status Akhir | Notifikasi Terkirim |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1. Membuat | `Bagian Umum` | (N/A) | `POST /api/letters` | `draft` | (Tidak ada) |
| 2. Mengajukan | `Bagian Umum` | `draft` | `PUT /api/letters/:id` | `perlu_verifikasi` | **ADC** & **Direktur** |
| 3. Verifikasi | `ADC` | `perlu_verifikasi` | `PUT /api/letters/:id` | `belum_disposisi` | **Direktur** |
| 4. Disposisi | `Direktur` | `belum_disposisi` | `PUT /api/letters/:id` | `sudah_disposisi` | **ADC** & **Bagian Umum** |

---

## 3. Alur Surat Keluar (Persetujuan)

Alur ini berlaku untuk `JenisSurat: "keluar"`. Pembuatan surat diizinkan untuk `RoleADC`.

| Langkah | Aktor | Status Awal | Tindakan | Status Akhir | Notifikasi Terkirim |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1. Membuat | `ADC` | (N/A) | `POST /api/letters` | `draft` | (Tidak ada) |
| 2. Mengajukan | `ADC` | `draft` | `PUT /api/letters/:id` | `perlu_persetujuan` | **Direktur** |
| 3. (Opsi A) Setuju | `Direktur` | `perlu_persetujuan` | `PUT /api/letters/:id` | `disetujui` | **ADC** |
| 3. (Opsi B) Tolak | `Direktur` | `perlu_persetujuan` | `PUT /api/letters/:id` | `perlu_revisi` | **ADC** |
| 4. Revisi | `ADC` | `perlu_revisi` | `PUT /api/letters/:id` | `perlu_persetujuan` | **Direktur** (Kembali ke Langkah 2) |
| 5. Finalisasi | `ADC` | `disetujui` | `PUT /api/letters/:id` | `terkirim` | (Tidak ada) |