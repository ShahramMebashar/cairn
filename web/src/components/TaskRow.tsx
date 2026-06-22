import { CheckCircle2, GitBranch, UserPlus } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { PriorityIcon } from "@/components/PriorityIcon";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { StatusIcon } from "@/components/StatusIcon";
import { useClaim, useTransition } from "@/lib/queries";
import { cn, initials, statusLabel, timeAgo } from "@/lib/utils";
import type { Status, Task } from "@/lib/api";

export function TaskRow({
  path,
  task,
  status,
  onOpen,
  selected,
}: {
  path: string;
  task: Task;
  status: Status;
  onOpen: (id: string) => void;
  selected?: boolean;
}) {
  const transition = useTransition(path);
  const claim = useClaim(path);

  const checks = task.checks ?? [];
  const passed = checks.filter((c) => c.result === "pass").length;
  const allPass = checks.length > 0 && passed === checks.length;
  const stop = (e: React.MouseEvent) => e.stopPropagation();

  return (
    <div
      id={`row-${task.id}`}
      role="button"
      tabIndex={0}
      onClick={() => onOpen(task.id)}
      onKeyDown={(e) => {
        if (e.key === "Enter") onOpen(task.id);
      }}
      className={cn(
        "group flex h-8 w-full cursor-pointer items-center gap-2 px-3 text-left transition-colors hover:bg-foreground/[0.04] focus-visible:bg-foreground/[0.04] focus-visible:outline-none",
        selected && "bg-foreground/[0.06] ring-1 ring-inset ring-brand/40",
      )}
    >
      {/* priority occupies the leading column (aligns with the section header chevron) */}
      <PriorityIcon priority={task.priority} className="shrink-0" />

      {/* inline status change */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild onClick={stop}>
          <button
            aria-label="Change status"
            className="grid size-4 shrink-0 place-items-center rounded hover:ring-2 hover:ring-foreground/10"
          >
            <StatusIcon
              status={task.status}
              closed={status.closed}
              initial={status.initial}
              className="size-3.5"
            />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" onClick={stop}>
          {(status.states ?? []).map((s) => (
            <DropdownMenuItem
              key={s}
              disabled={s === task.status || transition.isPending}
              onSelect={() => transition.mutate({ id: task.id, to: s })}
            >
              <StatusIcon status={s} closed={status.closed} initial={status.initial} />
              {statusLabel(s)}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>

      <span className="w-20 shrink-0 truncate font-mono text-xs whitespace-nowrap text-muted-foreground">
        {task.id}
      </span>
      <span className="flex-1 truncate text-[13px]">{task.title}</span>

      {task.labels?.slice(0, 2).map((l) => (
        <Badge
          key={l}
          variant="secondary"
          className="hidden h-4 shrink-0 px-1.5 text-[10px] font-normal sm:inline-flex"
        >
          {l}
        </Badge>
      ))}
      {task.labels && task.labels.length > 2 && (
        <span className="hidden shrink-0 text-[10px] text-muted-foreground sm:inline">
          +{task.labels.length - 2}
        </span>
      )}

      {task.ready && task.status === status.initial && (
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="size-1.5 shrink-0 rounded-full bg-brand" />
          </TooltipTrigger>
          <TooltipContent>Ready to start — deps closed</TooltipContent>
        </Tooltip>
      )}

      {task.deps && task.deps.length > 0 && (
        <span className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
          <GitBranch className="size-3" />
          {task.deps.length}
        </span>
      )}

      {checks.length > 0 && (
        <span
          className={cn(
            "flex shrink-0 items-center gap-1 text-xs tabular-nums",
            allPass ? "text-success" : "text-muted-foreground",
          )}
        >
          <CheckCircle2 className="size-3" />
          {passed}/{checks.length}
        </span>
      )}

      {task.updatedAt && (
        <span className="hidden shrink-0 text-xs text-muted-foreground tabular-nums sm:block">
          {timeAgo(task.updatedAt)}
        </span>
      )}

      {/* assignee, or a claim action on hover when unassigned */}
      {task.assignee ? (
        <Tooltip>
          <TooltipTrigger asChild>
            <Avatar className="size-5">
              <AvatarFallback className="text-[9px]">{initials(task.assignee)}</AvatarFallback>
            </Avatar>
          </TooltipTrigger>
          <TooltipContent>{task.assignee}</TooltipContent>
        </Tooltip>
      ) : (
        <button
          aria-label="Claim"
          disabled={claim.isPending}
          onClick={(e) => {
            stop(e);
            claim.mutate(task.id);
          }}
          className="grid size-5 shrink-0 place-items-center rounded text-muted-foreground opacity-0 hover:bg-foreground/10 hover:text-foreground group-hover:opacity-100 focus-visible:opacity-100"
        >
          <UserPlus className="size-3.5" />
        </button>
      )}
    </div>
  );
}
