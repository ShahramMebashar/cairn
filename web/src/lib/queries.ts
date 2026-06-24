// React Query hooks over the API. Queries are keyed by project path; mutations invalidate
// the task list (and the affected task) and surface success/failure via sonner toasts —
// crucially, gate refusals (failed checks, unclosed deps) show the backend's reason.

import { useEffect } from "react";
import {
  useMutation,
  useQuery,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";
import * as api from "./api";
import type { CreateInput } from "./api";

const tasksKey = (path: string) => ["tasks", path] as const;
const taskKey = (path: string, id: string) => ["task", path, id] as const;
const runsKey = (path: string, id: string) => ["runs", path, id] as const;
const sessionsKey = (path: string, id?: string) => ["sessions", path, ...(id ? [id] : [])] as const;

// isGatedStatus reports whether entering `to` runs the task's command checks server-side
// (i.e. the move blocks while checks run). Mirrors the backend gate: closed states or the
// review state. Reads the cached workspace status so callers needn't thread it through.
export function isGatedStatus(qc: QueryClient, path: string, to: string): boolean {
  const st = qc.getQueryData<api.Status>(["status", path]);
  if (!st) return false;
  return (st.closed ?? []).includes(to) || st.review === to;
}

export function useStatus(path: string | null) {
  return useQuery({
    queryKey: ["status", path],
    queryFn: () => api.getStatus(path as string),
    enabled: path !== null,
    retry: false,
  });
}

export function useTasks(path: string) {
  return useQuery({ queryKey: tasksKey(path), queryFn: () => api.listTasks(path) });
}

export function useTask(path: string, id: string | null) {
  return useQuery({
    queryKey: taskKey(path, id ?? ""),
    queryFn: () => api.getTask(path, id as string),
    enabled: !!id,
  });
}

export function useRuns(path: string, id: string | null) {
  return useQuery({
    queryKey: runsKey(path, id ?? ""),
    queryFn: () => api.getRuns(path, id as string),
    enabled: !!id,
  });
}

export function useTaskSessions(path: string, id: string | null) {
  return useQuery({
    queryKey: sessionsKey(path, id ?? ""),
    queryFn: () => api.listTaskSessions(path, id as string),
    enabled: !!id,
  });
}

// useTaskEvents subscribes to the server's SSE stream for `path` and invalidates the
// affected React Query caches, so the board and open task reflect changes made by ANY
// actor (including MCP agents in another process), not just this UI's own mutations. One
// EventSource per active path; the browser auto-reconnects on drop.
export function useTaskEvents(path: string) {
  const qc = useQueryClient();
  useEffect(() => {
    const es = new EventSource(`/api/events?path=${encodeURIComponent(path)}`);
    es.onmessage = (e) => {
      let msg: { type: string; id?: string; session?: string };
      try {
        msg = JSON.parse(e.data);
      } catch {
        return;
      }
      // Always refresh the list (covers create/delete). For a single-task change also
      // refresh that task and its runs so check output updates live.
      qc.invalidateQueries({ queryKey: tasksKey(path) });
      if (msg.type === "task-changed" && msg.id) {
        qc.invalidateQueries({ queryKey: taskKey(path, msg.id) });
        qc.invalidateQueries({ queryKey: runsKey(path, msg.id) });
        qc.invalidateQueries({ queryKey: sessionsKey(path, msg.id) });
      }
      if (msg.type === "session-changed") {
        qc.invalidateQueries({ queryKey: sessionsKey(path) });
        // Session events carry the session id, not the task id. Refresh open task
        // projections so execution state cannot lag behind the session timeline.
        qc.invalidateQueries({ queryKey: ["task", path] });
      }
    };
    return () => es.close();
  }, [path, qc]);
}

function refresh(qc: QueryClient, path: string, id?: string) {
  qc.invalidateQueries({ queryKey: tasksKey(path) });
  if (id) qc.invalidateQueries({ queryKey: taskKey(path, id) });
}

const fail = (e: unknown) => toast.error(e instanceof Error ? e.message : String(e));

export function useInitRepo(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (prefix?: string) => api.initRepo(path, prefix),
    onSuccess: (st) => {
      qc.setQueryData(["status", path], st);
      toast.success(`Initialized ${st.prefix}`);
    },
    onError: fail,
  });
}

