# autoShift

**AI-powered employee shift scheduler.** Aplikasi desktop & web untuk HR / scheduler dalam mengatur shift karyawan secara bulanan dengan bantuan AI — multi-tenant, human-in-the-loop.

## Fitur Utama

- **AI Auto-Generator** — Generate jadwal shift bulanan dengan constraint: jeda 12 jam, komposisi role, kuota libur, dan distribusi adil
- **Human-in-the-Loop** — Draf → review & drag-drop → publish
- **Multi-Tenant** — Setiap perusahaan (tenant) punya data terisolasi
- **Mode Libur** — Fixed (hari tetap) atau Random (hari acak), terintegrasi public holiday API
- **Role & Komposisi** — Tiap shift bisa punya komposisi role wajib (misal: ≥1 Supervisor)
- **Fairness Metrics** — Standar deviasi jam kerja, shift weekend, dan distribusi shift
- **Export** — PDF, Excel/CSV, Public Share Link (read-only)

## Tech Stack

| Layer | Teknologi |
|-------|-----------|
| **Frontend** | React 19 + Vite 8 + Tailwind CSS 4 + shadcn/ui (new-york) |
| **Backend** | Go 1.26 + Fiber v2 + GORM |
| **Desktop** | Tauri v2 + Rust sidecar manager |
| **Database** | PostgreSQL (default), SQLite, MySQL, SQL Server |
| **AI Provider** | Mock (default), OpenAI / Ollama / vLLM |
| **Holiday** | date.nager.at Public Holiday API |

## Arsitektur

### Web Mode

```
┌──────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Frontend    │────▶│  Backend (Go)    │────▶│  Database       │
│  React 19    │     │  Fiber + GORM    │     │  PostgreSQL/    │
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

### Desktop Mode (Tauri)

```
┌─────────────────────────────────────────────┐
│           Tauri Desktop App                  │
│  ┌────────────────────┐  ┌───────────────┐  │
│  │  Webview (React)   │  │  Go Sidecar   │  │
│  │  Vite build        │◄─►│  Fiber HTTP   │  │
│  │  shadcn/ui         │  │  localhost:PORT│  │
│  └────────────────────┘  └───────┬───────┘  │
│                                  │          │
│                           ┌──────▼───────┐  │
│                           │  SQLite .db  │  │
│                           │  (app_data)  │  │
│                           └──────────────┘  │
└─────────────────────────────────────────────┘
```

Detail arsitektur lengkap ada di [architecture.md](architecture.md) dan [docs/desktop.md](docs/desktop.md).

## Struktur Proyek

```
├── src-tauri/             # Tauri desktop app (Rust)
├── frontend/              # React + Vite + Tailwind
├── backend/               # Go + Fiber + GORM
├── scripts/               # Build & utility scripts
├── docs/                  # Dokumentasi modular
│   ├── quickstart.md      # Panduan instalasi & konfigurasi
│   ├── desktop.md         # Detail desktop app & sidecar
│   └── constraints.md     # AI constraints & workflow
├── PRD.md                 # Product Requirements (Indonesian)
├── architecture.md        # System architecture + diagrams
└── database_schema.sql    # Reference SQL schema
```

## Quick Start

```bash
# Frontend
cd frontend && npm install && npm run dev

# Backend (SQLite — tanpa DB eksternal)
cp backend/.env.example backend/.env
# edit DB_DRIVER=sqlite
cd backend && go run main.go
```

Panduan lengkap: [docs/quickstart.md](docs/quickstart.md)

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

Dokumentasi lengkap: [docs/](docs/)

## Lisensi

Distribusi di bawah lisensi **MIT**. Lihat [LICENSE](LICENSE) untuk detail.
