// Typed client for the cairn web API. Every call targets a project `path`; the server
// resolves it (falling back to its launch --repo) and reuses mcp.Service, so the web and
// agent front-ends share one rule-set.
import { currentActor } from "@/lib/identity";

export type Check = {
  desc: string;
  cmd?: string;
  type?: string;
  result?: string; // pending | pass | fail
  cwd?: string;
  timeout?: number;
};

export type Provenance = {
  id?: string; // stable id, present on note entries only (used to edit/delete)
  who: string;
  at: string;
  did: string;
  text?: string;
  editedAt?: string; // set when a note is edited in place
};

export type ExecutionState = "active" | "stalled" | "awaiting_review";

export type SessionLive = {
  session: string;
  heartbeatAt: string;
  progress?: string;
  worktree?: string;
};

export type AgentSession = {
  id: string;
  task: string;
  attempt: string;
  actor: string;
  client?: string;
  model?: string;
  status: "active" | "finished" | "canceled";
  idempotencyKey: string;
  startedAt: string;
  endedAt?: string;
  branch?: string;
  headStarted?: string;
  headFinished?: string;
  summary?: string;
  cancelReason?: string;
  health: "active" | "stalled" | "finished" | "canceled";
  live?: SessionLive;
};

// Run is one parsed check-run log from .cairn/runs (output stays out of the task file).
export type Run = {
  file: string;
  at?: string;
  cmd?: string;
  cwd?: string;
  exit: number;
  timedout: boolean;
  duration?: string;
  output?: string;
};

export type Task = {
  id: string;
  title: string;
  status: string;
  assignee?: string;
  deps?: string[];
  ready: boolean;
  updatedAt?: string; // newest provenance timestamp (list view)
  rank?: number; // manual board ordering (0/absent = unset)
  labels?: string[];
  priority?: string; // "" | low | medium | high | urgent
  parent?: string;
  checks?: Check[];
  provenance?: Provenance[];
  body?: string;
  activeAttempt?: string;
  executionState?: ExecutionState;
  sessionId?: string;
};

export type UpdateFields = {
  priority?: string;
  labels?: string[];
  parent?: string;
  title?: string;
  body?: string;
  checks?: Check[];
};

export type Status = {
  initialized: boolean;
  root: string;
  prefix?: string;
  suggestedPrefix: string;
  states?: string[];
  closed?: string[];
  initial?: string;
  review?: string; // state whose entry runs command checks (alongside closed states)
  actor?: string;
  suggestedActor?: string;
};

export type CreateInput = {
  title: string;
  body?: string;
  deps?: string[];
  checks?: Check[];
  labels?: string[];
  priority?: string;
  parent?: string;
};

const enc = encodeURIComponent;

async function req<T>(method: string, url: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  const actor = currentActor();
  if (actor) headers["X-Cairn-Actor"] = enc(actor); // who I am (URL-encoded for non-ASCII)
  const res = await fetch(url, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  const text = await res.text();
  const data = text ? JSON.parse(text) : null;
  if (!res.ok) throw new Error(data?.error || `${res.status} ${res.statusText}`);
  return data as T;
}

export const getStatus = (path: string) =>
  req<Status>("GET", `/api/status?path=${enc(path)}`);

export const initRepo = (path: string, prefix?: string) =>
  req<Status>("POST", "/api/init", { path, prefix });

export const listTasks = (path: string) =>
  req<{ tasks: Task[] }>("GET", `/api/tasks?path=${enc(path)}`).then((r) => r.tasks ?? []);

export const getTask = (path: string, id: string) =>
  req<Task>("GET", `/api/tasks/${id}?path=${enc(path)}`);

export const createTask = (path: string, input: CreateInput) =>
  req<Task>("POST", `/api/tasks?path=${enc(path)}`, input);

export const transitionTask = (path: string, id: string, to: string) =>
  req<Task>("POST", `/api/tasks/${id}/transition?path=${enc(path)}`, { to });

export const claimTask = (path: string, id: string) =>
  req<Task>("POST", `/api/tasks/${id}/claim?path=${enc(path)}`);

export const runChecks = (path: string, id: string, only?: number[]) =>
  req<Task>("POST", `/api/tasks/${id}/run_checks?path=${enc(path)}`, { only });

export const addNote = (path: string, id: string, text: string) =>
  req<Task>("POST", `/api/tasks/${id}/note?path=${enc(path)}`, { text });

export const updateTask = (path: string, id: string, fields: UpdateFields) =>
  req<Task>("POST", `/api/tasks/${id}/update?path=${enc(path)}`, fields);

export const reorderTask = (path: string, id: string, rank: number) =>
  req<Task>("POST", `/api/tasks/${id}/reorder?path=${enc(path)}`, { rank });

export const getRuns = (path: string, id: string) =>
  req<{ runs: Run[] }>("GET", `/api/tasks/${id}/runs?path=${enc(path)}`).then((r) => r.runs ?? []);

export const listTaskSessions = (path: string, id: string) =>
  req<{ sessions: AgentSession[] }>("GET", `/api/tasks/${id}/sessions?path=${enc(path)}`).then(
    (r) => r.sessions ?? [],
  );

export const attestTask = (path: string, id: string, index: number, pass: boolean) =>
  req<Task>("POST", `/api/tasks/${id}/attest?path=${enc(path)}`, { index, pass });

export const deleteTask = (path: string, id: string) =>
  req<{ id: string; deleted: boolean }>("DELETE", `/api/tasks/${id}?path=${enc(path)}`);

// noteURL addresses a note by its stable id, or by 0-based provenance index for a legacy
// note with no id (server sentinel segment "-" + ?index=).
const noteURL = (path: string, id: string, note?: string, index?: number) => {
  const seg = note ? enc(note) : "-";
  const idx = note ? "" : `&index=${index ?? -1}`;
  return `/api/tasks/${id}/notes/${seg}?path=${enc(path)}${idx}`;
};

export const editNote = (path: string, id: string, text: string, note?: string, index?: number) =>
  req<Task>("PATCH", noteURL(path, id, note, index), { text });

export const deleteNote = (path: string, id: string, note?: string, index?: number) =>
  req<Task>("DELETE", noteURL(path, id, note, index));
