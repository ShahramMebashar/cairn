// React hooks bridging the desktop shell to the UI. All no-op in the browser (the
// underlying calls in lib/desktop.ts are isTauri()-guarded).
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { useSessions, useTasks } from "@/lib/queries";
import { listWorkspaces } from "@/lib/workspaces";
import { timeAgo } from "@/lib/utils";
import {
  checkForUpdate,
  dndEnabled,
  onDeepLink,
  onDesktopEvent,
  onTrayEvent,
  openDeepLink,
  setDnd,
  updateTray,
  type DesktopEvent,
  type TrayItem,
  type TrayMenuModel,
} from "@/lib/desktop";

// useDeepLinks routes cairn:// opens (from the OS) to the right task/project.
export function useDeepLinks() {
  useEffect(() => {
    let off = () => {};
    void onDeepLink(openDeepLink).then((fn) => {
      off = fn;
    });
    return () => off();
  }, []);
}

export type TrayHandlers = {
  openTask: (id: string) => void;
  openFilter: (filter: string) => void;
  switchProject: (slug: string) => void;
  newTask: () => void;
  openSettings: () => void;
};

const trunc = (s: string, n = 42) => (s.length > n ? `${s.slice(0, n - 1)}…` : s);
const shortActor = (a?: string) => (a || "").replace(/^(agent|human):/, "");

// useTrayMenu pushes a live menu model (status counts + awaiting-review tasks + active agent
// sessions + actions) to the native tray, debounced/diffed, and dispatches tray clicks. The
// hidden window keeps SSE alive, so this stays current even when the app isn't visible.
export function useTrayMenu(path: string, handlers: TrayHandlers) {
  const { data: tasks } = useTasks(path);
  const { data: sessions } = useSessions(path);
  const [dnd, setDndState] = useState(dndEnabled());
  const lastJson = useRef("");
  const hRef = useRef(handlers);
  hRef.current = handlers;

  const workspaces = useMemo(() => listWorkspaces(), [path]);
  const currentSlug = useMemo(() => workspaces.find((w) => w.path === path)?.slug, [workspaces, path]);

  useEffect(() => {
    if (!tasks) return;
    const review = tasks.filter((t) => t.executionState === "awaiting_review");
    const active = tasks.filter((t) => t.executionState === "active");
    const stalled = tasks.filter((t) => t.executionState === "stalled");
    const ready = tasks.filter((t) => t.ready && !t.executionState).length;

    const counts: TrayItem[] = [];
    if (review.length) counts.push({ id: "filter:review", label: `● ${review.length} Awaiting review` });
    if (active.length) counts.push({ id: "filter:active", label: `▶ ${active.length} Active` });
    if (stalled.length) counts.push({ id: "filter:stalled", label: `■ ${stalled.length} Stalled` });
    if (ready) counts.push({ id: "filter:ready", label: `✦ ${ready} Ready` });
    if (!counts.length) counts.push({ id: "noop", label: "No active work", enabled: false });

    const sections: TrayItem[][] = [counts];

    if (review.length) {
      const sec: TrayItem[] = [{ id: "hdr:review", label: "Awaiting review", enabled: false }];
      review.slice(0, 5).forEach((t) => sec.push({ id: `task:${t.id}`, label: trunc(`${t.id}  ${t.title}`) }));
      sections.push(sec);
    }

    const live = (sessions ?? []).filter((s) => s.health === "active" || s.health === "stalled");
    if (live.length) {
      const sec: TrayItem[] = [{ id: "hdr:agents", label: "Active agents", enabled: false }];
      live.slice(0, 5).forEach((s) => {
        const prog = s.live?.progress || (s.health === "stalled" ? "stalled" : "working");
        const when = s.live?.heartbeatAt ? ` · ${timeAgo(s.live.heartbeatAt)}` : "";
        const flag = s.health === "stalled" ? "⚠ " : "";
        sec.push({ id: `task:${s.task}`, label: trunc(`${flag}${shortActor(s.actor)} · ${prog}${when}`, 46) });
      });
      sections.push(sec);
    }

    const actions: TrayItem[] = [{ id: "new_task", label: "New task…" }];
    if (workspaces.length > 1) {
      actions.push({ id: "hdr:projects", label: "Project", enabled: false });
      workspaces.forEach((w) =>
        actions.push({ id: `project:${w.slug}`, label: w.slug, checked: w.slug === currentSlug }),
      );
    }
    actions.push({ id: "toggle:dnd", label: "Do Not Disturb", checked: dnd });
    sections.push(actions);

    sections.push([
      { id: "tray_open", label: "Open Cairn" },
      { id: "settings", label: "Settings…" },
      { id: "tray_quit", label: "Quit Cairn" },
    ]);

    const model: TrayMenuModel = {
      tooltip: review.length ? `Cairn — ${review.length} awaiting review` : "Cairn",
      title: review.length ? String(review.length) : "",
      sections,
    };

    const json = JSON.stringify(model);
    if (json === lastJson.current) return;
    const timer = setTimeout(() => {
      lastJson.current = json;
      void updateTray(model);
    }, 250);
    return () => clearTimeout(timer);
  }, [tasks, sessions, dnd, workspaces, currentSlug]);

  useEffect(() => {
    let off = () => {};
    void onTrayEvent((id) => {
      const h = hRef.current;
      if (id.startsWith("task:")) h.openTask(id.slice(5));
      else if (id.startsWith("filter:")) h.openFilter(id.slice(7));
      else if (id.startsWith("project:")) h.switchProject(id.slice(8));
      else if (id === "new_task") h.newTask();
      else if (id === "settings") h.openSettings();
      else if (id === "toggle:dnd")
        setDndState((d) => {
          const next = !d;
          setDnd(next);
          return next;
        });
    }).then((fn) => {
      off = fn;
    });
    return () => off();
  }, []);
}

// useUpdater checks for an update on mount and returns a manual checker (for the
// Settings button / menu). An available update is offered via a toast action.
export function useUpdater() {
  const run = useCallback(async (manual = false) => {
    try {
      const update = await checkForUpdate();
      if (!update) {
        if (manual) toast.success("You're on the latest version.");
        return;
      }
      toast(`Update available: v${update.version}`, {
        description: "Download and restart to install.",
        duration: 12000,
        action: { label: "Install & restart", onClick: () => void update.install() },
      });
    } catch {
      if (manual) toast.error("Couldn't check for updates.");
    }
  }, []);

  useEffect(() => {
    void run(false);
  }, [run]);

  return run;
}

type MenuHandlers = Partial<Record<DesktopEvent, () => void>>;

// useDesktopMenu dispatches native menu/tray events to the matching UI action.
export function useDesktopMenu(handlers: MenuHandlers) {
  const ref = useRef(handlers);
  ref.current = handlers;
  useEffect(() => {
    let off = () => {};
    void onDesktopEvent((e) => ref.current[e]?.()).then((fn) => {
      off = fn;
    });
    return () => off();
  }, []);
}
