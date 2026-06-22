import { CircleAlert, Signal, SignalHigh, SignalLow, SignalMedium } from "lucide-react";
import { cn } from "@/lib/utils";

// Priority levels, highest first (for selects).
export const PRIORITIES = ["urgent", "high", "medium", "low"] as const;

const MAP: Record<string, { Icon: typeof Signal; cls: string }> = {
  urgent: { Icon: CircleAlert, cls: "text-destructive" },
  high: { Icon: SignalHigh, cls: "text-foreground" },
  medium: { Icon: SignalMedium, cls: "text-muted-foreground" },
  low: { Icon: SignalLow, cls: "text-muted-foreground" },
};

export function priorityLabel(p?: string): string {
  return p ? p.charAt(0).toUpperCase() + p.slice(1) : "No priority";
}

// PriorityIcon renders a signal-bar glyph; "none"/unknown shows a faint placeholder.
export function PriorityIcon({ priority, className }: { priority?: string; className?: string }) {
  const m = priority ? MAP[priority] : undefined;
  if (!m) return <Signal className={cn("size-3.5 text-muted-foreground/40", className)} />;
  return <m.Icon className={cn("size-3.5", m.cls, className)} />;
}
