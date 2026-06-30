# System Architecture & AI Flow: autoShift

---

## 1. Arsitektur Keseluruhan

```mermaid
flowchart TB
    subgraph Frontend ["Frontend (Next.js / React)"]
        ConfigPage["Konfigurasi Shift & Libur"]
        CalendarUI["Kalender Jadwal (Drag-Drop)"]
        ExportUI["Export PDF / Excel / Share Link"]
    end

    subgraph Backend ["Backend (Go Fiber / Laravel)"]
        API["REST API"]
        SchedulerEngine["Scheduler Engine"]
        Validator["Validator Constraints"]
        HolidayFetcher["Holiday API Fetcher"]
    end

    subgraph AI_Layer ["AI Layer"]
        BatchProcessor["Batch Processor\n(per minggu / per 20 karyawan)"]
        LLM["LLM (OpenAI / Ollama)"]
        PromptBuilder["Prompt Builder\n(constraints → prompt)"]
    end

    subgraph Storage ["Storage"]
        DB[("PostgreSQL")]
    end

    subgraph External ["External"]
        HolidayAPI[("Public Holiday API\ndate.nager.at)")]
    end

    ConfigPage -->|"POST /api/generate"| API
    CalendarUI -->|"PUT /api/shifts"| API
    ExportUI -->|"GET /api/export"| API

    API --> SchedulerEngine
    SchedulerEngine --> PromptBuilder
    SchedulerEngine --> Validator
    SchedulerEngine --> HolidayFetcher
    
    PromptBuilder --> BatchProcessor
    BatchProcessor -->|"batch prompt"| LLM
    LLM -->|"batch result"| BatchProcessor
    
    BatchProcessor -->|"gabung hasil"| Validator
    Validator -->|"valid ? simpan : retry"| DB

    HolidayFetcher -->|"fetch"| HolidayAPI
    HolidayAPI -->|"holidays"| HolidayFetcher
    HolidayFetcher --> DB

    API -->|"read/write"| DB
```

---

## 2. Alur Generate Jadwal (End-to-End)

```mermaid
sequenceDiagram
    actor Admin as Admin
    participant FE as Frontend
    participant BE as Backend
    participant AI as AI LLM
    participant DB as Database

    Admin->>FE: Pilih bulan & konfigurasi
    FE->>BE: POST /api/schedules (config)
    BE->>DB: Simpan config schedules
    BE->>BE: Ambil data: employees, roles, shift_templates, holidays

    Note over BE: --- BATCH LOOP: per 20 karyawan ---

    loop Setiap batch (20 karyawan)
        BE->>BE: Build prompt + constraints
        BE->>AI: Request generate batch
        AI-->>BE: Response batch jadwal
        BE->>BE: Parse & validasi batch
    end

    Note over BE: --- VALIDATION LOOP ---

    BE->>BE: Gabung semua batch
    BE->>BE: Validasi: jeda 12h, role, dobel, kuota
    
    alt Validasi Gagal
        BE->>BE: Flag error + lokasi conflict
        BE->>BE: Retry batch bermasalah (max 3x)
        alt Retry habis
            BE->>BE: Simpan partial + error list
        end
    end

    BE->>DB: Simpan schedule_shifts + schedule_employee_leaves
    BE-->>FE: Response: {status: "draft", errors: [...]}
    FE-->>Admin: Tampilkan kalender (Status: Pending Review)

    Admin->>FE: Review & drag-drop jika perlu
    FE->>BE: PUT /api/schedules/:id/publish
    BE->>DB: Update status → "published"
    BE-->>FE: Response: {status: "published"}
    FE-->>Admin: Jadwal final
```

---

## 3. Validation Loop Detail

```mermaid
flowchart TD
    Start(["Mulai Generate"]) --> BatchLoop["Loop per batch (20 karyawan)"]
    BatchLoop --> BuildPrompt["Build prompt\n(employees, shifts, roles, holidays, constraints)"]
    BuildPrompt --> CallAI["Call AI LLM"]
    CallAI --> Parse["Parse response"]
    Parse --> ValidateBatch["Validasi batch:"]

    ValidateBatch --> CheckRest["Cek jeda istirahat 12h"]
    ValidateBatch --> CheckRole["Cek komposisi role per shift"]
    ValidateBatch --> CheckDouble["Cek dobel data (kerja + libur)"]
    ValidateBatch --> CheckQuota["Cek kuota libur terlampaui"]

    CheckRest --> PassRest{"Lolos?"}
    PassRest -->|Ya| CR
    PassRest -->|Tidak| FlagRest["Flag conflict shift berurutan"]
    
    CheckRole --> PassRole{"Lolos?"}
    PassRole -->|Ya| CR2
    PassRole -->|Tidak| FlagRole["Flag role tidak terpenuhi"]

    CR --> CR2
    FlagRest --> CR2

    CheckDouble --> PassDouble{"Lolos?"}
    PassDouble -->|Ya| CR3
    PassDouble -->|Tidak| FlagDouble["Flag double-booking"]

    CheckQuota --> PassQuota{"Lolos?"}
    PassQuota -->|Ya| CR4
    PassQuota -->|Tidak| FlagQuota["Flag kuota libur lewat"]

    CR2 --> CheckDouble
    CR3 --> CheckQuota

    FlagRole --> CR3
    FlagDouble --> CR4
    FlagQuota --> CheckRetry

    CR4 --> CheckRetry{Ada error?}
    CheckRetry -->|Tidak ada| Simpan["Simpan ke database"]
    CheckRetry -->|Ada| RetryCount{Retry < 3x?}
    RetryCount -->|Ya| FixPrompt["Tambah error context ke prompt"]
    FixPrompt --> CallAI
    RetryCount -->|Tidak| SimpanPartial["Simpan partial + daftar conflict manual"]
    SimpanPartial --> End([Selesai - Status: Draft + Warnings])
    Simpan --> End2([Selesai - Status: Draft])
```

