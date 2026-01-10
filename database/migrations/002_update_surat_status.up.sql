-- FILE: 002_update_surat_status.up.sql

START TRANSACTION;

-- [STEP] 1. Ubah kolom status ke VARCHAR sementara
ALTER TABLE surat MODIFY COLUMN status VARCHAR(50) NOT NULL DEFAULT 'draft';

-- [STEP] 2. Migrasi Data Status
-- terkirim -> diarsipkan
UPDATE surat SET status = 'diarsipkan' WHERE status = 'terkirim';

-- [STEP] 3. Terapkan ENUM Status Baru (Final)
ALTER TABLE surat MODIFY COLUMN status ENUM(
    'draft',
    'perlu_verifikasi',
    'belum_disposisi',
    'sudah_disposisi',
    'perlu_persetujuan',
    'perlu_revisi',
    'disetujui',
    'diarsipkan'
    ) NOT NULL DEFAULT 'draft';

COMMIT;