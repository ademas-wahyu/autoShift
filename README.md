# autoShift

**AI-powered employee shift scheduler.** Aplikasi web untuk HR / scheduler dalam mengatur shift karyawan secara bulanan dengan bantuan AI — multi-tenant, human-in-the-loop.

---

## Fitur Utama

- **AI Auto-Generator** — Generate jadwal shift bulanan dengan constraint: jeda 12 jam, komposisi role, kuota libur, dan distribusi adil
- **Human-in-the-Loop** — Draf → review & drag-drop → publish
- **Multi-Tenant** — Setiap perusahaan (tenant) punya data terisolasi
- **Mode Libur** — Fixed (hari tetap) atau Random (hari acak), terintegrasi public holiday API
- **Role & Komposisi** — Tiap shift bisa punya komposisi role wajib (misal: ≥1 Supervisor)
- **Fairness Metrics** — Standar deviasi jam kerja, shift weekend, dan distribusi shift
- **Export** — PDF, Excel/CSV, Public Share Link (read-only)

---

## Arsitektur

```
┌──────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Frontend    │────▶│  Backend (Go)    │────▶│  Database       │
│  React 19    │     │  Fiber + GORM    │     │  PostgreSQL/     │
│  Vite 8      │     │  Batch Processor │     │  SQLite/dll     │
│  Tailwind 4  │◀────│  Validator       │◀────│                 │
│  shadcn/ui   │     │  AI Generator    │     └─────────────────┘
└──────────────┘     └────────┬─────────┘
                              │
                     ┌────────▼─────────┐
                     │  AI LLM          │
                     │  (OpenAI/Ollama/ │
                     │   Mock)          │
                     └──────────────────┘
```

- **Batch Processing**: Generate per 20 karyawan per batch, digabung & divalidasi backend
- **Validation Loop**: Validasi jeda 12h, role, double-booking, kuota libur — retry max 3x
- **Holiday API**: [date.nager.at](https://date.nager.at) — gratis, tanpa auth

---

## Tech Stack

| Layer | Teknologi |
|-------|-----------|
| **Frontend** | React 19 + Vite 8 + Tailwind CSS 4 + shadcn/ui (new-york) |
| **Backend** | Go 1.26 + Fiber v2 + GORM |
| **Database** | PostgreSQL (default), SQLite, MySQL, SQL Server |
| **AI Provider** | Mock (default), OpenAI / Ollama / vLLM |
| **Holiday** | date.nager.at Public Holiday API |

---

## Quick Start

### Prasyarat

- **Go** 1.26+
- **Node.js** 20+
- **npm** 10+

### Backend (SQLite — tanpa DB eksternal)

```bash
cp backend/.env.example backend/.env
# Edit DB_DRIVER=sqlite (atau set via env)
cd backend && go run main.go
```

### Frontend

```bash
cd frontend && npm install && npm run dev
```

Backend berjalan di `http://localhost:8080`, frontend di `http://localhost:5173`.

---

## Konfigurasi

Atur via environment variable atau `backend/.env`:

| Variabel | Default | Deskripsi |
|----------|---------|-----------|
| `DB_DRIVER` | `postgres` | `postgres` / `sqlite` / `mysql` / `sqlserver` |
| `SERVER_PORT` | `8080` | Port server |
| `JWT_SECRET` | — | Secret untuk JWT auth |
| `AI_PROVIDER` | `mock` | `mock` / `openai` |
| `AI_API_URL` | — | Endpoint OpenAI-compatible (Ollama, vLLM, dll) |
| `AI_MODEL` | `gpt-4o` | Model AI |
| `HOLIDAY_API_URL` | `https://date.nager.at/api/v3` | Holiday API |

Backend hardcodes: `BatchSize: 20`, `MinRestHours: 12`, `MaxRetries: 3`.

---

## API Endpoints

Semua endpoint di bawah `/api/v1`:

| Method | Endpoint | Auth | Fungsi |
|--------|----------|------|--------|
| `GET` | `/health` | ❌ | Health check |
| `POST` | `/login` | ❌ | Login |
| `POST` | `/schedules` | ✅ | Generate jadwal baru |
| `GET` | `/schedules/:id` | ✅ | Detail jadwal |
| `PUT` | `/schedules/:id/shifts` | ✅ | Update shift (drag-drop) |
| `PUT` | `/schedules/:id/publish` | ✅ | Publish jadwal |
| `GET` | `/schedules/:id/export` | ✅ | Export PDF/Excel |
| `GET` | `/schedules/:id/share` | ✅ | Public share link |
| `GET` | `/holidays` | ❌ | Daftar tanggal merah |

---

## Constraints AI

| Constraint | Penjelasan |
|------------|------------|
| **Jeda Istirahat** | Minimal 12 jam antar shift berurutan |
| **Role Composition** | Setiap shift terisi role yang dibutuhkan (misal: ≥1 Supervisor) |
| **No Double-Booking** | 1 karyawan = 1 shift per hari |
| **Leave Quota** | Kuota libur per karyawan/bulan tidak terlampaui |
| **Fairness** | Distribusi shift merata (diukur via std dev) |

---

## Alur Kerja

```
Admin pilih bulan & konfigurasi → AI Generate (batch) → Validasi
    → Draf (Pending Review) → Review & Drag-Drop → Publish → Export / Share
```

---

## Struktur Proyek

```
├── frontend/              # React + Vite + Tailwind
│   ├── src/
│   │   ├── components/    # UI components (shadcn/ui)
│   │   ├── hooks/         # Custom hooks
│   │   ├── lib/           # Utilitas
│   │   └── types/         # TypeScript types
│   └── ...
├── backend/               # Go + Fiber + GORM
│   ├── main.go            # Entrypoint
│   ├── ai/                # AI generator (mock / openai)
│   ├── config/            # Konfigurasi aplikasi
│   ├── handlers/          # Route handlers
│   ├── middleware/        # Auth middleware (JWT)
│   ├── models/            # GORM models + migrasi
│   └── services/          # Scheduler, validator, holiday
├── PRD.md                 # Product Requirements (Indonesian)
├── architecture.md        # System architecture + diagrams
├── api_contract.md        # API contract documentation
└── database_schema.sql    # Reference SQL schema
```

---

## Lisensi

Proyek internal — belum memiliki lisensi resmi.
