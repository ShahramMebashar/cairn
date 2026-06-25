// Desktop (Tauri) integration, all lazily imported and isTauri()-guarded so the plain
// browser build never depends on @tauri-apps/* at load time and degrades to no-ops.
import { isTauri } from "@/lib/tauri";
import { registerWorkspace } from "@/lib/workspaces";

export { isTauri };

// --- Navigation / deep links --------------------------------------------------------

// navigateToTask registers the workspace (so the slug resolves) and routes to the task.
export function navigateToTask(path: string, id: string): void {
  window.location.hash = `#/${registerWorkspace(path)}/task/${id}`;
}

// openDeepLink parses a cairn:// URL and navigates: cairn://task/<id>?repo=<path> or
// cairn://open?repo=<path>.
export function openDeepLink(urlStr: string): void {
  let u: URL;
  try {
    u = new URL(urlStr);
  } catch {
    return;
  }
  if (u.protocol !== "cairn:") return;
  const repo = u.searchParams.get("repo");
  if (!repo) return;
  if (u.hostname === "task") {
    const id = u.pathname.replace(/^\//, "");
    if (id) navigateToTask(repo, id);
  } else if (u.hostname === "open") {
    window.location.hash = `#/${registerWorkspace(repo)}/all`;
  }
}

// onDeepLink subscribes to cairn:// opens forwarded from the Rust shell.
export async function onDeepLink(handler: (url: string) => void): Promise<() => void> {
  if (!isTauri()) return () => {};
  const { listen } = await import("@tauri-apps/api/event");
  const un = await listen<string>("deep-link", (e) => handler(e.payload));
  return () => un();
}

// --- Live tray menu -----------------------------------------------------------------

export type TrayItem = { id: string; label: string; checked?: boolean; enabled?: boolean };
export type TrayMenuModel = { tooltip: string; title: string; sections: TrayItem[][] };

// updateTray pushes a full menu model to the native tray (Rust rebuilds it). Clicks come
// back via onTrayEvent. No-op outside the desktop app.
export async function updateTray(menu: TrayMenuModel): Promise<void> {
  if (!isTauri()) return;
  try {
    const { invoke } = await import("@tauri-apps/api/core");
    await invoke("update_tray", { menu });
  } catch {
    // best-effort; tray may not be ready yet
  }
}

// onTrayEvent subscribes to tray menu-item clicks; the payload is the item id. Returns an
// unsubscribe fn.
export async function onTrayEvent(handler: (id: string) => void): Promise<() => void> {
  if (!isTauri()) return () => {};
  const { listen } = await import("@tauri-apps/api/event");
  const un = await listen<string>("tray:menu", (e) => handler(e.payload));
  return () => un();
}

// --- Do Not Disturb -----------------------------------------------------------------

const DND_KEY = "cairn-dnd";

export function dndEnabled(): boolean {
  return localStorage.getItem(DND_KEY) === "on";
}

export function setDnd(on: boolean): void {
  localStorage.setItem(DND_KEY, on ? "on" : "off");
}

// --- OS notifications ---------------------------------------------------------------

const OS_NOTIF_KEY = "cairn-os-notifs";

export function osNotifEnabled(): boolean {
  return localStorage.getItem(OS_NOTIF_KEY) !== "off";
}

export function setOsNotifEnabled(on: boolean): void {
  localStorage.setItem(OS_NOTIF_KEY, on ? "on" : "off");
}

// Clicking an OS notification navigates here (the most recent task we alerted about). For
// the common single-alert case this is exactly right; cross-platform click delivery varies,
// so it degrades to "notification just shows".
let lastNotifTarget: { path: string; id: string } | null = null;
let notifClicksWired = false;

async function wireNotificationClicks(): Promise<void> {
  if (notifClicksWired || !isTauri()) return;
  notifClicksWired = true;
  try {
    const n = await import("@tauri-apps/plugin-notification");
    await n.onAction(() => {
      if (lastNotifTarget) navigateToTask(lastNotifTarget.path, lastNotifTarget.id);
    });
  } catch {
    // click events unsupported on this platform — the notification still shows
  }
}

// notify shows an OS notification when running in the desktop app and the user hasn't turned
// them off. Permission is requested lazily on first use. `target` makes the alert clickable.
export async function notify(
  title: string,
  body: string,
  target?: { path: string; id: string },
): Promise<void> {
  if (!isTauri() || !osNotifEnabled() || dndEnabled()) return;
  try {
    const n = await import("@tauri-apps/plugin-notification");
    let granted = await n.isPermissionGranted();
    if (!granted) granted = (await n.requestPermission()) === "granted";
    if (!granted) return;
    if (target) {
      lastNotifTarget = target;
      void wireNotificationClicks();
    }
    n.sendNotification({ title, body });
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
