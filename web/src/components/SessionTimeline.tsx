import { useEffect, useState } from "react";
import { Bot, GitBranch, Timer, Wrench } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { SessionStatus } from "@/components/SessionStatus";
import { cn, timeAgo } from "@/lib/utils";
import type { AgentSession, ExecutionState, Usage } from "@/lib/api";

/**
 * Returns the current timestamp and re-renders every `ms` ms when `active` is true.
 * When `active` is false no interval is created, so non-live sessions are free of overhead.
 */
function useNow(ms: number, active: boolean): number {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!active) return;
    const id = setInterval(() => setNow(Date.now()), ms);
    return () => clearInterval(id);
  }, [ms, active]);
  return now;
}

export function SessionTimeline({
  sessions,
  executionState,
  loading,
}: {
  sessions: AgentSession[];
  executionState?: ExecutionState;
  loading: boolean;
}) {
  if (loading) {
    return (
      <section className="mt-8 space-y-2">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-28 w-full" />
      </section>
    );
  }
  if (sessions.length === 0) return null;

  return (
    <section className="mt-8">
      <div className="flex items-center justify-between">
        <h2 className="text-xs font-medium text-muted-foreground">Agent sessions</h2>
        <SessionStatus state={executionState} />
      </div>
      <div className="mt-3 space-y-2">
        {sessions.map((session) => (
          <SessionCard key={session.id} session={session} />
        ))}
      </div>
    </section>
  );
}

function SessionCard({ session }: { session: AgentSession }) {
  const isLive = session.health === "active";
  // Tick every 30 s so timeAgo() recomputes; only runs when the session is live.
  useNow(30_000, isLive);

  const usage = session.live?.usage ?? session.usage;
  const heartbeat = session.live?.heartbeatAt;
  const detail = session.live?.progress || session.summary || session.cancelReason;
  const usageSummary = usageText(usage);

  return (
    <article className="rounded-lg border bg-background px-3.5 py-3">
      <div className="flex min-w-0 items-start gap-2.5">
        <span className="mt-0.5 grid size-6 shrink-0 place-items-center rounded-md bg-muted text-muted-foreground">
          <Bot className="size-3.5" />
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-1.5">
            <span className="text-sm font-medium">{session.actor}</span>
            <HealthBadge health={session.health} />
            {/* For live sessions show heartbeat age (ticks); for ended sessions show end/start time. */}
            <span className="ml-auto text-xs text-muted-foreground tabular-nums">
              {timeAgo(heartbeat ?? session.endedAt ?? session.startedAt)}
            </span>
          </div>
          {(session.client || session.model) && (
            <p className="mt-0.5 truncate text-xs text-muted-foreground">
              {[session.client, session.model].filter(Boolean).join(" · ")}
            </p>
          )}
        </div>
      </div>

      {detail && (
        <p className="mt-2.5 whitespace-pre-wrap rounded-md bg-muted/60 px-2.5 py-2 text-sm leading-relaxed">
          {detail}
        </p>
      )}

      <div className="mt-2.5 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
        {/* Branch/worktree */}
        {(session.branch || session.live?.worktree) && (
          <span className="flex min-w-0 items-center gap-1">
            <GitBranch className="size-3" />
            <span className="max-w-52 truncate">{session.branch || session.live?.worktree}</span>
          </span>
        )}
        {/* Heartbeat — only show the detail row for live sessions where the timestamp
            in the header might not be obvious enough. */}
        {isLive && heartbeat && (
          <span className="flex items-center gap-1">
            <Timer className="size-3" /> Heartbeat {timeAgo(heartbeat)}
          </span>
        )}
        {/* Usage — show whenever usage object is present, even if counts are zero. */}
        {usage !== undefined && usageSummary && (
          <span className="flex items-center gap-1">
            <Wrench className="size-3" /> {usageSummary}
          </span>
        )}
      </div>
    </article>
  );
}

function HealthBadge({ health }: { health: AgentSession["health"] }) {
  return (
    <Badge
      variant={health === "canceled" || health === "stalled" ? "destructive" : "outline"}
      className={cn(
        "h-4 px-1.5 text-[10px] font-normal",
        health === "active" && "text-brand",
        health === "finished" && "text-success",
      )}
    >
      {health}
    </Badge>
  );
}

function usageText(usage?: Usage) {
  if (!usage) return "";
  const tokens = (usage.inputTokens ?? 0) + (usage.outputTokens ?? 0);
  const parts = [];
  // Always show token count when usage is present (even 0 = session started but no tokens yet).
  parts.push(`${new Intl.NumberFormat(undefined, { notation: "compact" }).format(tokens)} tokens`);
  if (usage.toolCalls != null) parts.push(`${usage.toolCalls} tool calls`);
  return parts.join(" · ");
}
