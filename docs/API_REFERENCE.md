# Referensi API Sistem Penyuratan Digital

Dokumentasi ini menjelaskan spesifikasi endpoint API untuk aplikasi backend Sistem Penyuratan Digital (Golang Fiber).

## Informasi Umum

- **Base URL**: `http://localhost:8080/api` (Sesuaikan dengan deployment)
- **Authentication**: Menggunakan Header `Authorization: Bearer <token>`
- **Format Date**: RFC3339 (`YYYY-MM-DDTHH:mm:ssZ`) atau `YYYY-MM-DD`

## Standar Response

API ini menggunakan format response JSON yang konsisten untuk semua endpoint.

### 1. Response Sukses (200/201)

```json
{
  "success": true,
  "message": "Operasi berhasil",
  "data": {
    "id": 1,
    "nomor_surat": "001/PKL/2026"
  }
}
```

### 2. Response Error (400/401/403/404/500)

```json
{
  "success": false,
  "code": 400,
  "message": "Validasi gagal",
  "errors": {
    "judul_surat": "judul_surat is required",
    "scope": "scope is required"
  }
}
```

---

## 1. Authentication

### Login User
Masuk ke sistem untuk mendapatkan Access Token.

- **Endpoint**: `POST /auth/login`
- **Content-Type**: `application/json`

**Request Body:**
```json
{
  "email": "staf@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUz...",
    "expires_at": "2026-01-21T10:00:00Z",
    "user": {
      "id": 1,
      "email": "staf@example.com",
      "role": "staf"
    }
  }
}
```

### Register User
Pendaftaran pengguna baru (jika diizinkan publik, atau oleh admin).

- **Endpoint**: `POST /auth/register`
- **Content-Type**: `application/json`

**Request Body:**
```json
{
  "username": "budi_staf",
  "email": "budi@example.com",
  "password": "password123",
  "role": "staf",
  "first_name": "Budi",
  "last_name": "Santoso"
}
```

---

## 2. Utility & Helpers

### Get Available Verifiers
Mendapatkan daftar user yang berhak memverifikasi surat (Manajer PKL atau Verifikator Eksternal) berdasarkan scope.

- **Endpoint**: `GET /letters/verifiers`
- **Query Params**:
  - `scope`: `Internal` atau `Eksternal`

**Request Example:**
`GET /api/letters/verifiers?scope=Eksternal`

**Response:**
```json
{
  "success": true,
  "message": "List verifiers retrieved",
  "data": [
    {
      "id": 5,
      "username": "dosen_pembimbing",
      "jabatan": "Dosen Pembimbing Lapangan"
    }
  ]
}
```

---

## 3. Manajemen Surat Keluar (Outgoing)

### Create Surat Keluar
Membuat draft surat keluar baru.

- **Endpoint**: `POST /letters/keluar`
- **Content-Type**: `multipart/form-data` **(WAJIB)**
- **Akses**: Staf

**Logika Bisnis Scope:**
1. **Internal**: Jika `scope` = "Internal", backend otomatis menugaskan **Manajer PKL** sebagai verifikator. Field `assigned_verifier_id` diabaikan/kosongkan.
2. **Eksternal**: Jika `scope` = "Eksternal", user **WAJIB** mengirim `assigned_verifier_id` (pilih dari list verifiers).

**Form-Data Fields:**

| Key | Type | Required | Deskripsi |
| :--- | :--- | :--- | :--- |
| `judul_surat` | Text | Yes | Judul atau perihal surat |
| `nomor_surat` | Text | Yes | Nomor referensi surat |
| `isi_surat` | Text | Yes | Ringkasan isi surat |
| `pengirim` | Text | Yes | Nama pengirim |
| `jenis_surat` | Text | Yes | Value: `keluar` |
| `prioritas` | Text | Yes | Value: `biasa`, `segera`, atau `penting` |
| `scope` | Text | Yes | Value: `Internal` atau `Eksternal` |
| `assigned_verifier_id`| Number | Conditional | ID User Verifikator (Wajib jika Eksternal) |
| `tanggal_surat` | Date | No | Format: YYYY-MM-DD |
| `file` | File | Yes | File dokumen surat (PDF/Image) |

