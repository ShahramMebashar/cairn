use std::io::{Read, Write};
use std::net::{SocketAddr, TcpStream};
use std::sync::Mutex;
use std::time::Duration;

use tauri::menu::{CheckMenuItemBuilder, MenuBuilder, MenuItemBuilder, SubmenuBuilder};
use tauri::tray::TrayIconBuilder;
use tauri::{Emitter, Manager, RunEvent, WebviewUrl, WebviewWindowBuilder};

/// Preferred listen address for the bundled server. The Go side falls back to the
/// next free port if this is taken, and reports the port it actually bound.
const PREFERRED_ADDR: &str = "127.0.0.1:7777";

/// Default global shortcut that pops the quick-capture window.
const CAPTURE_SHORTCUT: &str = "CmdOrCtrl+Shift+K";

/// Full-page message shown if the bundled cairn server never comes up. Self-contained
/// so it works regardless of what the hidden webview had loaded.
const STARTUP_ERROR_JS: &str = "document.open();document.write('<!doctype html><meta charset=utf-8><body style=\"margin:0;height:100vh;display:grid;place-items:center;font-family:ui-sans-serif,system-ui,sans-serif;background:#0b0d12;color:#cbd5e1\"><div style=\"text-align:center\"><div style=\"font-weight:600;letter-spacing:.02em\">Cairn</div><div style=\"font-size:12px;opacity:.6;margin-top:6px\">Could not start the local server. Quit and reopen Cairn.</div></div></body>');document.close();";

/// Holds the spawned sidecar so we can kill it on a real quit. The Go side also watches
/// stdin (via --parent-watch) and dies on EOF, covering crashes/force-quits.
struct Sidecar(Mutex<Option<tauri_plugin_shell::process::CommandChild>>);

