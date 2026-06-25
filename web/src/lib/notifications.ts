// In-app notification inbox, derived by diffing the SSE-refreshed tasks query against the
// previous snapshot — no backend store. Events: ready, blocked, check-failed, assigned-to-me.

import { useEffect, useRef, useState } from "react";
import { useTasks } from "@/lib/queries";
import { notify } from "@/lib/desktop";
import type { Task } from "@/lib/api";

export type Notif = {
  key: string;
  kind: "ready" | "blocked" | "failed" | "assigned" | "review";
  taskId: string;
  text: string;
  at: number;
  read: boolean;
};

type Snap = {
  ready: boolean;
  status: string;
  assignee?: string;
  failed: number;
  execution?: string;
};

const MAX = 50;
const storeKey = (path: string) => `cairn-notifs:${path}`;

function load(path: string): Notif[] {
  try {
    return JSON.parse(localStorage.getItem(storeKey(path)) || "[]");
  } catch {
    return [];
  }
}

function snapshot(t: Task): Snap {
  return {
    ready: t.ready,
    status: t.status,
    assignee: t.assignee,
    failed: (t.checks ?? []).filter((c) => c.result === "fail").length,
    execution: t.executionState,
  };
}

let seq = 0; // monotonic suffix so keys are unique even within one millisecond

export function useNotifications(path: string, actor?: string) {
  const { data: tasks } = useTasks(path);
  const prev = useRef<Map<string, Snap> | null>(null);
  const [items, setItems] = useState<Notif[]>(() => load(path));
  const itemsRef = useRef(items);
  // Assign synchronously during render so closures always see the latest items
  // without the one-render lag that a useEffect assignment would introduce.
  itemsRef.current = items;

  const persist = (next: Notif[]) => {
    itemsRef.current = next;
    localStorage.setItem(storeKey(path), JSON.stringify(next));
    setItems(next);
  };

  // Switching workspace: drop the old snapshot + load the new workspace's inbox, so we
  // don't diff across unrelated id spaces.
  useEffect(() => {
    prev.current = null;
    const loaded = load(path);
    itemsRef.current = loaded;
    setItems(loaded);
  }, [path]);

  useEffect(() => {
    if (!tasks) return;
    const snap = new Map(tasks.map((t) => [t.id, snapshot(t)]));
    const fresh: Notif[] = [];
    const now = Date.now();
    const add = (kind: Notif["kind"], taskId: string, text: string) =>
      fresh.push({ key: `${taskId}-${kind}-${now}-${seq++}`, kind, taskId, text, at: now, read: false });

    if (prev.current) {
      for (const t of tasks) {
        const p = prev.current.get(t.id);
        if (!p) continue; // newly seen task — don't notify on first sight
        const s = snapshot(t);
        if (s.ready && !p.ready) add("ready", t.id, `${t.id} is ready to start`);
        else if (!s.ready && p.ready) add("blocked", t.id, `${t.id} is blocked by deps`);
        if (s.failed > p.failed) add("failed", t.id, `${t.id} — a check failed`);
        if (s.execution === "awaiting_review" && p.execution !== "awaiting_review")
          add("review", t.id, `${t.id} is awaiting review`);
        if (actor && s.assignee === actor && p.assignee !== actor)
          add("assigned", t.id, `You were assigned ${t.id}`);
      }
    }
    prev.current = snap;
    if (fresh.length) {
      persist([...fresh, ...itemsRef.current].slice(0, MAX));
      // Mirror the actionable ones to OS notifications when the app isn't focused
      // (desktop only; no-op in the browser). Avoids double-signalling the visible bell.
      if (!document.hasFocus()) {
        for (const n of fresh) {
          if (n.kind !== "blocked") void notify("Cairn", n.text);
        }
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tasks, actor, path]);
  const markAllRead = () => persist(items.map((n) => ({ ...n, read: true })));
  const markRead = (key: string) => persist(items.map((n) => (n.key === key ? { ...n, read: true } : n)));
  const clear = () => persist([]);
  const unread = items.filter((n) => !n.read).length;

  return { items, unread, markAllRead, markRead, clear };
}
