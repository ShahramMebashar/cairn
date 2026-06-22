import type { Filter } from "@/components/AppSidebar";
import type { Status, Task } from "@/lib/api";

export const FILTER_LABEL: Record<Filter, string> = {
  all: "All tasks",
  active: "Active",
  stalled: "Stalled",
  review: "Awaiting review",
  backlog: "Backlog",
  ready: "Ready",
};

// matches applies the base view filter (the sidebar nav) to a task.
export function matches(t: Task, filter: Filter, status: Status): boolean {
  const closed = status.closed ?? [];
  switch (filter) {
    case "active":
      return !(status.closed ?? []).includes(t.status) && t.status !== status.initial;
    case "stalled":
      return t.executionState === "stalled";
    case "review":
      return t.executionState === "awaiting_review";
    case "backlog":
      return t.status === status.initial;
    case "ready":
      return t.ready && !closed.includes(t.status);
    default:
      return true;
  }
}

// effectiveRank gives every task a comparable ordering value: its manual rank if set,
// else its numeric id so unranked tasks order by creation until first dragged.
export function effectiveRank(t: Task): number {
  if (t.rank) return t.rank;
  const n = Number(t.id.replace(/\D+/g, ""));
  return Number.isNaN(n) ? 0 : n;
}