/// The URL the bundled server bound to (from its stdout handshake). Used to open the
/// capture window at the right origin. None in dev or before the server is up.
struct ServerUrl(Mutex<Option<String>>);

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let mut builder = tauri::Builder::default();

    // single-instance MUST be registered first: a second launch focuses the running
    // window instead of spawning a duplicate sidecar (which would fight for the port).
    #[cfg(desktop)]
    {
        builder = builder.plugin(tauri_plugin_single_instance::init(|app, argv, _cwd| {
            show_main(app);
            // On Windows/Linux a cairn:// open while running arrives as a CLI arg to this
            // second instance — forward it to the UI.
            for arg in &argv {
                if arg.starts_with("cairn://") {
                    let _ = app.emit("deep-link", arg.clone());
                }
            }
        }));
    }

    builder
        .plugin(tauri_plugin_window_state::Builder::default().with_denylist(&["capture"]).build())
        .plugin(tauri_plugin_global_shortcut::Builder::new()
            .with_handler(|app, _shortcut, event| {
                if event.state == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    open_capture(app);
                }
            })
            .build())
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_deep_link::init())
        .plugin(tauri_plugin_updater::Builder::new().build())
        .plugin(tauri_plugin_process::init())
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_opener::init())
        .setup(|app| {
            // Autostart (opt-in; toggled from the UI). Desktop-only.
            #[cfg(desktop)]
            {
                use tauri_plugin_autostart::MacosLauncher;
                app.handle().plugin(tauri_plugin_autostart::init(
                    MacosLauncher::LaunchAgent,
                    None,
                ))?;
            }

            #[cfg(debug_assertions)]
            app.handle().plugin(
                tauri_plugin_log::Builder::default()
                    .level(log::LevelFilter::Info)
                    .build(),
            )?;

            app.manage(ServerUrl(Mutex::new(None)));

            // Native menu (gives macOS its Edit menu → copy/paste in inputs) + tray.
            let menu = build_menu(app.handle())?;
            app.set_menu(menu)?;
            app.on_menu_event(|app, event| handle_menu(app, event.id().as_ref()));
            build_tray(app)?;

            // Global quick-capture shortcut.
            #[cfg(desktop)]
            {
                use tauri_plugin_global_shortcut::GlobalShortcutExt;
                let _ = app.global_shortcut().register(CAPTURE_SHORTCUT);
            }

            // Deep links (cairn://): forward OS opens to the UI. register_all() is best-effort
            // for Linux/Windows runtime registration; macOS uses the bundled Info.plist scheme.
            #[cfg(desktop)]
            {
                use tauri_plugin_deep_link::DeepLinkExt;
                let _ = app.deep_link().register_all();
                let dh = app.handle().clone();
                app.deep_link().on_open_url(move |event| {
                    for url in event.urls() {
                        show_main(&dh);
                        let _ = dh.emit("deep-link", url.to_string());
                    }
                });
            }

            // Close-to-tray (prod only): the X hides the window so the server + MCP stay
            // up. Quit explicitly from the tray/menu. Dev keeps normal close for fast iter.
            if !cfg!(debug_assertions) {
                if let Some(w) = app.get_webview_window("main") {
                    let wc = w.clone();
                    w.on_window_event(move |ev| {
                        if let tauri::WindowEvent::CloseRequested { api, .. } = ev {
                            api.prevent_close();
                            let _ = wc.hide();
                        }
                    });
                }
            }

            if cfg!(debug_assertions) {
                // Dev: Vite (devUrl) serves the live UI and the developer runs
                // `cairn web` separately — just reveal the window.
                if let Some(w) = app.get_webview_window("main") {
                    let _ = w.show();
                }
                return Ok(());
            }

            // Production: spawn the bundled cairn server. It binds a local port
            // (7777, or the next free one), prints `CAIRN_WEB_URL=<url>` on stdout,
            // and dies with us thanks to --parent-watch.
            use tauri_plugin_shell::ShellExt;
            let (mut rx, child) = app
                .shell()
                .sidecar("cairn")?
                .args(["web", "--addr", PREFERRED_ADDR, "--parent-watch"])
                .spawn()?;
            app.manage(Sidecar(Mutex::new(Some(child))));

            // Read the sidecar's stdout for the URL line; record it (for the capture
            // window) and, once seen, poll /healthz on a worker thread before navigating.
            let handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                use tauri_plugin_shell::process::CommandEvent;
                while let Some(ev) = rx.recv().await {
                    if let CommandEvent::Stdout(b) | CommandEvent::Stderr(b) = ev {
                        let chunk = String::from_utf8_lossy(&b);
                        for line in chunk.lines() {
                            if let Some(url) = parse_url(line) {
                                *handle.state::<ServerUrl>().0.lock().unwrap() = Some(url.clone());
                                let h = handle.clone();
                                std::thread::spawn(move || finish_startup(&h, &url));
                            }
                        }
                        log::info!("cairn: {}", chunk.trim_end());
                    }
                }
            });

            // Safety net: if no URL ever arrives, show a clear error rather than a
            // window stuck hidden (or, worse, never shown at all).
            let h2 = app.handle().clone();
            std::thread::spawn(move || {
                std::thread::sleep(Duration::from_secs(20));
                if let Some(w) = h2.get_webview_window("main") {
                    if !w.is_visible().unwrap_or(false) {
                        let _ = w.eval(STARTUP_ERROR_JS);
                        let _ = w.show();
                    }
                }
            });

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![update_tray])
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|app, event| {
            // Belt to the Go-side stdin watcher: kill the child on a real quit.
            if matches!(event, RunEvent::ExitRequested { .. } | RunEvent::Exit) {
                if let Some(state) = app.try_state::<Sidecar>() {
                    if let Some(child) = state.0.lock().unwrap().take() {
                        let _ = child.kill();
                    }
                }
            }
        });
}

/// A tray menu pushed from the frontend. Rust is a dumb renderer: the UI owns the content
/// and the click logic (clicks come back as a single `tray:menu` event carrying the item id).
#[derive(serde::Deserialize)]
struct TrayItem {
    id: String,
    label: String,
    checked: Option<bool>,
    enabled: Option<bool>,
}

#[derive(serde::Deserialize)]
struct TrayMenu {
    tooltip: String,
    title: String,
    sections: Vec<Vec<TrayItem>>,
}

