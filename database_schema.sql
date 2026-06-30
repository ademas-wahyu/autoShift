-- ============================================================
-- DATABASE SCHEMA: autoShift
-- Target: PostgreSQL (adaptable ke MySQL/SQLite)
-- ============================================================

-- 1. MULTI-TENANT -------------------------------------------------
CREATE TABLE tenants (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW()
);

-- 2. ADMIN USERS --------------------------------------------------
CREATE TABLE users (
    id              SERIAL PRIMARY KEY,
    tenant_id       INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP DEFAULT NOW()
);

-- 3. KARYAWAN & ROLE ----------------------------------------------

CREATE TABLE employee_roles (
    id          SERIAL PRIMARY KEY,
    tenant_id   INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,   -- Supervisor, Staff, Intern
    level       INTEGER NOT NULL DEFAULT 0,  -- higher = more senior
    UNIQUE(tenant_id, name)
);

CREATE TABLE employees (
    id                SERIAL PRIMARY KEY,
    tenant_id         INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role_id           INTEGER NOT NULL REFERENCES employee_roles(id),
    name              VARCHAR(255) NOT NULL,
    email             VARCHAR(255),
    phone             VARCHAR(50),
    is_active         BOOLEAN DEFAULT TRUE,
    created_at        TIMESTAMP DEFAULT NOW()
);

-- 4. SHIFT TEMPLATES & ROLE REQUIREMENTS --------------------------

CREATE TABLE shift_templates (
    id            SERIAL PRIMARY KEY,
    tenant_id     INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name          VARCHAR(100) NOT NULL,     -- Pagi, Siang, Malam
    start_time    TIME NOT NULL,
    end_time      TIME NOT NULL,
    is_cross_day  BOOLEAN NOT NULL DEFAULT FALSE,  -- TRUE jika shift lintas hari (22:00-06:00)
    color_hex     VARCHAR(7) DEFAULT '#3B82F6',
    UNIQUE(tenant_id, name)
);

-- Komposisi role wajib per shift template
CREATE TABLE shift_role_requirements (
    id                  SERIAL PRIMARY KEY,
    shift_template_id   INTEGER NOT NULL REFERENCES shift_templates(id) ON DELETE CASCADE,
    role_id             INTEGER NOT NULL REFERENCES employee_roles(id) ON DELETE CASCADE,
    min_count           INTEGER NOT NULL DEFAULT 1,
    UNIQUE(shift_template_id, role_id)
);

-- 5. KUOTA LIBUR PER KARYAWAN ------------------------------------

CREATE TABLE employee_leave_quotas (
    id            SERIAL PRIMARY KEY,
    tenant_id     INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    employee_id   INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    month         INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year          INTEGER NOT NULL,
    quota_days    INTEGER NOT NULL DEFAULT 0,
    UNIQUE(employee_id, month, year)
);

-- 6. HARI LIBUR NASIONAL (cache dari public API) -----------------

CREATE TABLE holidays (
    id            SERIAL PRIMARY KEY,
    date          DATE NOT NULL UNIQUE,
    name          VARCHAR(255) NOT NULL,
    is_national   BOOLEAN DEFAULT TRUE
);

-- 7. SCHEDULE (1 baris per bulan yang di-generate) ----------------

CREATE TYPE leave_mode AS ENUM ('fixed', 'random');
CREATE TYPE schedule_status AS ENUM ('draft', 'published');

CREATE TABLE schedules (
    id              SERIAL PRIMARY KEY,
    tenant_id       INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    month           INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year            INTEGER NOT NULL,
    leave_mode      leave_mode NOT NULL,
    status          schedule_status NOT NULL DEFAULT 'draft',
    notes           TEXT,
    created_by      INTEGER NOT NULL REFERENCES users(id),
    created_at      TIMESTAMP DEFAULT NOW(),
    published_at    TIMESTAMP,
    UNIQUE(tenant_id, month, year)
);

-- 8. KONFIGURASI LIBUR PER SCHEDULE -------------------------------

-- Jika mode fixed: hari apa saja yang libur tiap minggu
CREATE TABLE schedule_fixed_leaves (
    id              SERIAL PRIMARY KEY,
    schedule_id     INTEGER NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    day_of_week     INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Minggu
    UNIQUE(schedule_id, day_of_week)
);

-- Jika mode random: berapa hari libur acak per minggu
CREATE TABLE schedule_random_leaves (
    id              SERIAL PRIMARY KEY,
    schedule_id     INTEGER NOT NULL REFERENCES schedules(id) ON DELETE CASCADE UNIQUE,
    days_per_week   INTEGER NOT NULL CHECK (days_per_week BETWEEN 1 AND 6)
);

-- 9. SHIFT ASSIGNMENTS --------------------------------------------

