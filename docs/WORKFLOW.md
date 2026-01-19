# Alur Kerja Dokumen (Workflow)

Dokumen ini menjelaskan siklus hidup dan alur kerja untuk berbagai jenis surat dalam Sistem Penyuratan Digital, yang telah disesuaikan dengan prosedur operasional organisasi saat ini.

## 1. Alur Surat Keluar (Outgoing Letter)

Alur kerja ini dibagi berdasarkan **Lingkup (Scope)** surat: **Internal** dan **Eksternal**.

### Peran & Tanggung Jawab

-   **Staf Program:** Membuat surat keluar dengan lingkup **Eksternal**.
-   **Staf Lembaga:** Membuat surat keluar dengan lingkup **Internal**.
-   **Manajer KPP / Pemas:** Memverifikasi surat **Eksternal** (Dipilih secara manual oleh Staf Program saat pembuatan surat).
-   **Manajer PKL:** Memverifikasi surat **Internal** (Ditugaskan secara **OTOMATIS** oleh sistem).
-   **Direktur:** Memberikan persetujuan akhir (Tanda tangan).

### Siklus Status (Lifecycle)

1.  **`DRAFT` / `PERLU_VERIFIKASI`**
    -   **Aksi:** Staf mengunggah surat (PDF/Gambar) dan mengisi data.
    -   **Logika Sistem:**
        -   Jika Lingkup **Eksternal**: Staf Program *wajib* memilih Verifikator (Manajer KPP atau Pemas).
        -   Jika Lingkup **Internal**: Sistem secara **otomatis** menugaskan Manajer PKL.
    -   Status langsung menjadi `PERLU_VERIFIKASI` setelah surat dibuat.

2.  **`PERLU_VERIFIKASI`**
    -   Surat berada dalam antrean Manajer yang ditugaskan.
    -   **Aksi (Manajer):**
        -   **Setujui (Verify):** Status berubah menjadi `PERLU_PERSETUJUAN`.
        -   **Tolak (Reject):** Status berubah menjadi `PERLU_REVISI` (Mengembalikan ke Staf).

3.  **`PERLU_REVISI`**
    -   Surat telah ditolak dan dikembalikan ke Staf pembuat.
    -   **Aksi (Staf):** Staf memperbaiki data atau **mengunggah ulang** file surat yang telah diperbaiki. Setelah disimpan, status kembali menjadi `PERLU_VERIFIKASI`.

4.  **`PERLU_PERSETUJUAN`**
    -   Surat telah diverifikasi oleh Manajer dan kini berada di antrean Direktur.
    -   **Aksi (Direktur):**
        -   **Setujui (Approve):** Surat disetujui. **Sistem secara OTOMATIS mengubah status menjadi `DIARSIPKAN`**.
        -   **Tolak (Reject):** Status berubah menjadi `PERLU_REVISI`.

5.  **`DIARSIPKAN`** (Final)
    -   Alur kerja selesai. Surat dianggap sah dan tersimpan sebagai arsip. Notifikasi dikirimkan kepada Staf pembuat bahwa surat telah final.

---

## 2. Alur Surat Masuk (Incoming Letter)

### Peran & Tanggung Jawab

-   **Staf Lembaga:** Mencatat (registrasi) semua surat yang masuk.
-   **Direktur:** Memeriksa surat dan memberikan disposisi.

### Siklus Status (Lifecycle)

1.  **`BELUM_DISPOSISI`**
    -   Staf Lembaga mencatat surat masuk baru beserta lampiran scan-nya.
    -   **Aksi:** Direktur menerima notifikasi, memeriksa surat, dan memberikan instruksi disposisi.

2.  **`SUDAH_DISPOSISI`**
    -   Direktur telah memberikan instruksi disposisi (Tujuan & Catatan).
    -   **Aksi:** Staf Lembaga menindaklanjuti disposisi tersebut (misalnya: mengarsipkan atau membuat surat balasan).

3.  **`DIARSIPKAN`**
    -   Status akhir untuk pencatatan arsip.