/// Rebuild the tray menu from a frontend-supplied model + set the tooltip / macOS menubar
/// title (the attention badge). Sections are separated by a divider.
#[tauri::command]
fn update_tray(app: tauri::AppHandle, menu: TrayMenu) -> Result<(), String> {
    let Some(tray) = app.tray_by_id("main") else {
        return Ok(());
    };
    let mut b = MenuBuilder::new(&app);
    for (i, section) in menu.sections.iter().enumerate() {
        if i > 0 {
            b = b.separator();
        }
        for item in section {
            if let Some(checked) = item.checked {
                let ci = CheckMenuItemBuilder::with_id(&item.id, &item.label)
                    .checked(checked)
                    .build(&app)
                    .map_err(|e| e.to_string())?;
                b = b.item(&ci);
            } else {
                let mi = MenuItemBuilder::with_id(&item.id, &item.label)
                    .enabled(item.enabled.unwrap_or(true))
                    .build(&app)
                    .map_err(|e| e.to_string())?;
                b = b.item(&mi);
            }
        }
    }
    let built = b.build().map_err(|e| e.to_string())?;
    tray.set_menu(Some(built)).map_err(|e| e.to_string())?;
    let _ = tray.set_tooltip(Some(&menu.tooltip));
    #[cfg(target_os = "macos")]
    {
        let _ = tray.set_title(Some(&menu.title));
    }
    Ok(())
}

/// Build the application menu. The Edit submenu is what gives macOS webviews working
/// copy/paste/select-all in text inputs. Custom items emit `menu:*` events to the UI.
fn build_menu(app: &tauri::AppHandle) -> tauri::Result<tauri::menu::Menu<tauri::Wry>> {
    let settings = MenuItemBuilder::with_id("settings", "Settings…")
        .accelerator("CmdOrCtrl+,")
        .build(app)?;
    let check_updates = MenuItemBuilder::with_id("check_updates", "Check for Updates…").build(app)?;
    let new_task = MenuItemBuilder::with_id("new_task", "New Task")
        .accelerator("CmdOrCtrl+N")
        .build(app)?;
    let open_folder = MenuItemBuilder::with_id("open_folder", "Open Folder…")
        .accelerator("CmdOrCtrl+O")
        .build(app)?;
    let board = MenuItemBuilder::with_id("board", "Board").build(app)?;
    let graph = MenuItemBuilder::with_id("graph", "Graph").build(app)?;

    let app_menu = SubmenuBuilder::new(app, "Cairn")
        .about(None)
        .item(&check_updates)
        .separator()
        .item(&settings)
        .separator()
        .services()
        .separator()
        .hide()
        .hide_others()
        .show_all()
        .separator()
        .quit()
        .build()?;
    let file_menu = SubmenuBuilder::new(app, "File")
        .item(&new_task)
        .item(&open_folder)
        .build()?;
    let edit_menu = SubmenuBuilder::new(app, "Edit")
        .undo()
        .redo()
        .separator()
        .cut()
        .copy()
        .paste()
        .select_all()
        .build()?;
    let view_menu = SubmenuBuilder::new(app, "View")
        .item(&board)
        .item(&graph)
        .build()?;
    let window_menu = SubmenuBuilder::new(app, "Window")
        .minimize()
        .separator()
        .close_window()
        .build()?;

    MenuBuilder::new(app)
        .items(&[&app_menu, &file_menu, &edit_menu, &view_menu, &window_menu])
        .build()
}

/// Map a menu/tray item id to a UI event (or a direct action).
fn handle_menu(app: &tauri::AppHandle, id: &str) {
    match id {
        "tray_open" | "open" => show_main(app),
        "tray_quit" | "quit" => app.exit(0),
        "new_task" | "tray_new_task" => {
            show_main(app);
            let _ = app.emit("menu:new_task", ());
        }
        "settings" | "tray_settings" => {
            show_main(app);
            let _ = app.emit("menu:settings", ());
        }
        "open_folder" => {
            show_main(app);
            let _ = app.emit("menu:open_folder", ());
        }
        "board" => {
            show_main(app);
            let _ = app.emit("menu:board", ());
        }
        "graph" => {
            show_main(app);
            let _ = app.emit("menu:graph", ());
        }
        "check_updates" => {
            show_main(app);
            let _ = app.emit("menu:check_updates", ());
        }
        // Dynamic tray items (task:<id>, filter:<f>, project:<slug>, capture, toggle:dnd, …):
        // the frontend owns the action. Toggles shouldn't steal focus; everything else reveals.
        other => {
            if !other.starts_with("toggle:") {
                show_main(app);
            }
            let _ = app.emit("tray:menu", other);
        }
    }
}

