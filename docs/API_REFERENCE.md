# API Reference

This document provides a reference for all the API endpoints available in the Sistem Penyuratan Digital.

## General Information

**Base URL:** `/api`

**Authentication:**

Authentication is required for most endpoints. The API uses JWT (JSON Web Tokens) for authentication. To authenticate, you need to include an `Authorization` header with the value `Bearer <your_access_token>` in your requests.

**Standard Response Format:**

All API responses follow a standard JSON format:

```json
{
  "status": "success" | "error",
  "message": "<A descriptive message>",
  "data": "<The requested data (on success)>",
  "errors": "<Validation errors or other error details (on error)>"
}
```

## Table ofContents

- [Authentication](#authentication)
  - [POST /api/auth/login](#post-apiauthlogin)
  - [POST /api/auth/register](#post-apiauthregister)
  - [POST /api/auth/refresh](#post-apiauthrefresh)
  - [POST /api/auth/logout](#post-apiauthlogout)
  - [POST /api/auth/forgot-password](#post-apiauthforgot-password)
  - [POST /api/auth/reset-password](#post-apiauthreset-password)
- [File Upload](#file-upload)
  - [POST /api/upload](#post-apiupload)
- [Surat Keluar (Outgoing Letters)](#surat-keluar-outgoing-letters)
  - [Workflow](#workflow)
  - [POST /api/letters/keluar](#post-apiletterskeluar)
  - [PUT /api/letters/keluar/:id](#put-apiletterskeluarid)
  - [POST /api/letters/keluar/:id/verify/approve](#post-apiletterskeluaridverifyapprove)
  - [POST /api/letters/keluar/:id/approve](#post-apiletterskeluaridapprove)
  - [GET /api/letters/keluar/my](#get-apiletterskeluarmy)
  - [GET /api/letters/keluar/need-verification](#get-apiletterskeluarneed-verification)
  - [GET /api/letters/keluar/need-approval](#get-apiletterskeluarneed-approval)
- [Surat Masuk (Incoming Letters)](#surat-masuk-incoming-letters)
  - [Workflow](#workflow-1)
  - [POST /api/letters/masuk](#post-apilettersmasuk)
  - [POST /api/letters/masuk/:id/dispose](#post-apilettersmasukiddispose)
  - [GET /api/letters/masuk/need-disposition](#get-apilettersmasukneed-disposition)
- [Common Letter Endpoints](#common-letter-endpoints)
  - [GET /api/letters/:id](#get-apilettersid)
  - [DELETE /api/letters/:id](#delete-apilettersid)
- [Admin](#admin)
- [User Settings](#user-settings)

---

## Authentication

### POST /api/auth/login

Logs in a user and returns an access token and a refresh token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

**Response (200 OK):**
```json
{
    "status": "success",
    "message": "Login successful",
    "data": {
        "access_token": "your_access_token",
        "refresh_token": "your_refresh_token",
        "token_type": "Bearer",
        "expires_at": "2025-01-12T12:00:00Z",
        "user": {
            "id": 1,
            "username": "testuser",
            "email": "user@example.com",
            "role": "staf"
        }
    }
}
```

### POST /api/auth/register

Registers a new user.

**Request Body:**
```json
{
  "username": "newuser",
  "first_name": "New",
  "last_name": "User",
  "email": "newuser@example.com",
  "password": "newpassword",
  "role": "staf",
  "jabatan": "Staff"
}
```

**Response (201 Created):**
```json
{
    "status": "success",
    "message": "User registered successfully",
    "data": {
        "id": 2,
        "username": "newuser",
        "email": "newuser@example.com",
        "role": "staf"
    }
}
```

---

## File Upload

### POST /api/upload

Uploads a file and returns its path.

**Request Body:**
- `multipart/form-data` with a single file field named `file`.

**Response (200 OK):**
```json
{
    "status": "success",
    "message": "File uploaded successfully",
    "data": {
        "file_path": "uploads/2025/01/unique-file-name.pdf"
    }
}
```

---

## Surat Keluar (Outgoing Letters)

This workflow manages the creation, verification, and approval of outgoing letters.

### Workflow

1.  **Staf:** Creates a `DRAFT` letter.
2.  **Staf:** Submits the draft for verification, changing its status to `PERLU_VERIFIKASI`.
3.  **Manajer:** Approves the letter (status becomes `PERLU_PERSETUJUAN`) or rejects it (status becomes `PERLU_REVISI`).
4.  **Direktur:** Approves the letter (status becomes `DISETUJUI`) or rejects it (status becomes `PERLU_REVISI`).
5.  **Staf:** Archives the approved letter (status becomes `DIARSIPKAN`).

### POST /api/letters/keluar

Creates a new outgoing letter (as a draft).
**Required Role:** `staf`

**Request Body:**
```json
{
  "nomor_surat": "SURAT/KELUAR/001",
  "judul_surat": "Judul Surat Keluar",
  "tujuan": "Tujuan Surat",
  "isi_surat": "Isi dari surat keluar...",
  "scope": "Internal",
  "assigned_verifier_id": 2,
  "file_path": "uploads/2025/01/file.pdf"
}
```

**Response (201 Created):**
```json
{
    "status": "success",
    "message": "Surat keluar created successfully",
    "data": {
        "id": 1,
        "status": "draft",
        "judul_surat": "Judul Surat Keluar"
    }
}
```

### PUT /api/letters/keluar/:id

Updates a draft or a letter needing revision.
**Required Role:** `staf`

**Request Body:**
```json
{
  "judul_surat": "Judul Surat Keluar (Revisi)"
}
```

### POST /api/letters/keluar/:id/verify/approve

Approves a letter for verification (Manajer).
**Required Role:** `manajer`

**Response (200 OK):**
```json
{
    "status": "success",
    "message": "Letter approved for verification",
    "data": {
        "id": 1,
        "status": "perlu_persetujuan"
    }
}
```

### POST /api/letters/keluar/:id/approve

Approves a letter (Direktur).
**Required Role:** `direktur`

**Response (200 OK):**
```json
{
    "status": "success",
    "message": "Letter approved successfully",
    "data": {
        "id": 1,
        "status": "disetujui"
    }
}
```

### GET /api/letters/keluar/my

Gets the dashboard for the current Staf user, showing their letters.
**Required Role:** `staf`

### GET /api/letters/keluar/need-verification

Gets the dashboard for Manajer, showing letters that need verification.
**Required Role:** `manajer`

### GET /api/letters/keluar/need-approval

Gets the dashboard for Direktur, showing letters that need approval.
**Required Role:** `direktur`

---

## Surat Masuk (Incoming Letters)

This workflow manages the registration and disposition of incoming letters.

### Workflow

1.  **Staf:** Registers an incoming letter. The status is set to `BELUM_DISPOSISI`.
2.  **Direktur:** Adds disposition instructions to the letter. The status changes to `SUDAH_DISPOSISI`.
3.  **Staf:** Archives the letter (status becomes `DIARSIPKAN`).

### POST /api/letters/masuk

Registers a new incoming letter.
**Required Role:** `staf`

**Request Body:**
```json
{
  "nomor_surat": "SURAT/MASUK/001",
  "pengirim": "Pengirim Eksternal",
  "judul_surat": "Judul Surat Masuk",
  "tanggal_surat": "2025-01-10",
  "tanggal_masuk": "2025-01-12",
  "scope": "Eksternal",
  "file_scan_path": "uploads/2025/01/scan.pdf",
  "prioritas": "biasa",
  "isi_surat": "Isi dari surat masuk..."
}
```

**Response (201 Created):**
```json
{
    "status": "success",
    "message": "Surat masuk created successfully",
    "data": {
        "id": 2,
        "status": "belum_disposisi",
        "judul_surat": "Judul Surat Masuk"
    }
}
```

### POST /api/letters/masuk/:id/dispose

Adds a disposition to an incoming letter.
**Required Role:** `direktur`

**Request Body:**
```json
{
  "disposition": "Segera proses dan laporkan hasilnya.",
  "disposition_user_ids": [3, 4]
}
```

**Response (200 OK):**
```json
{
    "status": "success",
    "message": "Letter disposition added successfully",
    "data": {
        "id": 2,
        "status": "sudah_disposisi"
    }
}
```

### GET /api/letters/masuk/need-disposition

Gets the dashboard for Direktur, showing letters that need disposition.
**Required Role:** `direktur`

---

## Common Letter Endpoints

### GET /api/letters/:id

Gets the details of any letter by its ID. Access is determined by user role and letter state.

### DELETE /api/letters/:id

Soft deletes a letter. Access is restricted.

---

## Admin

Endpoints under `/api/admin/users/*` are available for managing users. These are restricted to users with the `admin` role.

## User Settings

Endpoints under `/api/settings/*` are available for users to manage their own settings, such as changing their password.