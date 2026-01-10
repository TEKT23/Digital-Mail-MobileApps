-- FILE: 003_set_default_scope.up.sql

START TRANSACTION;

-- [STEP] 1. Isi data scope yang kosong (Logic Mapping)
-- Internal -> Internal
UPDATE surat SET scope = 'Internal' WHERE jenis_surat = 'internal';

-- Masuk/Keluar -> Eksternal (Default)
UPDATE surat SET scope = 'Eksternal' WHERE jenis_surat IN ('masuk', 'keluar') AND scope IS NULL;

-- Fallback safety (jika masih ada yang null)
UPDATE surat SET scope = 'Eksternal' WHERE scope IS NULL;

-- [STEP] 2. Setelah semua data terisi, baru kunci kolomnya jadi NOT NULL
-- Perhatikan: Ini perintah MODIFY, bukan ADD
ALTER TABLE surat MODIFY COLUMN scope ENUM('Internal', 'Eksternal') NOT NULL;

COMMIT;