/// Tray icon: left-click opens the live menu (built by the frontend via update_tray); the
/// default menu below is the pre-hydration fallback. "Open Cairn" reveals the window.
fn build_tray(app: &mut tauri::App) -> tauri::Result<()> {
    let open_i = MenuItemBuilder::with_id("tray_open", "Open Cairn").build(app)?;
    let new_i = MenuItemBuilder::with_id("tray_new_task", "New Task").build(app)?;
    let settings_i = MenuItemBuilder::with_id("tray_settings", "Settings…").build(app)?;
    let quit_i = MenuItemBuilder::with_id("tray_quit", "Quit Cairn").build(app)?;
    let tray_menu = MenuBuilder::new(app)
        .items(&[&open_i, &new_i, &settings_i])
        .separator()
        .item(&quit_i)
        .build()?;

    let icon = app
        .default_window_icon()
        .cloned()
        .expect("app must have a default icon");

    TrayIconBuilder::with_id("main")
        .icon(icon)
        .tooltip("Cairn")
        .menu(&tray_menu)
        .show_menu_on_left_click(true)
        .on_menu_event(|app, event| handle_menu(app, event.id().as_ref()))
        .build(app)?;
    Ok(())
}

/// Show + focus the main window (restoring it if minimized/hidden).
fn show_main(app: &tauri::AppHandle) {
    if let Some(w) = app.get_webview_window("main") {
        let _ = w.show();
        let _ = w.unminimize();
        let _ = w.set_focus();
    }
}

/// Open (or focus) the quick-capture window at the running server's `#capture` route.
/// Falls back to showing the main window when the server URL isn't known yet (dev).
fn open_capture(app: &tauri::AppHandle) {
    if let Some(w) = app.get_webview_window("capture") {
        let _ = w.show();
        let _ = w.set_focus();
        return;
    }
    let url = app.state::<ServerUrl>().0.lock().unwrap().clone();
    let Some(base) = url else {
        return show_main(app);
    };
    let Ok(parsed) = tauri::Url::parse(&format!("{base}/#capture")) else {
        return;
    };
    let _ = WebviewWindowBuilder::new(app, "capture", WebviewUrl::External(parsed))
        .title("Quick add — Cairn")
        .inner_size(560.0, 220.0)
        .resizable(false)
        .always_on_top(true)
        .center()
        .build();
}

/// Extract the URL from a `CAIRN_WEB_URL=<url>` stdout line.
fn parse_url(line: &str) -> Option<String> {
    line.trim()
        .strip_prefix("CAIRN_WEB_URL=")
        .map(|u| u.trim().to_string())
}

/// Poll the server's /healthz until it answers "ok" (or we give up), then navigate
/// the main window there and show it. On timeout, show the startup error instead.
fn finish_startup(handle: &tauri::AppHandle, url: &str) {
    let mut ready = false;
    for _ in 0..100 {
        if server_ready(url) {
            ready = true;
            break;
        }
        std::thread::sleep(Duration::from_millis(150));
    }
    if let Some(w) = handle.get_webview_window("main") {
        if ready {
            if let Ok(u) = tauri::Url::parse(url) {
                let _ = w.navigate(u);
            }
        } else {
            let _ = w.eval(STARTUP_ERROR_JS);
        }
        let _ = w.show();
    }
}

/// Verified readiness probe: a bare TCP connect only proves the port is busy, so we
/// send a real `GET /healthz` and require cairn's `ok` body before navigating.
fn server_ready(url: &str) -> bool {
    let authority = url.trim_start_matches("http://");
    let addr: SocketAddr = match authority.parse() {
        Ok(a) => a,
        Err(_) => return false,
    };
    let mut stream = match TcpStream::connect_timeout(&addr, Duration::from_millis(300)) {
        Ok(s) => s,
        Err(_) => return false,
    };
    let _ = stream.set_read_timeout(Some(Duration::from_millis(500)));
    let _ = stream.set_write_timeout(Some(Duration::from_millis(500)));
    // HTTP/1.0 so the server closes the connection after the body (read terminates).
    let req = format!("GET /healthz HTTP/1.0\r\nHost: {authority}\r\n\r\n");
    if stream.write_all(req.as_bytes()).is_err() {
        return false;
    }
    let mut buf = String::new();
    let _ = stream.read_to_string(&mut buf);
    buf.contains(" 200") && buf.trim_end().ends_with("ok")
}
