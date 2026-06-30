# autoShift Backend API

Backend REST API untuk aplikasi penjadwalan shift karyawan berbasis AI. Dibangun dengan **Go**, **Fiber v2**, **GORM**, dan **PostgreSQL**.

## Persyaratan

- Go 1.21+
- PostgreSQL 14+
- (Opsional) Kunci API OpenAI untuk generator AI nyata

## Instalasi

```bash
# Clone repo
git clone <repo-url>
cd backend

# Salin konfigurasi
cp .env.example .env
# Edit .env sesuai lingkungan Anda

# Jalankan
go run main.go
```

Server akan:
1. Membuat koneksi ke database
2. Migrasi tabel secara otomatis
3. Mengisi data awal (tenant default, 3 role, 3 shift template, 5 karyawan)

## Konfigurasi (.env)

| Variabel | Default | Deskripsi |
|---|---|---|
| `DB_HOST` | `localhost` | Host PostgreSQL |
| `DB_PORT` | `5432` | Port PostgreSQL |
| `DB_USER` | `postgres` | User PostgreSQL |
| `DB_PASSWORD` | `postgres` | Password PostgreSQL |
| `DB_NAME` | `autoshift` | Nama database |
| `JWT_SECRET` | `autoshift-secret-change-me` | Secret key JWT |
| `SERVER_PORT` | `8080` | Port server |
| `AI_PROVIDER` | `mock` | `mock` atau `openai` |
| `OPENAI_API_KEY` | - | Kunci API OpenAI |
| `OPENAI_MODEL` | `gpt-4o` | Model OpenAI |

## Data Awal (Seed)

Saat pertama kali dijalankan, server mengisi:

| Data | Detail |
|---|---|
| **Tenant** | `Default Company` |
| **Role** | Supervisor (level 2), Staff (level 1), Intern (level 0) |
| **Shift** | Pagi (08:00-16:00), Siang (16:00-00:00, cross-day), Malam (22:00-06:00, cross-day) |
| **Role Req** | Setiap shift minimal 1 Supervisor |
| **Karyawan** | Budi (Supervisor), Siti (Staff), Andi (Staff), Dewi (Staff), Rudi (Supervisor) |

## Alur Autentikasi

1. **Login** → dapatkan JWT token
2. Sertakan token di header: `Authorization: Bearer <token>`
3. Token mengandung klaim `sub` (user ID) dan `tenant_id`

## API Endpoints

### Publik (tanpa autentikasi)

---

#### `GET /api/v1/health`

Cek status server.

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

---

#### `POST /api/v1/login`

Login pengguna (stub — akan diimplementasikan).

**Response:**
```json
{
  "message": "login endpoint"
}
```

---

#### `GET /api/v1/holidays?year=2026`

Daftar hari libur nasional.

