# Quick Start

## Prasyarat

- **Go** 1.26+
- **Node.js** 20+
- **npm** 10+
- **Rust** 1.80+ (untuk desktop)
- **Tauri CLI**: `cargo install tauri-cli --version "^2"`

## Backend (SQLite — tanpa DB eksternal)

```bash
cp backend/.env.example backend/.env
# Edit DB_DRIVER=sqlite (atau set via env)
cd backend && go run main.go
```

## Frontend (Web Dev)

```bash
cd frontend && npm install && npm run dev
```

## Desktop (Tauri Dev)

```bash
# 1. Build Go sidecar
bash scripts/build-sidecar.sh

# 2. Jalankan Tauri dev (otomatis build frontend + start sidecar)
PKG_CONFIG_PATH=/usr/lib/x86_64-linux-gnu/pkgconfig:/usr/share/pkgconfig \
  cargo tauri dev
```

## Production Build

```bash
# Build Go sidecar untuk semua target
bash scripts/build-sidecar.sh

# Build Tauri app (output di src-tauri/target/release/bundle/)
cargo tauri build
```

Backend berjalan di `http://localhost:8080`, frontend di `http://localhost:1420` (web) atau dalam Tauri window (desktop).

## Environment Variables

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

- [Deskripsi constraints & alur kerja](constraints.md)
- [Detail desktop app & sidecar](desktop.md)
