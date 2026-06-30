# API Contract: autoShift

Base URL: `/api/v1`
Format: JSON
Auth: Bearer Token (JWT)

---

## Daftar Endpoint

| # | Method | Path | Fungsi |
|---|---|---|---|
| 1 | `POST` | `/schedules` | Buat schedule + trigger AI generate |
| 2 | `GET` | `/schedules/:id` | Ambil detail schedule (shift + libur) |
| 3 | `GET` | `/schedules/:id/validate` | Validasi manual tanpa generate ulang |
| 4 | `PUT` | `/schedules/:id/shifts` | Update shift (drag-drop manual) |
| 5 | `PUT` | `/schedules/:id/publish` | Publish jadwal |
| 6 | `GET` | `/schedules/:id/export` | Export (PDF/Excel) |
| 7 | `POST` | `/schedules/:id/regenerate` | Generate ulang batch tertentu |
| 8 | `GET` | `/schedules/:id/share` | Public share link (read-only) |

---

## 1. POST /schedules — Buat Schedule + Generate

Endpoint utama. Frontend kirim semua konfigurasi, backend simpan draft lalu trigger AI.

### Request

```jsonc
{
  "tenant_id": 1,
  "month": 7,
  "year": 2026,

  // Mode libur
  "leave_mode": "random",         // "fixed" | "random"
  "fixed_leaves": null,           // jika "fixed": [0, 6] (Minggu & Sabtu)
  "random_days_per_week": null,    // jika "random": 2

  // Shift templates yang aktif bulan ini
  "shift_templates": [
    {
      "id": 1,
      "capacity": 5                // jumlah karyawan untuk shift ini
    },
    {
      "id": 2,
      "capacity": 4
    },
    {
      "id": 3,
      "capacity": 3
    }
  ],

  // Role requirements (dari shift_role_requirements, bisa di-override per schedule)
  "role_requirements": [
    { "shift_template_id": 1, "role_id": 1, "min_count": 1 },
    { "shift_template_id": 2, "role_id": 1, "min_count": 1 },
    { "shift_template_id": 3, "role_id": 1, "min_count": 1 }
  ],

  // Karyawan aktif
  "employee_ids": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
}
```

### Response (201 Created)

```jsonc
{
  "success": true,
  "data": {
    "schedule_id": 42,
    "status": "generating",        // "generating" | "draft" | "published"
    "message": "Jadwal sedang di-generate. Proses batch 1/1 selesai."
  }
}
```

---

## 2. GET /schedules/:id — Ambil Detail Schedule

Digunakan frontend untuk render kalender setelah generate selesai.

### Response (200 OK)

```jsonc
{
  "success": true,
  "data": {
    "schedule": {
      "id": 42,
      "tenant_id": 1,
      "month": 7,
      "year": 2026,
      "leave_mode": "random",
      "status": "draft",
      "notes": null,
      "created_by": 1,
      "created_at": "2026-06-30T10:00:00Z",
      "published_at": null
    },

    // Shift templates yang dipakai bulan ini
    "shift_templates": [
      {
        "id": 1,
        "name": "Pagi",
        "start_time": "08:00",
        "end_time": "16:00",
        "is_cross_day": false,
        "color_hex": "#3B82F6",
        "capacity": 5
      },
      {
        "id": 2,
        "name": "Siang",
        "start_time": "16:00",
        "end_time": "00:00",
        "is_cross_day": true,
        "color_hex": "#F59E0B",
        "capacity": 4
      },
      {
        "id": 3,
        "name": "Malam",
        "start_time": "22:00",
        "end_time": "06:00",
        "is_cross_day": true,
        "color_hex": "#6366F1",
        "capacity": 3
      }
    ],

    // Semua shift assignment per hari
    "shifts": [
      {
        "id": 1001,
        "employee_id": 1,
        "employee_name": "Budi",
        "employee_role": "Supervisor",
        "shift_template_id": 1,
        "shift_name": "Pagi",
        "date": "2026-07-01",
        "is_override": false
      }
      // ... banyak baris
    ],

    // Hari libur yang sudah di-plot AI
    "leaves": [
      {
        "id": 501,
        "employee_id": 1,
        "employee_name": "Budi",
        "date": "2026-07-02",
        "is_override": false
      }
      // ... banyak baris
    ],

    // Hari libur nasional (dari API)
    "holidays": [
      {
        "date": "2026-07-05",
        "name": "Tahun Baru Islam",
        "is_national": true
      }
    ],

    // Error/warning dari AI generation (jika ada)
    "generation_errors": [
      {
        "batch": 1,
        "type": "rest_time_violation",
        "message": "Budi (Supervisor) shift Malam 1 Juli 22:00-06:00 lalu Pagi 2 Juli 08:00-16:00 — jeda hanya 2 jam, perlu manual fix.",
        "employee_id": 1,
        "date": "2026-07-02",
        "shift_template_id": 1
      }
    ],

    // Metrik keadilan
    "fairness_metrics": {
      "total_employees": 10,
      "avg_shift_hours": 176.0,
      "std_dev_hours": 4.2,
      "std_dev_weekend_shifts": 0.8,
      "weekend_shifts_per_employee": {
        "min": 2,
        "max": 4,
        "avg": 3.1
      }
    },

    "generation_summary": {
      "total_batches": 1,
      "completed_batches": 1,
      "failed_batches": 0,
      "total_shifts_generated": 210,
      "total_leaves_generated": 60,
      "total_errors": 1
    }
  }
}
```

