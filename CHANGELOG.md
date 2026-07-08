# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0] - 2026-07-08

### Added
- **Employee Management** — Full CRUD components and UI for managing employees (`frontend/src/components/employees/`)
- **CI Workflow** — GitHub Actions workflow for automated frontend linting (oxlint), backend testing, and build verification
- **Playwright E2E Testing** — Initial Playwright test configuration for end-to-end testing with example test cases

### Changed
- Project structure now includes `frontend/tests/` for E2E test files

---

## [0.2.0] - 2026-06-22

### Added
- **Modular Documentation** — Created `docs/` folder with quickstart guide, desktop architecture docs, and AI constraints documentation
- **Tauri Desktop App** — Full integration with Tauri v2, Rust sidecar manager, and Go sidecar binary
- **Calendar Grid Component** — Shift visualization with interactive calendar UI
- **SQLite Support** — Local development without external database (`DB_DRIVER=sqlite`)
- **Humanize Library** — Indonesian-friendly time formatting

### Changed
- Updated import statements for better module organization
- Architecture links and project structure details in README

### Removed
- Outdated documentation files (replaced by modular docs)

---

## [0.1.0] - 2026-06-15

### Added
- **Initial Release** — Core autoShift application with full stack implementation
- **Backend** — Go 1.26 + Fiber v2 + GORM with PostgreSQL/SQLite support
- **Frontend** — React 19 + Vite 8 + Tailwind CSS 4 + shadcn/ui
- **AI Scheduler** — Mock generator (default) with OpenAI/Ollama compatibility
- **API Endpoints** — Full REST API for schedules, validation, publish, export, and share
- **Authentication** — JWT-based auth with tenant isolation
- **Scheduling Constraints** — 12h rest, role composition, no double-booking, leave quota, fairness metrics
- **Holiday Integration** — date.nager.at public holiday API
- **MIT License**
