import { useEffect, useMemo, useState } from "react";
import { Loader2, Plus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import * as api from "@/lib/api";
import { closeCaptureWindow } from "@/lib/desktop";
import { lastWorkspace, listWorkspaces } from "@/lib/workspaces";

// CaptureView is the body of the global quick-add window (#capture). It targets the
// last-opened project by default with a switcher, creates a task, and closes the window.
export function CaptureView() {
  const workspaces = useMemo(() => listWorkspaces(), []);
  const [path, setPath] = useState(() => lastWorkspace()?.path ?? workspaces[0]?.path ?? "");
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [busy, setBusy] = useState(false);

  // Esc closes the capture window from anywhere in it.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") void closeCaptureWindow();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  const submit = async () => {
    const t = title.trim();
    if (!t || !path || busy) return;
    setBusy(true);
    try {
      await api.createTask(path, { title: t, body: body.trim() || undefined });
      toast.success("Task created");
      void closeCaptureWindow();
    } catch (e) {
      setBusy(false);
      toast.error(e instanceof Error ? e.message : "Could not create task");
    }
  };

  if (workspaces.length === 0) {
    return (
      <div className="flex h-screen items-center justify-center bg-background p-6 text-center text-sm text-muted-foreground">
        Open a project in Cairn first, then quick-add will capture into it.
      </div>
    );
  }

  return (
    <div className="flex h-screen flex-col gap-2.5 bg-background p-3 text-foreground">
      <div className="flex items-center gap-2">
        <Input
          autoFocus
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && !e.shiftKey && submit()}
          placeholder="Quick add a task…"
          className="h-9"
        />
        {workspaces.length > 1 && (
          <Select value={path} onValueChange={setPath}>
            <SelectTrigger className="h-9 w-40 shrink-0">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {workspaces.map((w) => (
                <SelectItem key={w.slug} value={w.path}>
                  {w.slug}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </div>
      <Textarea
        value={body}
        onChange={(e) => setBody(e.target.value)}
        onKeyDown={(e) => (e.key === "Enter" && (e.metaKey || e.ctrlKey) ? submit() : undefined)}
        placeholder="Details (optional) — ⌘↵ to add"
        className="min-h-0 flex-1 resize-none text-sm"
      />
      <div className="flex items-center justify-end gap-2">
        <Button variant="ghost" size="sm" onClick={() => void closeCaptureWindow()}>
          Cancel
        </Button>
        <Button size="sm" onClick={submit} disabled={!title.trim() || busy}>
          {busy ? <Loader2 className="animate-spin" /> : <Plus />}
          Add task
        </Button>
      </div>
    </div>
  );
}
