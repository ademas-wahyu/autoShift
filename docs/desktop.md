# Desktop App (Tauri)

## Arsitektur Runtime

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

## Flow Runtime Detail

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

## Sidecar Manager (Rust)

File: `src-tauri/src/lib.rs`

```rust
#[tauri::command]
fn get_api_port(state: tauri::State<'_, Mutex<ApiPort>>) -> u16 {
    state.lock().unwrap().0
}
```

## Go Graceful Shutdown

Backend Go menerima sinyal SIGTERM/SIGINT untuk shutdown bersih saat sidecar di-kill oleh Tauri:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
app.Shutdown()
```

## Frontend Auto-Detect

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

## Build Sidecar

Script `scripts/build-sidecar.sh` otomatis mendeteksi OS & arsitektur, lalu compile Go binary ke `src-tauri/binaries/autoshift-server-{target-triple}`.

Target triple yang didukung:
| OS | Arch | Target Triple |
|----|------|---------------|
| Linux | x86_64 | `x86_64-unknown-linux-gnu` |
| macOS | Apple Silicon | `aarch64-apple-darwin` |
| Windows | x86_64 | `x86_64-pc-windows-msvc` |

Ukuran binary ~19MB (bisa di-strip/UPX jadi ~8MB).

## Catatan Penting

| Item | Keterangan |
|------|------------|
| **Zero rewrite backend** | Semua kode Go tetap utuh, hanya tambah signal handler |
| **Frontend minimal berubah** | Hanya `api/index.ts` (dynamic base URL) dan `vite.config.ts` (port 1420) |
| **Database** | SQLite auto-create di app data directory. Migrasi & seed via GORM tetap sama |
| **Bundle size** | ~25MB final (Go binary + frontend + Tauri runtime) |
| **Cross-platform** | Go + Tauri support Linux, macOS, Windows |
