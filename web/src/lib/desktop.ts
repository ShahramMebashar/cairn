// Desktop (Tauri) integration, all lazily imported and isTauri()-guarded so the plain
// browser build never depends on @tauri-apps/* at load time and degrades to no-ops.
import { isTauri } from "@/lib/tauri";

export { isTauri };

// --- Tray badge (awaiting-review count) ---------------------------------------------

export async function setTrayBadge(count: number): Promise<void> {
  if (!isTauri()) return;
  try {
    const { invoke } = await import("@tauri-apps/api/core");
    await invoke("set_tray_badge", { count });
  } catch {
    // best-effort; tray may not be ready yet
  }
}

// --- OS notifications ---------------------------------------------------------------

const OS_NOTIF_KEY = "cairn-os-notifs";

export function osNotifEnabled(): boolean {
  return localStorage.getItem(OS_NOTIF_KEY) !== "off";
}

export function setOsNotifEnabled(on: boolean): void {
  localStorage.setItem(OS_NOTIF_KEY, on ? "on" : "off");
}

// notify shows an OS notification when running in the desktop app and the user hasn't
// turned them off. Permission is requested lazily on first use.
export async function notify(title: string, body: string): Promise<void> {
  if (!isTauri() || !osNotifEnabled()) return;
  try {
    const n = await import("@tauri-apps/plugin-notification");
    let granted = await n.isPermissionGranted();
    if (!granted) granted = (await n.requestPermission()) === "granted";
    if (granted) n.sendNotification({ title, body });
  } catch {
    // notifications are best-effort
  }
}

// --- Autostart (launch at login) ----------------------------------------------------

export async function autostartEnabled(): Promise<boolean> {
  if (!isTauri()) return false;
  try {
    const a = await import("@tauri-apps/plugin-autostart");
    return await a.isEnabled();
  } catch {
    return false;
  }
}

export async function setAutostart(on: boolean): Promise<void> {
  if (!isTauri()) return;
  const a = await import("@tauri-apps/plugin-autostart");
  if (on) await a.enable();
  else await a.disable();
}

// --- Auto-updater -------------------------------------------------------------------

export type PendingUpdate = { version: string; install: () => Promise<void> };

// checkForUpdate returns the available update (with an install+relaunch action) or null.
export async function checkForUpdate(): Promise<PendingUpdate | null> {
  if (!isTauri()) return null;
  const { check } = await import("@tauri-apps/plugin-updater");
  const update = await check();
  if (!update) return null;
  return {
    version: update.version,
    install: async () => {
      await update.downloadAndInstall();
      const { relaunch } = await import("@tauri-apps/plugin-process");
      await relaunch();
    },
  };
}

// --- Menu / tray events -------------------------------------------------------------

export type DesktopEvent =
  | "menu:new_task"
  | "menu:open_folder"
  | "menu:board"
  | "menu:graph"
  | "menu:settings"
  | "menu:check_updates";

const DESKTOP_EVENTS: DesktopEvent[] = [
  "menu:new_task",
  "menu:open_folder",
  "menu:board",
  "menu:graph",
  "menu:settings",
  "menu:check_updates",
];

// onDesktopEvent subscribes to native menu/tray events; returns an unsubscribe fn.
export async function onDesktopEvent(handler: (e: DesktopEvent) => void): Promise<() => void> {
  if (!isTauri()) return () => {};
  const { listen } = await import("@tauri-apps/api/event");
  const unlistens = await Promise.all(
    DESKTOP_EVENTS.map((e) => listen(e, () => handler(e))),
  );
  return () => unlistens.forEach((u) => u());
}

// --- Capture window -----------------------------------------------------------------

export async function closeCaptureWindow(): Promise<void> {
  if (!isTauri()) return;
  try {
    const { getCurrentWindow } = await import("@tauri-apps/api/window");
    await getCurrentWindow().close();
  } catch {
    // ignore
  }
}
