-- FILE: 001_refactor_roles_and_schema.up.sql

START TRANSACTION;

-- [SAFETY] 1. Handle user dengan role yang tidak dikenali (jika ada data kotor)
-- Set ke 'admin' agar tidak error saat ENUM berubah
UPDATE users SET role = 'admin'
WHERE role NOT IN ('bagian_umum', 'adc', 'direktur', 'admin');

-- [STEP] 2. Ubah kolom role ke VARCHAR sementara (menghindari data truncated)
ALTER TABLE users MODIFY COLUMN role VARCHAR(50) NOT NULL;

-- [STEP] 3. Migrasi Data Role Lama ke Baru
-- bagian_umum -> staf_lembaga
UPDATE users SET role = 'staf_lembaga' WHERE role = 'bagian_umum';
-- adc -> staf_program
UPDATE users SET role = 'staf_program' WHERE role = 'adc';

-- [STEP] 4. Terapkan ENUM Role Baru (Final)
ALTER TABLE users MODIFY COLUMN role ENUM(
    'admin',
    'direktur',
    'staf_program',
    'staf_lembaga',
    'manajer_kpp',
    'manajer_pemas',
    'manajer_pkl'
    ) NOT NULL;

-- [STEP] 5. Tambah kolom baru di tabel surat
-- scope dibuat NULL dulu karena data existing belum punya nilai
ALTER TABLE surat ADD COLUMN scope ENUM('Internal', 'Eksternal') NULL AFTER jenis_surat;
ALTER TABLE surat ADD COLUMN assigned_verifier_id BIGINT UNSIGNED NULL AFTER disposed_by_id;

-- [STEP] 6. Tambah Foreign Key untuk assigned_verifier_id
ALTER TABLE surat ADD CONSTRAINT fk_surat_assigned_verifier
    FOREIGN KEY (assigned_verifier_id) REFERENCES users(id) ON DELETE SET NULL;

-- [STEP] 7. Re-create Indexes (Drop dulu jika ada untuk menghindari error duplicate)
-- Note: Abaikan error "Can't drop check..." jika index belum ada
DROP INDEX IF EXISTS idx_users_role ON users;
CREATE INDEX idx_users_role ON users(role);

DROP INDEX IF EXISTS idx_surat_scope ON surat;
CREATE INDEX idx_surat_scope ON surat(scope);

DROP INDEX IF EXISTS idx_surat_assigned_verifier ON surat;
CREATE INDEX idx_surat_assigned_verifier ON surat(assigned_verifier_id);

COMMIT;