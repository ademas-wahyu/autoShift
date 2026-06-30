use std::net::TcpListener;
use std::sync::Mutex;
use std::time::Duration;

use tauri::Manager;
use tauri_plugin_shell::ShellExt;

struct ApiPort(u16);

fn find_free_port() -> u16 {
    let listener = TcpListener::bind("127.0.0.1:0").expect("failed to bind to port 0");
    listener.local_addr().unwrap().port()
}

async fn wait_for_health(port: u16, timeout_secs: u64) -> Result<(), String> {
    let client = reqwest::Client::new();
    let url = format!("http://127.0.0.1:{}/api/v1/health", port);
    let deadline = std::time::Instant::now() + Duration::from_secs(timeout_secs);

    while std::time::Instant::now() < deadline {
        if let Ok(resp) = client.get(&url).send().await {
            if resp.status().is_success() {
                return Ok(());
            }
        }
        tokio::time::sleep(Duration::from_millis(200)).await;
    }

    Err(format!("sidecar health check timeout after {}s", timeout_secs))
}

#[tauri::command]
fn get_api_port(state: tauri::State<'_, Mutex<ApiPort>>) -> u16 {
    state.lock().unwrap().0
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            let port = find_free_port();
            println!("[tauri] assigned free port: {}", port);

            let app_data_dir = app
                .path()
                .app_data_dir()
                .expect("failed to get app data dir");
            std::fs::create_dir_all(&app_data_dir).expect("failed to create app data dir");
            let db_path = app_data_dir.join("autoshift.db");
            let db_path_str = db_path.to_string_lossy().to_string();

            let sidecar_command = app.shell().sidecar("autoshift-server")
                .expect("failed to create sidecar command")
                .env("PORT", port.to_string())
                .env("DB_PATH", db_path_str)
                .env("DB_DRIVER", "sqlite");

            let (mut _rx, child) = sidecar_command
                .spawn()
                .expect("failed to spawn sidecar");

            println!("[tauri] sidecar spawned, PID: {:?}", child.pid());

            let app_handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                match wait_for_health(port, 10).await {
                    Ok(()) => println!("[tauri] sidecar ready on port {}", port),
                    Err(e) => {
                        eprintln!("[tauri] {}", e);
                        let _ = child.kill();
                        app_handle.exit(1);
                    }
                }
            });

            app.manage(Mutex::new(ApiPort(port)));

            Ok(())
        })
        .on_window_event(|_window, event| {
            if let tauri::WindowEvent::CloseRequested { .. } = event {
                println!("[tauri] window closing, sidecar will be killed automatically");
            }
        })
        .invoke_handler(tauri::generate_handler![get_api_port])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}