**Response:**
```json
{
  "success": true,
  "message": "Surat keluar created successfully",
  "data": {
    "id": 10,
    "status": "draft",
    "file_path": "/uploads/surat_keluar/doc_10.pdf"
  }
}
```

### Update / Revisi Surat Keluar
Mengedit draft surat atau melakukan revisi jika status `perlu_revisi`.

- **Endpoint**: `PUT /letters/keluar/:id`
- **Content-Type**: `multipart/form-data`
- **Akses**: Staf

**Catatan Revisi File:**
- User dapat mengupload file baru di field `file` untuk mengganti dokumen lama.
- Jika field `file` dikosongkan, dokumen lama akan tetap digunakan.

**Form-Data Fields:**
(Sama seperti Create, semua field opsional kecuali yang ingin diubah).

### Verify Surat (Approve/Reject)
Proses verifikasi oleh Manajer (Internal) atau Verifikator (Eksternal).

- **Endpoint Approve**: `POST /letters/keluar/:id/verify/approve`
- **Endpoint Reject**: `POST /letters/keluar/:id/verify/reject`
- **Akses**: Manajer / Verifikator

**Request Body (JSON):**
```json
{
  "catatan": "Dokumen sudah sesuai, lanjut ke direktur."
}
```

**Logika:**
- **Approve**: Status berubah menjadi `perlu_persetujuan` (menunggu Direktur).
- **Reject**: Status berubah menjadi `perlu_revisi` (kembali ke Staf).

### Approve Surat (Finalisasi Direktur)
Persetujuan akhir oleh Direktur.

- **Endpoint**: `POST /letters/keluar/:id/approve`
- **Akses**: Direktur

**Request Body (JSON):**
```json
{
  "catatan": "Disetujui, silakan kirim."
}
```

**Logika Baru:**
- Aksi ini akan langsung mengubah status surat menjadi **DIARSIPKAN** (Final).
- Melewati status 'Disetujui' karena dianggap proses surat keluar selesai saat ditandatangani/disetujui Direktur.

---

## 4. Manajemen Surat Masuk (Incoming)

### Create Surat Masuk
Mencatat surat yang diterima dari pihak luar.

- **Endpoint**: `POST /letters/masuk`
- **Content-Type**: `multipart/form-data`
- **Akses**: Staf

**Form-Data Fields:**

| Key | Type | Required | Deskripsi |
| :--- | :--- | :--- | :--- |
| `judul_surat` | Text | Yes | Perihal surat masuk |
| `nomor_surat` | Text | Yes | Nomor surat dari pengirim |
| `pengirim` | Text | Yes | Instansi/Orang pengirim |
| `jenis_surat` | Text | Yes | Value: `masuk` |
| `prioritas` | Text | Yes | Value: `biasa`, `segera`, atau `penting` |
| `tanggal_masuk` | Date | Yes | Tanggal diterima (YYYY-MM-DD) |
| `file` | File | Yes | Scan file surat masuk (PDF/Image) |

**Response:**
```json
{
  "success": true,
  "message": "Surat masuk recorded successfully",
  "data": {
    "id": 25,
    "status": "belum_disposisi"
  }
}
```

### Disposisi Surat Masuk
Direktur memberikan instruksi disposisi ke bawahan.

- **Endpoint**: `POST /letters/masuk/:id/dispose`
- **Content-Type**: `application/json`
- **Akses**: Direktur

**Request Body:**
```json
{
  "disposisi": "Tindak lanjuti segera",
  "bidang_tujuan": "Divisi TI",
  "catatan": "Koordinasikan dengan Pak Budi"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Surat disposed successfully",
  "data": {
    "id": 25,
    "status": "sudah_disposisi"
  }
}
```
