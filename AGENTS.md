# autoShift — AGENTS.md

## Project

AI-powered employee shift scheduler. Multi-tenant, human-in-the-loop.

- **Desktop**: `src-tauri/` — Tauri v2 + Go sidecar. Rust manages sidecar lifecycle
- **Frontend**: `frontend/` — React 19 + Vite 8 + Tailwind CSS 4 + shadcn/ui (new-york)
- **Backend**: `backend/` — Go 1.26 + Fiber v2 + GORM (PostgreSQL, SQLite, MySQL, SQL Server)
- **Database**: Choose via `DB_DRIVER` env var. Default PostgreSQL. SQLite for local dev (`DB_DRIVER=sqlite`). AutoMigrate on startup, seeds default data on empty DB
- **AI**: Mock generator by default (`AI_PROVIDER=mock`), optional OpenAI / self-hosted (`AI_PROVIDER=openai`). Endpoint configurable via `AI_API_URL` — bisa pakai Ollama, vLLM, dll

## Quick start

```bash
# Frontend
cd frontend && npm install && npm run dev

# Backend (SQLite — no external DB needed)
cp backend/.env.example backend/.env
# edit DB_DRIVER=sqLite (or set via env)
cd backend && go run main.go
```

## Commands

| Scope | Command | Notes |
|-------|---------|-------|
| Frontend | `npm run dev` | Vite dev server |
| Frontend | `npm run build` | `tsc -b && vite build` |
| Frontend | `npm run lint` | oxlint |
| Backend | `go run main.go` | runs from `backend/` |
| Desktop | `bash scripts/build-sidecar.sh` | Build Go binary for current platform |
| Desktop | `PKG_CONFIG_PATH=... cargo tauri dev` | Tauri dev mode (requires PKG_CONFIG_PATH) |

## Architecture

- **Backend entrypoint**: `backend/main.go` — creates DB connection, migrates, seeds, registers routes, starts Fiber
- **Frontend entrypoint**: `frontend/src/main.tsx` → `App.tsx`
- **Path alias**: `@/` → `frontend/src/` (configured in vite.config.ts + tsconfig paths)
- **API base**: `/api/v1`, JWT auth on `/schedules/*` routes (public: /health, /login, /holidays)
- **Desktop**: Tauri v2 + Go sidecar binary. Rust cari port kosong, spawn Go binary, poll health check. Frontend auto-detect via `__TAURI_INTERNALS__`
- **No tests** exist yet in either package
- **No CI/CD** configured

## Key conventions

- Frontend lint: `oxlint` (not eslint). Config in `frontend/.oxlintrc.json`
- TypeScript: `verbatimModuleSyntax` — use `import type` for type-only imports
- TypeScript target: `es2023`, module: `esnext` (bundler mode)
- TS config uses project references: `tsconfig.json` → `tsconfig.app.json` + `tsconfig.node.json`
- Backend env loaded by godotenv from `backend/.env`
- CORS wide open (`*`) for development
- `joho/godotenv` loads env vars (can also use OS env vars)
- Backend hardcodes `BatchSize: 20`, `MinRestHours: 12`, `MaxRetries: 3` in `config.Load()`
- Set `DB_DRIVER=sqlite` for local dev (no external DB), `postgres`/`mysql`/`sqlserver` otherwise

## Documents

- `PRD.md` — product requirements (Indonesian)
- `architecture.md` — system architecture with Mermaid diagrams
- `api_contract.md` — full API contract with request/response schemas
- `database_schema.sql` — reference SQL schema (GORM AutoMigrate is source of truth)
- `backend/README.md` — detailed backend API docs (Indonesian)

## Constraints

- **12h minimum rest** between consecutive shifts
- **Role composition** per shift (e.g. ≥1 Supervisor per shift)
- **No double-booking**: one employee = one shift per day
- **Leave quota** not exceeded per employee/month
- Weekend/holiday distribution fairness tracked via std dev metrics