**Parameter:**
| Nama | Tipe | Default | Deskripsi |
|---|---|---|---|
| `year` | query | `2026` | Tahun |

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "date": "2026-01-01",
      "name": "Tahun Baru Masehi",
      "is_national": true
    }
  ]
}
```

---

#### `POST /api/v1/holidays/fetch?year=2026&country=ID`

Ambil data hari libur dari [date.nager.at](https://date.nager.at).

**Parameter:**
| Nama | Tipe | Default | Deskripsi |
|---|---|---|---|
| `year` | query | `2026` | Tahun |
| `country` | query | `ID` | Kode negara (ISO) |

**Response:**
```json
{
  "success": true,
  "message": "Holidays fetched"
}
```

---

### Terproteksi (memerlukan JWT)

Semua endpoint di bawah memerlukan header:
```
Authorization: Bearer <jwt_token>
```

---

#### `POST /api/v1/schedules/`

Buat jadwal baru. Generate shift berjalan secara **async** di background.

**Request Body:**
```json
{
  "month": 8,
  "year": 2026,
  "leave_mode": "fixed",
  "fixed_leaves": [0, 6],
  "employee_ids": [1, 2, 3, 4, 5]
}
```

| Field | Tipe | Wajib | Deskripsi |
|---|---|---|---|
| `month` | int | ya | Bulan (1-12) |
| `year` | int | ya | Tahun |
| `leave_mode` | string | ya | `"fixed"` atau `"random"` |
| `fixed_leaves` | []int | jika fixed | Hari libur tetap (0=Minggu, 6=Sabtu) |
| `random_days_per_week` | int | jika random | Jumlah hari libur acak per minggu |
| `employee_ids` | []uint | ya | ID karyawan (jika kosong, pakai semua) |

**Response (201):**
```json
{
  "success": true,
  "data": {
    "schedule_id": 1,
    "status": "generating",
    "message": "Jadwal sedang di-generate."
  }
}
```

---

#### `GET /api/v1/schedules/:id`

Ambil detail jadwal lengkap.

**Response:**
```json
{
  "success": true,
  "data": {
    "schedule": {
      "id": 1,
      "tenant_id": 1,
      "month": 8,
      "year": 2026,
      "leave_mode": "fixed",
      "status": "draft",
      "notes": "string (errors)",
      "created_by": 1,
      "created_at": "2026-08-01T00:00:00Z",
      "published_at": null
    },
    "shift_templates": [
      { "id": 1, "name": "Pagi", "start_time": "08:00", "end_time": "16:00", "is_cross_day": false, "color_hex": "#3B82F6" }
    ],
    "shifts": [
      { "id": 1, "schedule_id": 1, "employee_id": 1, "shift_template_id": 1, "date": "2026-08-01", "is_override": false }
    ],
    "leaves": [
      { "id": 1, "schedule_id": 1, "employee_id": 1, "date": "2026-08-17", "is_override": false }
    ],
    "holidays": [
      { "id": 7, "date": "2026-08-17", "name": "Hari Kemerdekaan", "is_national": true }
    ],
    "fairness_metrics": {
      "total_employees": 5,
      "avg_shift_hours": 240,
      "std_dev_hours": 0,
      "std_dev_weekend_shifts": 0,
      "weekend_shifts_per_employee": { "min": 10, "max": 10, "avg": 10 }
    },
    "generation_summary": {
      "total_batches": 1,
      "completed_batches": 0,
      "failed_batches": 0,
      "total_shifts_generated": 150,
      "total_leaves_generated": 5,
      "total_errors": 0
    }
  }
}
```

---

#### `GET /api/v1/schedules/:id/validate`

Validasi jadwal terhadap constraint (waktu istirahat, komposisi role, double-booking, kuota cuti).

**Response:**
```json
{
  "success": true,
  "data": {
    "is_valid": false,
    "violations": [
      {
        "type": "rest_time_violation",
        "severity": "error",
        "message": "Jeda hanya 2 jam (min 12 jam)",
        "employee_id": 1,
        "employee_name": "Budi",
        "date": "2026-08-01"
      }
    ]
  }
}
```

---

#### `PUT /api/v1/schedules/:id/shifts`

Update shift secara manual (drag-drop override dari frontend). Validasi ulang otomatis setelah perubahan.

**Request Body:**
```json
{
  "shift_changes": [
    {
      "action": "update",
      "schedule_shift_id": 1,
      "employee_id": 2,
      "shift_template_id": 2,
      "date": "2026-08-01",
      "reason": "Tukar shift"
    }
  ],
  "leave_changes": [
    {
      "action": "create",
      "employee_id": 3,
      "date": "2026-08-15",
      "reason": "Cuti mendadak"
    }
  ]
}
```

| `action` | Deskripsi |
|---|---|
| `update` | Ubah shift yang sudah ada |
| `delete` | Hapus shift/cuti |
| `create` | Tambah shift/cuti baru |

**Response (422) jika validasi gagal:**
```json
{
  "success": false,
  "error": "VALIDATION_FAILED",
  "message": "Perubahan menyebabkan pelanggaran constraint",
  "data": { "validation_result": { "is_valid": false, "violations": [...] } }
}
```

---

#### `PUT /api/v1/schedules/:id/publish`

Publikasikan jadwal. Setelah dipublikasi, jadwal tidak bisa diedit.

**Response:**
```json
{
  "success": true,
  "data": {
    "schedule_id": 1,
    "status": "published",
    "published_at": "2026-08-01T00:00:00Z",
    "message": "Jadwal berhasil dipublikasikan"
  }
}
```

---

#### `POST /api/v1/schedules/:id/regenerate`

Generate ulang jadwal (hapus shift yang ada lalu generate baru).

**Response:**
```json
{
  "success": true,
  "data": {
    "schedule_id": 1,
    "status": "generating",
    "message": "Generate ulang dimulai."
  }
}
```

---

#### `GET /api/v1/schedules/:id/export?format=xlsx`

Ekspor jadwal. Format default `xlsx` (saat ini mengembalikan CSV).

**Response (CSV):**
```csv
Karyawan,Role,Tanggal,Shift
Budi,Supervisor,2026-08-01,Pagi
Siti,Staff,2026-08-01,Siang
```

---

#### `GET /api/v1/schedules/:id/share`

Buat tautan berbagi jadwal.

**Response:**
```json
{
  "success": true,
  "data": {
    "share_url": "https://autoshift.app/s/s1",
    "is_active": true
  }
}
```

---

## Kode Error

| HTTP | Kode | Deskripsi |
|---|---|---|
| 400 | - | Request tidak valid (body/parameter) |
| 401 | `unauthorized` | Token JWT tidak valid/kadaluarsa |
| 404 | `schedule not found` | Jadwal tidak ditemukan |
| 409 | `SCHEDULE_ALREADY_PUBLISHED` | Jadwal sudah dipublikasi |
| 422 | `VALIDATION_FAILED` | Pelanggaran constraint |
| 500 | - | Internal server error |

## Struktur Direktori

```
backend/
├── main.go              # Entry point, routes, seed data
├── config/
│   └── config.go        # Konfigurasi dari .env
├── models/
│   ├── db.go            # Koneksi PostgreSQL + AutoMigrate
│   ├── tenant.go        # Model tenant
│   ├── user.go          # Model user + login
│   ├── employee.go      # Employee, EmployeeRole, LeaveQuota
│   ├── schedule.go      # Schedule, ShiftTemplate, RoleRequirement, dll
│   └── holiday.go       # Holiday + tipe response API
├── handlers/
│   ├── schedule.go      # Handler CRUD jadwal
│   └── export.go        # Handler ekspor & berbagi
├── services/
│   ├── scheduler.go     # Orchestrator generate jadwal
│   ├── validator.go     # Validasi constraint + fairness metrics
│   └── holiday.go       # Fetch libur dari date.nager.at
├── ai/
│   └── generator.go     # Mock generator + integrasi OpenAI
└── middleware/
    └── auth.go          # Middleware JWT
