import type { LucideIcon } from "lucide-react";
import { Button } from "@/components/ui/button";

// EmptyState is the shared "nothing here" panel used across list / board / graph.
export function EmptyState({
  icon: Icon,
  title,
  message,
  action,
}: {
  icon: LucideIcon;
  title: string;
  message?: string;
  action?: { label: string; icon?: LucideIcon; onClick: () => void };
}) {
  return (
    <div className="flex h-full flex-col items-center justify-center px-6 text-center">
      <span className="grid size-12 place-items-center rounded-full bg-muted text-muted-foreground">
        <Icon className="size-5" />
      </span>
      <h2 className="mt-4 text-sm font-medium">{title}</h2>
      {message && <p className="mt-1 max-w-xs text-sm text-muted-foreground">{message}</p>}
      {action && (
        <Button className="mt-4" size="sm" onClick={action.onClick}>
          {action.icon && <action.icon />} {action.label}
        </Button>
      )}
    </div>
  );
}
