import { useEffect, useRef } from "react";
import {
  Bell,
  CircleAlert,
  CircleDot,
  GitBranch,
  Sparkles,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useNotifications, type Notif } from "@/lib/notifications";
import { cn, timeAgo } from "@/lib/utils";

const ICON: Record<Notif["kind"], { Icon: typeof Bell; cls: string }> = {
  ready: { Icon: Sparkles, cls: "text-brand" },
  blocked: { Icon: GitBranch, cls: "text-muted-foreground" },
  failed: { Icon: CircleAlert, cls: "text-destructive" },
  assigned: { Icon: CircleDot, cls: "text-foreground" },
};

export function NotificationBell({
  path,
  actor,
  onOpenTask,
}: {
  path: string;
  actor?: string;
  onOpenTask: (id: string) => void;
}) {
  const { items, unread, markAllRead, markRead, clear } = useNotifications(path, actor);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);
  useEffect(() => () => clearTimeout(timer.current), []);

  const onOpenChange = (open: boolean) => {
    clearTimeout(timer.current);
    if (open && unread) timer.current = setTimeout(markAllRead, 1500);
  };

  return (
    <DropdownMenu onOpenChange={onOpenChange}>
      <DropdownMenuTrigger asChild>
        <button
          aria-label="Notifications"
          className="relative grid size-7 shrink-0 place-items-center rounded-md text-muted-foreground hover:bg-foreground/5 hover:text-foreground"
        >
          <Bell className="size-4" />
          {unread > 0 && (
            <span className="absolute -top-0.5 -right-0.5 grid h-3.5 min-w-3.5 place-items-center rounded-full bg-brand px-0.5 text-[9px] font-medium text-brand-foreground">
              {unread > 9 ? "9+" : unread}
            </span>
          )}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-72">
        <div className="flex items-center justify-between px-2 py-1.5">
          <span className="text-sm font-medium">Inbox</span>
          {items.length > 0 && (
            <button onClick={clear} className="text-xs text-muted-foreground hover:text-foreground">
              Clear
            </button>
          )}
        </div>
        <DropdownMenuSeparator />
        {items.length === 0 ? (
          <div className="px-2 py-6 text-center text-xs text-muted-foreground">
            You're all caught up
          </div>
        ) : (
          <div className="max-h-80 overflow-y-auto">
            {items.map((n) => {
              const { Icon, cls } = ICON[n.kind];
              return (
                <DropdownMenuItem
                  key={n.key}
                  onSelect={() => {
                    markRead(n.key);
                    onOpenTask(n.taskId);
                  }}
                  className="items-start gap-2"
                >
                  <Icon className={cn("mt-0.5 size-3.5 shrink-0", cls)} />
                  <span className="min-w-0 flex-1">
                    <span className={cn("block text-sm", !n.read && "font-medium")}>{n.text}</span>
                    <span className="text-xs text-muted-foreground">{timeAgo(new Date(n.at).toISOString())}</span>
                  </span>
                  {!n.read && <span className="mt-1.5 size-1.5 shrink-0 rounded-full bg-brand" />}
                </DropdownMenuItem>
              );
            })}
          </div>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
