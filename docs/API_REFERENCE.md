# API Reference

This document provides a reference for all the API endpoints available in the Sistem Penyuratan Digital.

## General Information

**Base URL:** `/api`

**Authentication:**

Authentication is required for most endpoints. The API uses JWT (JSON Web Tokens) for authentication. To authenticate, you need to include an `Authorization` header with the value `Bearer <your_access_token>` in your requests.

## Table of Contents

- [Authentication](#authentication)
  - [POST /api/auth/login](#post-apiauthlogin)
  - [POST /api/auth/register](#post-apiauthregister)
  - [POST /api/auth/refresh](#post-apiauthrefresh)
  - [POST /api/auth/logout](#post-apiauthlogout)
  - [POST /api/auth/forgot-password](#post-apiauthforgot-password)
  - [POST /api/auth/reset-password](#post-apiauthreset-password)
- [Letters](#letters)
  - [POST /api/letters](#post-apiletters)
  - [GET /api/letters](#get-apiletters)
  - [GET /api/letters/:id](#get-apilettersid)
  - [PUT /api/letters/:id](#put-apilettersid)
  - [DELETE /api/letters/:id](#delete-apilettersid)
- [Admin](#admin)
  - [POST /api/admin/users](#post-apiadminusers)
  - [GET /api/admin/users](#get-apiadminusers)
  - [GET /api/admin/users/:id](#get-apiadminusersid)
  - [PUT /api/admin/users/:id](#put-apiadminusersid)
  - [DELETE /api/admin/users/:id](#delete-apiadminusersid)

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

| Field    | Type   | Description      |
| :------- | :----- | :--------------- |
| `email`  | string | User's email     |
| `password` | string | User's password  |

**Response (200 OK):**

| Field           | Type   | Description         |
| :-------------- | :----- | :------------------ |
| `access_token`  | string | Access token        |
| `refresh_token` | string | Refresh token       |
| `token_type`    | string | Token type (Bearer) |
| `expires_at`    | string | Expiration time     |
| `user`          | object | User information    |

**Error Response (401 Unauthorized):**

```json
{
  "status": "error",
  "message": "invalid email or password",
  "errors": null
}
```

### POST /api/auth/register

**Request Body:**

```json
{
  "username": "newuser",
  "first_name": "New",
  "last_name": "User",
  "email": "newuser@example.com",
  "password": "newpassword",
  "role": "bagian_umum",
  "jabatan": "Staff",
  "atribut": ""
}
```

| Field       | Type   | Description        |
| :---------- | :----- | :----------------- |
| `username`  | string | Username           |
| `first_name`| string | First name         |
| `last_name` | string | Last name          |
| `email`     | string | Email              |
| `password`  | string | Password           |
| `role`      | string | Role               |
| `jabatan`   | string | Jabatan (Position) |
| `atribut`   | string | Atribut            |

**Error Response (400 Bad Request):**

```json
{
  "status": "error",
  "message": "validation error",
  "errors": {
    "username": "username must be at least 3 characters",
    "email": "invalid email format",
    "password": "password must be at least 8 characters",
    "role": "invalid role provided"
  }
}
```

### POST /api/auth/refresh

Refreshes an access token using a refresh token.

**Request Body:**

```json
{
  "refresh_token": "your_refresh_token"
}
```

| Field           | Type   | Description   |
| :-------------- | :----- | :------------ |
| `refresh_token` | string | Refresh token |

**Response (200 OK):**

| Field           | Type   | Description         |
| :-------------- | :----- | :------------------ |
| `access_token`  | string | New access token    |
| `refresh_token` | string | New refresh token   |
| `token_type`    | string | Token type (Bearer) |
| `expires_at`    | string | Expiration time     |

**Error Response (401 Unauthorized):**

```json
{
  "status": "error",
  "message": "invalid or expired refresh token",
  "errors": null
}
```

### POST /api/auth/logout

Logs out a user by revoking their refresh token.

**Request Body:**

```json
{
  "refresh_token": "your_refresh_token"
}
```

| Field           | Type   | Description   |
| :-------------- | :----- | :------------ |
| `refresh_token` | string | Refresh token |

**Response (200 OK):**

```json
{
  "message": "logout successful"
}
```

### POST /api/auth/forgot-password

Requests a password reset link to be sent to the user's email.

**Request Body:**

```json
{
  "email": "user@example.com"
}
```

| Field   | Type   | Description |
| :------ | :----- | :---------- |
| `email` | string | User's email|

**Response (200 OK):**

```json
{
  "message": "if the email exists, a reset link has been sent"
}
```

### POST /api/auth/reset-password

Resets a user's password using a reset token.

**Request Body:**

```json
{
  "token": "your_reset_token",
  "password": "newpassword",
  "confirm_password": "newpassword"
}
```

| Field                | Type   | Description           |
| :------------------- | :----- | :-------------------- |
| `token`              | string | Password reset token  |
| `password`           | string | New password          |
| `confirm_password`   | string | Confirm new password  |

**Response (200 OK):**

```json
{
  "message": "password has been reset successfully"
}
```

**Error Response (400 Bad Request):**

```json
{
  "status": "error",
  "message": "invalid or expired token",
  "errors": null
}
```

## Letters

Authentication is required for all letter endpoints.

Allowed roles: `bagian_umum`, `adc`, `direktur`, `admin`.

### POST /api/letters

Creates a new letter.

**Request Body:**

```json
{
  "pengirim": "Pengirim",
  "nomor_surat": "SURAT/001",
  "nomor_agenda": "AGENDA/001",
  "disposisi": "Disposisi",
  "tanggal_disposisi": "2025-11-09T00:00:00Z",
  "bidang_tujuan": "Bidang Tujuan",
  "jenis_surat": "masuk",
  "prioritas": "biasa",
  "isi_surat": "Isi Surat",
  "tanggal_surat": "2025-11-09T00:00:00Z",
  "tanggal_masuk": "2025-11-09T00:00:00Z",
  "judul_surat": "Judul Surat",
  "kesimpulan": "Kesimpulan",
  "file_path": "/path/to/file",
  "status": "draft"
}
```

| Field              | Type   | Description          |
| :----------------- | :----- | :------------------- |
| `pengirim`         | string | Sender               |
| `nomor_surat`      | string | Letter number        |
| `nomor_agenda`     | string | Agenda number        |
| `disposisi`        | string | Disposition          |
| `tanggal_disposisi`| string | Disposition date     |
| `bidang_tujuan`    | string | Destination field    |
| `jenis_surat`      | string | Letter type          |
| `prioritas`        | string | Priority             |
| `isi_surat`        | string | Letter content       |
| `tanggal_surat`    | string | Letter date          |
| `tanggal_masuk`    | string | Received date        |
| `judul_surat`      | string | Letter title         |
| `kesimpulan`       | string | Conclusion           |
| `file_path`        | string | File path            |
| `status`           | string | Status               |

**Response (201 Created):**

A letter object.

**Error Response (400 Bad Request):**

```json
{
  "status": "error",
  "message": "validation error",
  "errors": {
    "pengirim": "pengirim is required",
    "nomor_surat": "nomor_surat is required",
    "judul_surat": "judul_surat is required"
  }
}
```

### GET /api/letters

Lists all letters with pagination.

**Query Parameters:**

| Field   | Type   | Description      |
| :------ | :----- | :--------------- |
| `page`  | number | Page number      |
| `limit` | number | Number of items  |

**Response (200 OK):**

A paginated list of letter objects.

### GET /api/letters/:id

Gets a letter by its ID.

**Response (200 OK):**

A letter object.

**Error Response (404 Not Found):**

```json
{
  "status": "error",
  "message": "letter not found",
  "errors": null
}
```

### PUT /api/letters/:id

Updates a letter by its ID.

**Request Body:**

```json
{
  "pengirim": "New Pengirim",
  "status": "perlu_disposisi"
}
```

| Field   | Type   | Description      |
| :------ | :----- | :--------------- |
| `pengirim`| string | Sender           |
| `status`  | string | Status           |

**Response (200 OK):**

A letter object.

**Error Response (404 Not Found):**

```json
{
  "status": "error",
  "message": "letter not found",
  "errors": null
}
```

### DELETE /api/letters/:id

Deletes a letter by its ID.

**Response (200 OK):**

```json
{
  "message": "letter deleted successfully"
}
```

**Error Response (404 Not Found):**

```json
{
  "status": "error",
  "message": "letter not found",
  "errors": null
}
```

## Admin

Authentication and `admin` role are required for all admin endpoints.

### POST /api/admin/users

Creates a new user.

**Request Body:**

```json
{
  "username": "newuser",
  "first_name": "New",
  "last_name": "User",
  "email": "newuser@example.com",
  "password": "newpassword",
  "role": "bagian_umum",
  "jabatan": "Staff",
  "atribut": ""
}
```

| Field       | Type   | Description        |
| :---------- | :----- | :----------------- |
| `username`  | string | Username           |
| `first_name`| string | First name         |
| `last_name` | string | Last name          |
| `email`     | string | Email              |
| `password`  | string | Password           |
| `role`      | string | Role               |
| `jabatan`   | string | Jabatan (Position) |
| `atribut`   | string | Atribut            |

**Response (201 Created):**

A user object.

**Error Response (400 Bad Request):**

```json
{
  "status": "error",
  "message": "validation error",
  "errors": {
    "username": "username is required",
    "email": "email is required",
    "password": "password must be at least 8 characters"
  }
}
```

### GET /api/admin/users

Lists all users with pagination and filtering.

**Query Parameters:**

| Field | Type   | Description        |
| :---- | :----- | :----------------- |
| `page`| number | Page number        |
| `limit`| number | Number of items    |
| `role`| string | Role to filter by  |
| `q`   | string | Search query       |

**Response (200 OK):**

A paginated list of user objects.

### GET /api/admin/users/:id

Gets a user by their ID.

**Response (200 OK):**

A user object.

**Error Response (404 Not Found):**

```json
{
  "status": "error",
  "message": "user not found",
  "errors": null
}
```

### PUT /api/admin/users/:id

Updates a user by their ID.

**Request Body:**

```json
{
  "username": "newusername",
  "role": "adc"
}
```

| Field      | Type   | Description |
| :--------- | :----- | :---------- |
| `username` | string | Username    |
| `role`     | string | Role        |

**Response (200 OK):**

A user object.

**Error Response (404 Not Found):**

```json
{
  "status": "error",
  "message": "user not found",
  "errors": null
}
```

### DELETE /api/admin/users/:id

Deletes a user by their ID.

**Response (200 OK):**

```json
{
  "message": "user deleted successfully"
}
```

**Error Response (404 Not Found):**

```json
{
  "status": "error",
  "message": "user not found",
  "errors": null
}
```

**Note:** This documentation was generated by a large language model.
