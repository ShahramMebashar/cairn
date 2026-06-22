import { CircleDot, ClockAlert, ScanEye } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import type { ExecutionState } from "@/lib/api";

const states = {
  active: { label: "Active", icon: CircleDot, className: "text-brand" },
  stalled: { label: "Stalled", icon: ClockAlert, className: "text-destructive" },
  awaiting_review: { label: "Awaiting review", icon: ScanEye, className: "text-foreground" },
} satisfies Record<ExecutionState, { label: string; icon: typeof CircleDot; className: string }>;

export function SessionStatus({ state, compact = false }: { state?: ExecutionState; compact?: boolean }) {
  if (!state) return null;
  const { label, icon: Icon, className } = states[state];

  if (compact) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <span aria-label={label} className={cn("inline-flex shrink-0", className)}>
            <Icon className="size-3.5" />
          </span>
        </TooltipTrigger>
        <TooltipContent>{label}</TooltipContent>
      </Tooltip>
    );
  }

  return (
    <Badge variant="outline" className={cn("font-normal", className)}>
      <Icon />
      {label}
    </Badge>
  );
}