-- Jadwal kerja (shift) per hari per karyawan
CREATE TABLE schedule_shifts (
    id                  SERIAL PRIMARY KEY,
    schedule_id         INTEGER NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    employee_id         INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    shift_template_id   INTEGER NOT NULL REFERENCES shift_templates(id),
    date                DATE NOT NULL,
    is_override         BOOLEAN DEFAULT FALSE,
    UNIQUE(schedule_id, employee_id, date)
);

-- 10. JADWAL LIBUR (EKSPLISIT) ------------------------------------

-- Hari libur yang sudah di-plot oleh AI (agar frontend bisa bedakan
-- "libur terjadwal" vs "belum di-generate")
CREATE TABLE schedule_employee_leaves (
    id              SERIAL PRIMARY KEY,
    schedule_id     INTEGER NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    employee_id     INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    is_override     BOOLEAN DEFAULT FALSE,
    UNIQUE(schedule_id, employee_id, date)
);

-- 11. LOG GENERATION (untuk audit batch AI) -----------------------

CREATE TABLE generation_logs (
    id                SERIAL PRIMARY KEY,
    schedule_id       INTEGER NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    batch_index       INTEGER,
    status            VARCHAR(50),      -- success, retry, failed
    prompt_tokens     INTEGER,
    response_time_ms  INTEGER,
    error_message     TEXT,
    created_at        TIMESTAMP DEFAULT NOW()
);

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX idx_employees_tenant ON employees(tenant_id);
CREATE INDEX idx_schedules_tenant ON schedules(tenant_id, month, year);
CREATE INDEX idx_schedule_shifts_schedule ON schedule_shifts(schedule_id);
CREATE INDEX idx_schedule_shifts_employee ON schedule_shifts(employee_id);
CREATE INDEX idx_schedule_shifts_date ON schedule_shifts(date);
CREATE INDEX idx_schedule_leaves_schedule ON schedule_employee_leaves(schedule_id);
CREATE INDEX idx_schedule_leaves_employee ON schedule_employee_leaves(employee_id);
CREATE INDEX idx_holidays_date ON holidays(date);
CREATE INDEX idx_leave_quotas_employee ON employee_leave_quotas(employee_id, month, year);

-- ============================================================
-- CONSTRAINT VALIDATION QUERIES
-- ============================================================

-- 1. CEK JEDA ISTIRAHAT < 12 JAM (dengan is_cross_day fix)
-- SELECT a.employee_id,
--        a.date AS tgl_shift1,
--        b.date AS tgl_shift2
-- FROM schedule_shifts a
-- JOIN schedule_shifts b ON a.employee_id = b.employee_id
--     AND b.date = a.date + INTERVAL '1 day'
-- JOIN shift_templates sa ON a.shift_template_id = sa.id
-- JOIN shift_templates sb ON b.shift_template_id = sb.id
-- WHERE a.schedule_id = $1
--   AND (
--       (b.date + sb.start_time) -
--       (a.date + sa.end_time + CASE WHEN sa.is_cross_day THEN INTERVAL '1 day' ELSE INTERVAL '0' END)
--   ) < INTERVAL '12 hours';

-- 2. CEK KOMPOSISI ROLE PER SHIFT PER HARI
-- SELECT ss.date, ss.shift_template_id, er.name, COUNT(*) AS total
-- FROM schedule_shifts ss
-- JOIN employees e ON ss.employee_id = e.id
-- JOIN employee_roles er ON e.role_id = er.id
-- WHERE ss.schedule_id = $1
-- GROUP BY ss.date, ss.shift_template_id, er.name
-- HAVING COUNT(*) < (
--     SELECT min_count FROM shift_role_requirements
--     WHERE shift_template_id = ss.shift_template_id AND role_id = er.id
-- );

-- 3. CEK DOBEL DATA (kerja + libur di hari yang sama)
-- SELECT ss.employee_id, ss.date
-- FROM schedule_shifts ss
-- JOIN schedule_employee_leaves sl
--     ON ss.schedule_id = sl.schedule_id
--     AND ss.employee_id = sl.employee_id
--     AND ss.date = sl.date
-- WHERE ss.schedule_id = $1;

-- 4. CEK KUOTA LIBUR TERLAMPAUI
-- SELECT sl.employee_id, COUNT(*) AS total_leave_days, q.quota_days
-- FROM schedule_employee_leaves sl
-- JOIN employee_leave_quotas q
--     ON sl.employee_id = q.employee_id
--     AND EXTRACT(MONTH FROM sl.date) = q.month
--     AND EXTRACT(YEAR FROM sl.date) = q.year
-- WHERE sl.schedule_id = $1
-- GROUP BY sl.employee_id, q.quota_days
-- HAVING COUNT(*) > q.quota_days;