---

## 4. REST API Endpoints

| Method | Endpoint | Fungsi |
|---|---|---|
| `POST` | `/api/schedules` | Buat schedule baru + trigger AI generate |
| `GET` | `/api/schedules/:id` | Ambil detail schedule + semua shift |
| `PUT` | `/api/schedules/:id/shifts` | Update shift (drag-drop manual) |
| `PUT` | `/api/schedules/:id/publish` | Publish jadwal (draft → published) |
| `GET` | `/api/schedules/:id/export` | Export PDF/Excel |
| `GET` | `/api/schedules/:id/share` | Generate public share link (read-only) |
| `GET` | `/api/holidays?year=2026` | Ambil daftar tanggal merah |

---

## 5. Batch Processing Strategy

```
Input: 200 employees × 30 hari
                │
                ▼
        Split per 20 karyawan
                │
        ┌───────┼───────┐
        ▼       ▼       ▼
    Batch 1  Batch 2  Batch 3 ... Batch 10
    (20 org) (20 org) (20 org)   (20 org)
        │       │       │           │
        ▼       ▼       ▼           ▼
    AI gen   AI gen   AI gen      AI gen
        │       │       │           │
        └───────┼───────┼───────────┘
                ▼       ▼
            Validator Gabung
                │
                ▼
          Simpan ke DB
```

Setiap batch adalah 1 prompt AI yang berisi:
- 20 karyawan + role mereka
- 1 minggu penuh (7 hari) — atau 30 hari jika karyawannya sedikit
- Shift templates + jam kerja
- Mode libur + holidays
- Constraints: jeda 12h, komposisi role, kuota libur
- **Error context dari batch sebelumnya** (jika retry)

---

## 6. Prompt Structure (Template)

```
Kamu adalah asisten penjadwal shift. Buat jadwal yang ADIL.

KONFIGURASI:
- Periode: 1-7 Juli 2026
- Shift: Pagi (08:00-16:00), Siang (16:00-00:00), Malam (22:00-06:00, cross-day)
- Mode libur: Random, 2 hari/minggu (tidak terpengaruh tanggal merah)
- Tanggal merah: 0 (tidak ada di minggu ini)
- Jeda minimal antar shift: 12 jam

KARYAWAN (batch 1/10):
1. Budi (Supervisor) - kuota libur: 2 hari
2. Siti (Staff) - kuota libur: 2 hari
3. Andi (Staff) - kuota libur: 2 hari
...

ROLE REQUIREMENTS:
- Shift Pagi: min 1 Supervisor
- Shift Siang: min 1 Supervisor
- Shift Malam: min 1 Supervisor

HASIL GENERATE SEBELUMNYA (untuk batch 1):
(Batch ini adalah batch pertama, tidak ada konflik sebelumnya)

FORMAT OUTPUT (JSON):
{
  "shifts": [
    {"employee_id": 1, "date": "2026-07-01", "shift_template_id": 1},
    ...
  ],
  "leaves": [
    {"employee_id": 1, "date": "2026-07-02"},
    ...
  ]
}
```

---

## 7. Error Handling & Retry Strategy

```mermaid
flowchart LR
    A[AI Response] --> V{Validator}
    V -->|OK| DB[(Database)]
    V -->|Conflict| C{Retry < 3?}
    C -->|Ya| R[Rebuild Prompt + Error Context]
    R --> A
    C -->|Tidak| W[Simpan Partial + Conflict List]
    W --> UI[Admin lihat conflict → manual fix]
```

| Error Type | Retry Action | Final Action |
|---|---|---|
| Jeda istirahat < 12h | Tambah constraint ke prompt | Flag manual fix |
| Role kurang (no supervisor) | Tambah role requirement ke prompt | Flag manual fix |
| Kuota libur terlampaui | Kurangi jatah libur di prompt | Tampilkan warning |
| Double-booking | Prioritaskan shift, hapus libur | Tampilkan warning |
| Format JSON tidak valid | Re-prompt dengan format tegas | Error ke admin |
| AI timeout/hallucination | Retry dengan batch lebih kecil | Simpan partial |

---

## 8. Tech Stack Recommendation

| Layer | Opsi 1 (Recommended) | Opsi 2 | Alasan |
|---|---|---|---|
| Frontend | Next.js 14+ (React) | Vue 3 + Nuxt | Komunitas besar, kalender library banyak |
| Backend | Go Fiber | Laravel | Performa tinggi untuk batch processing |
| Database | PostgreSQL | MySQL | Dukungan INTERVAL, TIMESTAMP, ENUM |
| ORM | GORM (Go) / sqlc | Eloquent (Laravel) | — |
| AI LLM | OpenAI GPT-4o / Ollama (local) | Claude API | Pilih based on budget & privacy |
| Holiday API | date.nager.at | GitHub public API | Gratis, no auth needed |
| UI Calendar | @fullcalendar/react (drag-drop) | Custom | Sudah support drag & drop bawaan |
