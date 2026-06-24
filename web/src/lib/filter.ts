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

// Lowercase Crockford base32 alphabet, matching the Go minter (store/id.go). Strictly
// ascending, so decoding preserves chronological order.
const CROCKFORD = "0123456789abcdefghjkmnpqrstvwxyz";

// effectiveRank gives every task a comparable ordering value: its manual rank if set, else a
// creation-order proxy derived from its id so unranked tasks order by creation until first
// dragged. Legacy ids (`PROJ-001`) decode to their small integer; time-ordered ids
// (`PROJ-<base32 UnixMilli>...`) decode the 10-char timestamp prefix to a millisecond value.
// Legacy integers (≤ a few thousand) always sort before time-ordered millis, so a mixed
// board keeps old tasks ahead of new ones.
export function effectiveRank(t: Task): number {
  if (t.rank) return t.rank;
  const suffix = t.id.slice(t.id.indexOf("-") + 1);
  if (/^\d+$/.test(suffix)) return Number(suffix); // legacy counter id
  let ms = 0;
  for (let i = 0; i < 10 && i < suffix.length; i++) {
    const v = CROCKFORD.indexOf(suffix[i]);
    if (v < 0) return 0;
    ms = ms * 32 + v;
  }
  return ms;
}
