# Digital Mail API
Backend untuk aplikasi manajemen surat Digital Mail yang dibangun dengan Fiber (Go).

## Dokumentasi

- [API Reference](docs/API_REFERENCE.md)

## Konfigurasi CORS untuk Klien Mobile

Server API menggunakan middleware CORS bawaan Fiber dengan konfigurasi berikut:

- **Origins yang diizinkan**: `capacitor://localhost`, `ionic://localhost`, `http://localhost`, `https://localhost`
- **Metode HTTP yang diizinkan**: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`

Pastikan aplikasi mobile Anda menggunakan salah satu origin di atas dan melakukan permintaan menggunakan metode yang diizinkan agar dapat terhubung ke API tanpa kendala CORS.