fn main() {
    // Declare our app commands so the ACL can authorize them for the remote webview origin
    // (the UI runs from the Go server's http://127.0.0.1:* URL, not tauri://). Without this,
    // invoke() of a custom command is rejected with "not allowed by ACL".
    tauri_build::try_build(
        tauri_build::Attributes::new()
            .app_manifest(tauri_build::AppManifest::new().commands(&["update_tray"])),
    )
    .expect("failed to run tauri-build");
}
