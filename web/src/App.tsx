import { useEffect, useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import { Button } from "@/components/ui/button";
import { AppSidebar, type Filter } from "@/components/AppSidebar";
import { CreateTaskDialog } from "@/components/CreateTaskDialog";
import { OpenProject } from "@/pages/OpenProject";
import { InitProject } from "@/pages/InitProject";
import { Board } from "@/pages/Board";
import { BoardView } from "@/pages/BoardView";
import { TaskDetail } from "@/pages/TaskDetail";
import { Graph } from "@/pages/Graph";
import { Connect } from "@/pages/Connect";
import { CommandPalette } from "@/components/CommandPalette";
import { CaptureView } from "@/components/CaptureView";
import { SettingsDialog } from "@/components/SettingsDialog";
import { useStatus, useTaskEvents } from "@/lib/queries";
import { useDeepLinks, useDesktopMenu, useTrayMenu, useUpdater } from "@/lib/desktop-hooks";
import { isTauri, pickFolder } from "@/lib/tauri";
import {
  forget,
  lastWorkspace,
  registerWorkspace,
  resolveSlug,
} from "@/lib/workspaces";
import type { Status } from "@/lib/api";

const queryClient = new QueryClient({
  defaultOptions: { queries: { refetchOnWindowFocus: false, retry: false } },
});

// The global quick-add window loads the SPA at #capture and renders only the capture UI.
function isCaptureRoute(): boolean {
  return window.location.hash.replace(/^#\/?/, "") === "capture";
}

// macOS uses a frameless window (titleBarStyle: Overlay) — the traffic lights float over a
// slim draggable strip, Linear-style. Other platforms keep their native title bar.
function isMacDesktop(): boolean {
  return isTauri() && typeof navigator !== "undefined" && /Mac/i.test(navigator.userAgent);
}

// TitleBar is the draggable region that replaces the native title bar on macOS. The OS
// renders the traffic lights over its left; the rest drags the window.
function TitleBar() {
  if (!isMacDesktop()) return null;
  return <div data-tauri-drag-region className="h-7 shrink-0 bg-app" />;
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider delayDuration={200}>
        {isCaptureRoute() ? (
          <CaptureView />
        ) : (
          <div className="flex h-screen flex-col bg-app">
            <TitleBar />
            <div className="min-h-0 flex-1">
              <Flow />
            </div>
          </div>
        )}
        <Toaster richColors />
      </TooltipProvider>
    </QueryClientProvider>
  );
}

// --- routing: #/<workspace-slug>/<view> ---
//   #/cairn            -> all tasks      #/cairn/active|stalled|review|backlog|ready
//   #/cairn/task/ACME-004
type View =
  | { kind: "list"; filter: Filter }
  | { kind: "task"; id: string }
  | { kind: "graph" }
  | { kind: "board" }
  | { kind: "connect" };
type Route = { slug: string | null; view: View };

const FILTERS: Filter[] = ["all", "active", "stalled", "review", "backlog", "ready"];

function parseHash(): Route {
  const parts = window.location.hash
    .replace(/^#\/?/, "")
    .split("/")
    .filter(Boolean)
    .map(decodeURIComponent);
  const slug = parts[0] ?? null;
  const rest = parts.slice(1);
  let view: View = { kind: "list", filter: "all" };
  if (rest[0] === "task" && rest[1]) view = { kind: "task", id: rest[1] };
  else if (rest[0] === "graph") view = { kind: "graph" };
  else if (rest[0] === "board") view = { kind: "board" };
  else if (rest[0] === "connect") view = { kind: "connect" };
  else if (FILTERS.includes(rest[0] as Filter)) view = { kind: "list", filter: rest[0] as Filter };
  return { slug, view };
}

function hashFor(slug: string, view: View): string {
  if (view.kind === "task") return `#/${slug}/task/${encodeURIComponent(view.id)}`;
  if (view.kind === "graph") return `#/${slug}/graph`;
  if (view.kind === "board") return `#/${slug}/board`;
  if (view.kind === "connect") return `#/${slug}/connect`;
  return `#/${slug}/${view.filter}`;
}

function useRoute(): Route {
  const [route, setRoute] = useState<Route>(parseHash);
  useEffect(() => {
    const onHash = () => setRoute(parseHash());
    window.addEventListener("hashchange", onHash);
    return () => window.removeEventListener("hashchange", onHash);
  }, []);
  return route;
}

function Flow() {
  const route = useRoute();
  useDeepLinks(); // route cairn:// opens (desktop only; no-op in the browser)

  const open = (path: string) => {
    window.location.hash = hashFor(registerWorkspace(path), { kind: "list", filter: "all" });
  };

  // No slug in the URL: jump to the last-opened workspace if we have one.
  useEffect(() => {
    if (route.slug) return;
    const last = lastWorkspace();
    if (last) window.location.hash = hashFor(last.slug, { kind: "list", filter: "all" });
  }, [route.slug]);

  if (!route.slug) return <OpenProject onOpen={open} />;

  const path = resolveSlug(route.slug);
  if (!path)
    return (
      <OpenProject
        onOpen={open}
        notice={`Unknown workspace "${route.slug}" — open its folder to continue.`}
      />
    );

  const navigate = (view: View) => {
    window.location.hash = hashFor(route.slug!, view);
  };
  const changeFolder = () => {
    forget(path);
    window.location.hash = "#/";
  };

  return <Project path={path} view={route.view} navigate={navigate} onChangeFolder={changeFolder} />;
}

function Project({
  path,
  view,
  navigate,
  onChangeFolder,
}: {
  path: string;
  view: View;
  navigate: (v: View) => void;
  onChangeFolder: () => void;
}) {
  const { data: status, isLoading, error } = useStatus(path);

  if (isLoading)
    return (
      <Centered>
        <Loader2 className="size-5 animate-spin text-muted-foreground" />
      </Centered>
    );
  if (error || !status)
    return (
      <Centered>
        <div className="text-center">
          <p className="text-sm text-destructive">
            {error instanceof Error ? error.message : "Could not open folder"}
          </p>
          <Button variant="outline" size="sm" className="mt-4" onClick={onChangeFolder}>
            Choose another folder
          </Button>
        </div>
      </Centered>
    );
  if (!status.initialized)
    return <InitProject path={path} status={status} onChangeFolder={onChangeFolder} />;
  return (
    <Workspace
      path={path}
      status={status}
      view={view}
      navigate={navigate}
      onChangeFolder={onChangeFolder}
    />
  );
}

function Workspace({
  path,
  status,
  view,
  navigate,
  onChangeFolder,
}: {
  path: string;
  status: Status;
  view: View;
  navigate: (v: View) => void;
  onChangeFolder: () => void;
}) {
  const [creating, setCreating] = useState(false);
  const [createParent, setCreateParent] = useState<string | undefined>(undefined);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const newTask = () => {
    setCreateParent(undefined);
    setCreating(true);
  };
  const addSubtask = (parentId: string) => {
    setCreateParent(parentId);
    setCreating(true);
  };

  useTaskEvents(path); // live board/task updates from any actor via SSE

  // Desktop integration (no-ops in the browser): live tray menu, update checks, native menu.
  const checkUpdates = useUpdater();
  const openFolder = async () => {
    const picked = await pickFolder();
    if (picked) window.location.hash = `#/${registerWorkspace(picked)}/all`;
  };
  useTrayMenu(path, {
    openTask: (id) => navigate({ kind: "task", id }),
    openFilter: (f) => navigate({ kind: "list", filter: f as Filter }),
    switchProject: (slug) => {
      window.location.hash = `#/${slug}/all`;
    },
    newTask,
    openSettings: () => setSettingsOpen(true),
  });
  useDesktopMenu({
    "menu:new_task": newTask,
    "menu:open_folder": () => void openFolder(),
    "menu:board": () => navigate({ kind: "board" }),
    "menu:graph": () => navigate({ kind: "graph" }),
    "menu:settings": () => setSettingsOpen(true),
    "menu:check_updates": () => void checkUpdates(true),
  });

  return (
    <div className="flex h-full overflow-hidden bg-app text-foreground">
      <AppSidebar
        path={path}
        status={status}
        active={view.kind === "list" ? view.filter : null}
        graphActive={view.kind === "graph"}
        boardActive={view.kind === "board"}
        connectActive={view.kind === "connect"}
        onFilter={(f) => navigate({ kind: "list", filter: f })}
        onGraph={() => navigate({ kind: "graph" })}
        onBoard={() => navigate({ kind: "board" })}
        onConnect={() => navigate({ kind: "connect" })}
        onChangeFolder={onChangeFolder}
        onNewTask={newTask}
        onOpenTask={(id) => navigate({ kind: "task", id })}
        onOpenSettings={() => setSettingsOpen(true)}
      />
      <main className="min-w-0 flex-1 p-2 pl-0">
        <div className="flex h-full flex-col overflow-hidden rounded-xl border bg-panel shadow-xs">
          {view.kind === "list" ? (
            <Board
              path={path}
              status={status}
              filter={view.filter}
              onOpenTask={(id) => navigate({ kind: "task", id })}
              onNewTask={newTask}
              onPickFilter={(f) => navigate({ kind: "list", filter: f })}
            />
          ) : view.kind === "graph" ? (
            <Graph
              path={path}
              status={status}
              onOpenTask={(id) => navigate({ kind: "task", id })}
              onBack={() => navigate({ kind: "list", filter: "all" })}
            />
          ) : view.kind === "board" ? (
            <BoardView
              path={path}
              status={status}
              onOpenTask={(id) => navigate({ kind: "task", id })}
              onNewTask={newTask}
            />
          ) : view.kind === "connect" ? (
            <Connect path={path} status={status} />
          ) : (
            <TaskDetail
              path={path}
              id={view.id}
              status={status}
              onBack={() => navigate({ kind: "list", filter: "all" })}
              onOpenTask={(id) => navigate({ kind: "task", id })}
              onAddSubtask={addSubtask}
            />
          )}
        </div>
      </main>
      <CreateTaskDialog
        path={path}
        open={creating}
        onOpenChange={setCreating}
        defaultParent={createParent}
      />
      <CommandPalette
        path={path}
        status={status}
        onView={(f) => navigate({ kind: "list", filter: f })}
        onOpenTask={(id) => navigate({ kind: "task", id })}
        onNewTask={newTask}
        onChangeFolder={onChangeFolder}
        onGraph={() => navigate({ kind: "graph" })}
        onBoard={() => navigate({ kind: "board" })}
      />
      <SettingsDialog
        open={settingsOpen}
        onOpenChange={setSettingsOpen}
        onCheckUpdates={() => void checkUpdates(true)}
      />
    </div>
  );
}

function Centered({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-full items-center justify-center bg-background p-6 text-foreground">
      {children}
    </div>
  );
}