---

## 3. GET /schedules/:id/validate — Validasi Manual

Cek semua constraint tanpa generate ulang. Dipanggil setelah admin drag-drop.

### Response (200 OK)

```jsonc
{
  "success": true,
  "data": {
    "is_valid": false,
    "violations": [
      {
        "type": "rest_time_violation",
        "severity": "error",          // "error" | "warning"
        "message": "Budi: shift Malam 1 Jul → Pagi 2 Jul, jeda hanya 2 jam",
        "employee_id": 1,
        "employee_name": "Budi",
        "date": "2026-07-02"
      },
      {
        "type": "role_composition",
        "severity": "error",
        "message": "Shift Siang 3 Jul 2026: tidak ada Supervisor",
        "date": "2026-07-03",
        "shift_template_id": 2
      },
      {
        "type": "leave_quota_exceeded",
        "severity": "warning",
        "message": "Siti: kuota libur 2 hari, sekarang 3 hari",
        "employee_id": 2,
        "employee_name": "Siti",
        "quota": 2,
        "actual": 3
      },
      {
        "type": "double_booking",
        "severity": "error",
        "message": "Andi: terdaftar shift Pagi dan libur di 5 Jul 2026",
        "employee_id": 3,
        "employee_name": "Andi",
        "date": "2026-07-05"
      }
    ],
    "violation_count": 4
  }
}
```

---

## 4. PUT /schedules/:id/shifts — Update Shift (Drag-Drop)

Frontend kirim daftar perubahan setelah admin melakukan drag-drop di kalender.

### Request

```jsonc
{
  // Perubahan shift (pindah/ganti/tambah/hapus)
  "shift_changes": [
    {
      "action": "update",             // "update" | "delete" | "create"
      "schedule_shift_id": 1001,       // null jika "create"
      "employee_id": 1,
      "shift_template_id": 2,          // pindah dari Pagi → Siang
      "date": "2026-07-02",
      "reason": "Manual override by admin"
    },
    {
      "action": "delete",
      "schedule_shift_id": 1002,
      "date": "2026-07-03"
    },
    {
      "action": "create",
      "employee_id": 5,
      "shift_template_id": 1,
      "date": "2026-07-04"
    }
  ],

  // Perubahan hari libur
  "leave_changes": [
    {
      "action": "create",              // jadikan libur
      "employee_id": 1,
      "date": "2026-07-05"
    },
    {
      "action": "delete",              // hapus libur, jadi kerja
      "schedule_leave_id": 501
    }
  ]
}
```

### Response (200 OK)

```jsonc
{
  "success": true,
  "data": {
    "applied_changes": 5,
    "validation_result": {
      "is_valid": true,
      "violations": []
    }
  }
}
```

Jika perubahan menyebabkan pelanggaran:

```jsonc
{
  "success": false,
  "error": "VALIDATION_FAILED",
  "message": "Perubahan menyebabkan 2 pelanggaran constraint",
  "data": {
    "applied_changes": 3,
    "rejected_changes": 2,
    "validation_result": {
      "is_valid": false,
      "violations": [
        {
          "type": "rest_time_violation",
          "severity": "error",
          "message": "Budi: shift Siang 2 Jul 16:00-00:00 → Pagi 3 Jul 08:00-16:00, jeda 8 jam",
          "employee_id": 1,
          "date": "2026-07-03"
        }
      ]
    }
  }
}
```

---

## 5. PUT /schedules/:id/publish — Publish Jadwal

### Request (body opsional)

```jsonc
{
  "notes": "Jadwal Juli 2026, sudah di-review dan di-approve"
}
```

### Response (200 OK)

```jsonc
{
  "success": true,
  "data": {
    "schedule_id": 42,
    "status": "published",
    "published_at": "2026-06-30T14:30:00Z",
    "message": "Jadwal berhasil dipublikasikan"
  }
}
```

---

## 6. GET /schedules/:id/export — Export

### Query

```
GET /api/v1/schedules/42/export?format=pdf
GET /api/v1/schedules/42/export?format=xlsx
```

