// Typed client for the cairn web API. Every call targets a project `path`; the server
// resolves it (falling back to its launch --repo) and reuses mcp.Service, so the web and
// agent front-ends share one rule-set.

export type Check = {
  desc: string;
  cmd?: string;
  type?: string;
  result?: string; // pending | pass | fail
  cwd?: string;
  timeout?: number;
};

export type Provenance = { who: string; at: string; did: string; text?: string };

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
};

export type UpdateFields = {
  priority?: string;
  labels?: string[];
  parent?: string;
};

export type Status = {
  initialized: boolean;
  root: string;
  prefix?: string;
  suggestedPrefix: string;
  states?: string[];
  closed?: string[];
  initial?: string;
  actor?: string;
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
  const res = await fetch(url, {
    method,
    headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
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

export const attestTask = (path: string, id: string, index: number, pass: boolean) =>
  req<Task>("POST", `/api/tasks/${id}/attest?path=${enc(path)}`, { index, pass });