export function useCreateTask(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateInput) => api.createTask(path, input),
    onSuccess: (t) => {
      refresh(qc, path);
      toast.success(`Created ${t.id}`);
    },
    onError: fail,
  });
}

export function useClaim(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.claimTask(path, id),
    onSuccess: (t) => {
      refresh(qc, path, t.id);
      toast.success(`Claimed ${t.id}`);
    },
    onError: fail,
  });
}

export function useTransition(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, to }: { id: string; to: string }) => api.transitionTask(path, id, to),
    // Reflect the new status immediately for free (non-gated) moves so the board and open
    // detail update with no lag. Gated moves (review/closed) run command checks server-side and
    // may be refused, so we DON'T fake them — the UI shows a "running checks…" state until the
    // server confirms. onError rolls back, onSettled reconciles with the real document.
    onMutate: async ({ id, to }) => {
      if (isGatedStatus(qc, path, to)) return { id };
      await qc.cancelQueries({ queryKey: tasksKey(path) });
      await qc.cancelQueries({ queryKey: taskKey(path, id) });
      const prevList = qc.getQueryData<api.Task[]>(tasksKey(path));
      const prevTask = qc.getQueryData<api.Task>(taskKey(path, id));
      qc.setQueryData<api.Task[]>(tasksKey(path), (old) =>
        old?.map((t) => (t.id === id ? { ...t, status: to } : t)),
      );
      qc.setQueryData<api.Task>(taskKey(path, id), (old) => (old ? { ...old, status: to } : old));
      return { prevList, prevTask, id };
    },
    onError: (err, _vars, ctx) => {
      if (ctx?.prevList !== undefined) qc.setQueryData(tasksKey(path), ctx.prevList);
      if (ctx?.prevTask !== undefined) qc.setQueryData(taskKey(path, ctx.id), ctx.prevTask);
      fail(err);
    },
    onSuccess: (t) => toast.success(`${t.id} → ${t.status}`),
    onSettled: (_d, _e, vars) => refresh(qc, path, vars.id),
  });
}

export function useRunChecks(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, only }: { id: string; only?: number[] }) => api.runChecks(path, id, only),
    onSuccess: (t) => {
      refresh(qc, path, t.id);
      const failed = (t.checks ?? []).filter((c) => c.result === "fail").length;
      if (failed) toast.error(`${failed} check(s) failed`);
      else toast.success("Checks passed");
    },
    onError: fail,
  });
}

export function useAttest(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, index, pass }: { id: string; index: number; pass: boolean }) =>
      api.attestTask(path, id, index, pass),
    onSuccess: (t) => {
      refresh(qc, path, t.id);
      toast.success("Check attested");
    },
    onError: fail,
  });
}

export function useReorder(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, rank }: { id: string; rank: number }) => api.reorderTask(path, id, rank),
    // silent: the board is optimistic; just reconcile on settle.
    onSettled: () => qc.invalidateQueries({ queryKey: tasksKey(path) }),
    onError: fail,
  });
}

export function useUpdateTask(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, fields }: { id: string; fields: import("./api").UpdateFields }) =>
      api.updateTask(path, id, fields),
    onSuccess: (t) => {
      refresh(qc, path, t.id);
      toast.success(`Updated ${t.id}`);
    },
    onError: fail,
  });
}

export function useAddNote(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, text }: { id: string; text: string }) => api.addNote(path, id, text),
    onSuccess: (t) => {
      refresh(qc, path, t.id);
      toast.success("Note added");
    },
    onError: fail,
  });
}

export function useDeleteTask(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.deleteTask(path, id),
    onSuccess: (r) => {
      qc.invalidateQueries({ queryKey: tasksKey(path) });
      qc.removeQueries({ queryKey: taskKey(path, r.id) });
      toast.success(`Deleted ${r.id}`);
    },
    onError: fail,
  });
}

export function useEditNote(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, text, note, index }: { id: string; text: string; note?: string; index?: number }) =>
      api.editNote(path, id, text, note, index),
    onSuccess: (t) => {
      refresh(qc, path, t.id);
      toast.success("Note updated");
    },
    onError: fail,
  });
}

export function useDeleteNote(path: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, note, index }: { id: string; note?: string; index?: number }) =>
      api.deleteNote(path, id, note, index),
    onSuccess: (t) => {
      refresh(qc, path, t.id);
      toast.success("Note deleted");
    },
    onError: fail,
  });
}
