import { useEffect, useState } from "react";
import { ArrowRight, Clock, FolderOpen, FolderSearch, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import * as api from "@/lib/api";
import { isTauri, pickFolder } from "@/lib/tauri";

const RECENT_KEY = "cairn-recent-folders";

function recentFolders(): string[] {
  try {
    return JSON.parse(localStorage.getItem(RECENT_KEY) || "[]");
  } catch {
    return [];
  }
}

function remember(path: string) {
  const list = [path, ...recentFolders().filter((p) => p !== path)].slice(0, 5);
  localStorage.setItem(RECENT_KEY, JSON.stringify(list));
}

export function OpenProject({
  onOpen,
  notice,
}: {
  onOpen: (path: string) => void;
  notice?: string;
}) {
  const [path, setPath] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const recents = recentFolders();

  useEffect(() => {
    // Prefill with the server's launch folder so the common case is one click.
    api.getStatus("").then((s) => setPath((p) => p || s.root)).catch(() => {});
  }, []);

  async function open(target: string) {
    target = target.trim();
    if (!target) return;
    setBusy(true);
    setError(null);
    try {
      await api.getStatus(target); // validates the folder exists
      remember(target);
      onOpen(target);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      setBusy(false);
    }
  }

  // Native OS folder picker, desktop only. Selecting a folder opens it immediately.
  async function browse() {
    const picked = await pickFolder();
    if (picked) {
      setPath(picked);
      open(picked);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-6 text-foreground">
      <div className="w-full max-w-md">
        <div className="mb-6 flex items-center gap-3">
          <span className="grid size-9 place-items-center rounded-lg bg-primary text-primary-foreground">
            <FolderOpen className="size-5" />
          </span>
          <div>
            <h1 className="text-lg font-semibold tracking-tight">Open a project</h1>
            <p className="text-sm text-muted-foreground">Point cairn at a project folder.</p>
          </div>
        </div>

        {notice && (
          <div className="mb-3 rounded-lg border border-brand/30 bg-brand/5 px-3 py-2 text-sm text-muted-foreground">
            {notice}
          </div>
        )}

        <div className="rounded-xl border bg-card p-5 text-card-foreground shadow-xs">
          <label htmlFor="path" className="text-sm font-medium">
            Project folder
          </label>
          <div className="mt-1.5 flex gap-2">
            <Input
              id="path"
              autoFocus
              value={path}
              onChange={(e) => setPath(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && open(path)}
              placeholder="/path/to/project"
              className="font-mono text-sm"
            />
            <Button onClick={() => open(path)} disabled={!path.trim() || busy}>
              {busy ? <Loader2 className="animate-spin" /> : <ArrowRight />}
              Open
            </Button>
          </div>
          {isTauri() && (
            <Button variant="outline" onClick={browse} disabled={busy} className="mt-2 w-full">
              <FolderSearch className="size-4" />
              Choose folder…
            </Button>
          )}
          {error && <p className="mt-2 text-sm text-destructive">{error}</p>}

          {recents.length > 0 && (
            <div className="mt-5">
              <p className="mb-1.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Recent
              </p>
              <div className="-mx-1">
                {recents.map((r) => (
                  <button
                    key={r}
                    onClick={() => open(r)}
                    className="flex w-full items-center gap-2 truncate rounded-md px-1 py-1.5 text-left text-sm hover:bg-muted"
                  >
                    <Clock className="size-3.5 shrink-0 text-muted-foreground" />
                    <span className="truncate font-mono text-xs">{r}</span>
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