```

## Constraint (Validasi)

Generator dan validator memeriksa:

| Constraint | Deskripsi |
|---|---|
| **Waktu Istirahat** | Minimal 12 jam antara shift berurutan |
| **Komposisi Role** | Setiap shift memiliki jumlah minimum role tertentu (misal: ≥ 1 Supervisor) |
| **Double-Booking** | Satu karyawan hanya satu shift per hari |
| **Kuota Cuti** | Tidak melebihi jatah cuti bulanan |

## Metrik Keadilan (Fairness)

Setelah generate, sistem menghitung:

| Metrik | Deskripsi |
|---|---|
| `avg_shift_hours` | Rata-rata total jam shift per karyawan |
| `std_dev_hours` | Standar deviasi jam shift (semakin kecil = semakin adil) |
| `std_dev_weekend_shifts` | Standar deviasi shift akhir pekan |
| `weekend_shifts_per_employee` | Statistik shift Sabtu-Minggu per karyawan |

## Generator AI

Saat `AI_PROVIDER=mock` (default):
- Membagi karyawan round-robin ke shift secara deterministik
- Tidak memperhatikan constraint waktu istirahat (validator menangkapnya)

Saat `AI_PROVIDER=openai`:
- Mengirim prompt ke OpenAI per batch (20 karyawan per batch)
- Mencoba ulang hingga 3 kali jika gagal
- Prompt berisi data lengkap: karyawan, shift template, role requirements, hari libur, constraint

## Teknologi

| Komponen | Pustaka |
|---|---|
| HTTP Framework | [Fiber v2](https://gofiber.io/) |
| ORM | [GORM v2](https://gorm.io/) |
| Database | PostgreSQL |
| Autentikasi | [golang-jwt v5](https://github.com/golang-jwt/jwt) |
| AI | OpenAI API (opsional) |
| Libur Nasional | [date.nager.at](https://date.nager.at) |