### Response

File binary (PDF / Excel), Content-Type sesuai.

---

## 7. POST /schedules/:id/regenerate — Generate Ulang

Digunakan jika admin ingin AI generate ulang untuk batch tertentu (misal setelah manual edit ternyata kacau).

### Request

```jsonc
{
  "employee_ids": [1, 2, 3],    // kosongkan untuk regenerate semua
  "date_from": "2026-07-15",
  "date_to": "2026-07-21"       // regenerate seminggu
}
```

### Response (200 OK)

Sama seperti POST /schedules.

---

## 8. GET /schedules/:id/share — Public Share Link

### Response (200 OK)

```jsonc
{
  "success": true,
  "data": {
    "share_url": "https://autoshift.app/s/abc123def456",
    "expires_at": "2026-08-31T23:59:59Z",
    "is_active": true
  }
}
```

---

## LAMPIRAN: AI Prompt → Response Contract

Ini format prompt yang dikirim backend ke AI.

### Prompt ke AI

```jsonc
{
  "model": "gpt-4o",
  "messages": [
    {
      "role": "system",
      "content": "Kamu adalah asisten penjadwal shift. Tugasmu membuat jadwal shift yang ADIL dan SEIMBANG. Output hanya JSON array, tanpa teks lain."
    },
    {
      "role": "user",
      "content": "---
batch_info:
  batch_index: 1
  total_batches: 3
  period_start: '2026-07-06'
  period_end: '2026-07-12'

shift_templates:
  - id: 1
    name: Pagi
    start_time: '08:00'
    end_time: '16:00'
    is_cross_day: false
    capacity: 5
  - id: 2
    name: Siang
    start_time: '16:00'
    end_time: '00:00'
    is_cross_day: true
    capacity: 4
  - id: 3
    name: Malam
    start_time: '22:00'
    end_time: '06:00'
    is_cross_day: true
    capacity: 3

leave_mode: random
random_days_per_week: 2
holidays: []

role_requirements:
  - shift_template_id: 1
    role_id: 1
    min_count: 1
  - shift_template_id: 2
    role_id: 1
    min_count: 1
  - shift_template_id: 3
    role_id: 1
    min_count: 1

constraints:
  min_rest_hours: 12

employees:
  - id: 1
    name: Budi
    role_id: 1
    role_name: Supervisor
    leave_quota: 2
  - id: 2
    name: Siti
    role_id: 2
    role_name: Staff
    leave_quota: 2
  - id: 3
    name: Andi
    role_id: 2
    role_name: Staff
    leave_quota: 2
  - id: 4
    name: Dewi
    role_id: 2
    role_name: Staff
    leave_quota: 2
  - id: 5
    name: Rudi
    role_id: 1
    role_name: Supervisor
    leave_quota: 2

previous_errors: []"
    }
  ],
  "response_format": { "type": "json_object" },
  "temperature": 0.3
}
```

### Response Wajib dari AI

```jsonc
{
  "shifts": [
    {
      "employee_id": 1,
      "date": "2026-07-06",
      "shift_template_id": 1
    },
    {
      "employee_id": 2,
      "date": "2026-07-06",
      "shift_template_id": 1
    }
  ],
  "leaves": [
    {
      "employee_id": 3,
      "date": "2026-07-06"
    },
    {
      "employee_id": 4,
      "date": "2026-07-06"
    }
  ],
  "batch_info": {
    "batch_index": 1,
    "shift_count": 25,
    "leave_count": 10
  }
}
```

Aturan ketat:
- `shifts[]` — setiap entry = 1 karyawan × 1 tanggal × 1 shift template
- `leaves[]` — setiap entry = 1 karyawan × 1 tanggal libur
- 1 karyawan di 1 tanggal: **HANYA 1 entry** — shifts ATAU leaves, tidak keduanya
- Total shifts = (total hari kerja per minggu × total kapasitas shift per hari) — leaves yang sudah di-plot

---

## LAMPIRAN: Error Code Referensi

| Code | HTTP | Deskripsi |
|---|---|---|
| `VALIDATION_FAILED` | 422 | Perubahan melanggar constraint |
| `SCHEDULE_NOT_FOUND` | 404 | Schedule tidak ditemukan |
| `SCHEDULE_ALREADY_PUBLISHED` | 409 | Tidak bisa edit jadwal yang sudah publish |
| `AI_GENERATION_FAILED` | 500 | AI gagal generate setelah 3x retry |
| `INVALID_BATCH_CONFIG` | 400 | Konfigurasi batch tidak valid |
| `UNAUTHORIZED` | 401 | Token tidak valid |
| `FORBIDDEN` | 403 | Tidak punya akses ke tenant ini |
