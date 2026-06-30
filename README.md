# autoShift

**AI-powered employee shift scheduler.** Aplikasi desktop & web untuk HR / scheduler dalam mengatur shift karyawan secara bulanan dengan bantuan AI — multi-tenant, human-in-the-loop.

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

### Web Mode
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

- **Batch Processing**: Generate per 20 karyawan per batch, digabung & divalidasi backend
- **Validation Loop**: Validasi jeda 12h, role, double-booking, kuota libur — retry max 3x
- **Holiday API**: [date.nager.at](https://date.nager.at) — gratis, tanpa auth

---

## Tech Stack

| Layer | Teknologi |
|-------|-----------|
| **Frontend** | React 19 + Vite 8 + Tailwind CSS 4 + shadcn/ui (new-york) |
| **Backend** | Go 1.26 + Fiber v2 + GORM |
| **Desktop** | Tauri v2 + Rust sidecar manager |
| **Database** | PostgreSQL (default), SQLite, MySQL, SQL Server |
| **AI Provider** | Mock (default), OpenAI / Ollama / vLLM |
| **Holiday** | date.nager.at Public Holiday API |

---

## Quick Start

### Prasyarat

- **Go** 1.26+
- **Node.js** 20+
- **npm** 10+
- **Rust** 1.80+ (untuk desktop)
- **Tauri CLI**: `cargo install tauri-cli --version "^2"`

### Backend (SQLite — tanpa DB eksternal)

```bash
cp backend/.env.example backend/.env
# Edit DB_DRIVER=sqlite (atau set via env)
cd backend && go run main.go
```

### Frontend (Web Dev)

```bash
cd frontend && npm install && npm run dev
```

### Desktop (Tauri Dev)

```bash
# 1. Build Go sidecar
bash scripts/build-sidecar.sh

# 2. Jalankan Tauri dev (otomatis build frontend + start sidecar)
PKG_CONFIG_PATH=/usr/lib/x86_64-linux-gnu/pkgconfig:/usr/share/pkgconfig \
  cargo tauri dev
```

### Production Build

```bash
# Build Go sidecar untuk semua target
bash scripts/build-sidecar.sh

# Build Tauri app (output di src-tauri/target/release/bundle/)
cargo tauri build
```

Backend berjalan di `http://localhost:8080`, frontend di `http://localhost:1420` (web) atau dalam Tauri window (desktop).

---

---

## Desktop App (Tauri)

### Arsitektur Runtime

```
User buka app
  ↓
Tauri webview start
  ↓
Rust sidecar manager:
  1. Cari port kosong (TcpListener::bind)
  2. Spawn Go binary --env PORT=xxxx --env DB_PATH=<app_data>/autoshift.db
  3. Poll /api/v1/health sampai ready (timeout 10s)
  4. Register Tauri command get_api_port()
  ↓
React frontend invoke get_api_port() → http://localhost:{PORT}/api/v1/...
  ↓
Go backend serve, read/write SQLite
  ↓
User tutup app → Rust kill Go process → selesai
```

### Flow Runtime Detail

| Tahap | Komponen | Aksi |
|-------|----------|------|
| 1 | **Rust** | `find_free_port()` — bind ke `127.0.0.1:0` untuk dapat port kosong |
| 2 | **Rust** | `sidecar("autoshift-server").env("PORT", port).env("DB_PATH", path).env("DB_DRIVER", "sqlite")` |
| 3 | **Rust** | `wait_for_health(port, 10s)` — polling GET `/api/v1/health` tiap 200ms |
| 4 | **Rust** | Jika health check gagal → kill sidecar + exit(1) |
| 5 | **Rust** | `#[tauri::command] get_api_port() -> u16` — exposed ke frontend |
| 6 | **Frontend** | Deteksi `__TAURI_INTERNALS__` → `invoke('get_api_port')` → dynamic base URL |
| 7 | **Frontend** | Fallback ke `VITE_API_URL` jika bukan di Tauri (web mode) |
| 8 | **Rust** | `on_window_event(CloseRequested)` → sidecar otomatis di-kill |

### Sidecar Manager (Rust)

File: `src-tauri/src/lib.rs`

```rust
#[tauri::command]
fn get_api_port(state: tauri::State<'_, Mutex<ApiPort>>) -> u16 {
    state.lock().unwrap().0
}
```

### Go Graceful Shutdown

Backend Go menerima sinyal SIGTERM/SIGINT untuk shutdown bersih saat sidecar di-kill oleh Tauri:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
app.Shutdown()
```

### Frontend Auto-Detect

File: `frontend/src/lib/api/index.ts`

```typescript
if (typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window) {
  const { invoke } = await import('@tauri-apps/api/core')
  const port = await invoke<number>('get_api_port')
  base = `http://localhost:${port}/api/v1`
} else {
  base = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
}
```

### Build Sidecar

Script `scripts/build-sidecar.sh` otomatis mendeteksi OS & arsitektur, lalu compile Go binary ke `src-tauri/binaries/autoshift-server-{target-triple}`.

Target triple yang didukung:
| OS | Arch | Target Triple |
|----|------|---------------|
| Linux | x86_64 | `x86_64-unknown-linux-gnu` |
| macOS | Apple Silicon | `aarch64-apple-darwin` |
| Windows | x86_64 | `x86_64-pc-windows-msvc` |

Ukuran binary ~19MB (bisa di-strip/UPX jadi ~8MB).

### Catatan Penting

| Item | Keterangan |
|------|------------|
| **Zero rewrite backend** | Semua kode Go tetap utuh, hanya tambah signal handler |
| **Frontend minimal berubah** | Hanya `api/index.ts` (dynamic base URL) dan `vite.config.ts` (port 1420) |
| **Database** | SQLite auto-create di app data directory. Migrasi & seed via GORM tetap sama |
| **Bundle size** | ~25MB final (Go binary + frontend + Tauri runtime) |
| **Cross-platform** | Go + Tauri support Linux, macOS, Windows |

---

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
├── src-tauri/             # Tauri desktop app (Rust)
│   ├── src/
│   │   ├── main.rs        # Entrypoint → panggil autoshift_lib::run()
│   │   └── lib.rs         # Sidecar lifecycle: find_port → spawn → health_check → command
│   ├── Cargo.toml         # Dependencies: tauri v2, tauri-plugin-shell, reqwest, tokio
│   ├── tauri.conf.json    # Window (1280x800), bundle, CSP, externalBin config
│   ├── capabilities/
│   │   └── default.json   # Tauri v2 permissions (shell:spawn, core:default)
│   ├── build.rs           # Tauri build script
│   ├── icons/             # App icons (32x32, 128x128, icns, ico — generated)
│   └── binaries/          # Go sidecar binary per target (build-sidecar.sh)
├── frontend/              # React + Vite + Tailwind
│   ├── src/
│   │   ├── components/    # UI components (shadcn/ui)
│   │   ├── hooks/         # Custom hooks
│   │   ├── lib/           # Utilitas (api client auto-detect Tauri)
│   │   └── types/         # TypeScript types
│   └── ...
├── backend/               # Go + Fiber + GORM
│   ├── main.go            # Entrypoint (graceful shutdown)
│   ├── ai/                # AI generator (mock / openai)
│   ├── config/            # Konfigurasi aplikasi
│   ├── handlers/          # Route handlers
│   ├── middleware/        # Auth middleware (JWT)
│   ├── models/            # GORM models + migrasi
│   └── services/          # Scheduler, validator, holiday
├── scripts/
│   └── build-sidecar.sh   # Cross-compile Go sidecar
├── PRD.md                 # Product Requirements (Indonesian)
├── architecture.md        # System architecture + diagrams
├── api_contract.md        # API contract documentation
└── database_schema.sql    # Reference SQL schema
```

---

## Lisensi

Proyek internal — belum memiliki lisensi resmi.
