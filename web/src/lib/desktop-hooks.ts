// React hooks bridging the desktop shell to the UI. All no-op in the browser (the
// underlying calls in lib/desktop.ts are isTauri()-guarded).
import { useCallback, useEffect, useRef } from "react";
import { toast } from "sonner";
import { useTasks } from "@/lib/queries";
import { checkForUpdate, onDesktopEvent, setTrayBadge, type DesktopEvent } from "@/lib/desktop";

// useTrayBadge keeps the tray tooltip / macOS menubar badge in sync with the
// awaiting-review count for the open workspace.
export function useTrayBadge(path: string) {
  const { data: tasks } = useTasks(path);
  useEffect(() => {
    if (!tasks) return;
    const count = tasks.filter((t) => t.executionState === "awaiting_review").length;
    void setTrayBadge(count);
  }, [tasks]);
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
