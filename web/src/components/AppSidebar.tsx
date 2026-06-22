import {
  ChevronsUpDown,
  CircleDashed,
  CircleDot,
  ListTodo,
  Moon,
  Network,
  PenSquare,
  Sparkles,
  SquareKanban,
  Sun,
} from "lucide-react";
import { useState } from "react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { getTheme, toggleTheme, type Theme } from "@/lib/theme";
import { NotificationBell } from "@/components/NotificationBell";
import type { Status } from "@/lib/api";

export type Filter = "all" | "active" | "backlog" | "ready";

const NAV: { key: Filter; label: string; icon: typeof ListTodo }[] = [
  { key: "all", label: "All tasks", icon: ListTodo },
  { key: "active", label: "Active", icon: CircleDot },
  { key: "backlog", label: "Backlog", icon: CircleDashed },
  { key: "ready", label: "Ready", icon: Sparkles },
];

export function AppSidebar({
  path,
  status,
  active,
  graphActive,
  boardActive,
  onFilter,
  onGraph,
  onBoard,
  onChangeFolder,
  onNewTask,
  onOpenTask,
}: {
  path: string;
  status: Status;
  active: Filter | null;
  graphActive: boolean;
  boardActive: boolean;
  onFilter: (f: Filter) => void;
  onGraph: () => void;
  onBoard: () => void;
  onChangeFolder: () => void;
  onNewTask: () => void;
  onOpenTask: (id: string) => void;
}) {
  const [theme, setTheme] = useState<Theme>(getTheme());
  const folderName = status.root.split("/").filter(Boolean).pop() ?? status.root;

  return (
    <aside className="flex w-[15rem] shrink-0 flex-col">
      <div className="flex items-center gap-1 px-2 py-2.5">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="flex min-w-0 flex-1 items-center gap-2 rounded-md px-2 py-1.5 text-left hover:bg-foreground/5">
              <span className="grid size-5 shrink-0 place-items-center rounded bg-primary text-[10px] font-semibold text-primary-foreground">
                {status.prefix?.slice(0, 1) ?? "C"}
              </span>
              <span className="truncate text-[13px] font-medium">{status.prefix}</span>
              <ChevronsUpDown className="ml-auto size-3.5 shrink-0 text-muted-foreground" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="w-56">
            <DropdownMenuLabel className="truncate font-mono text-xs font-normal text-muted-foreground">
              {status.root}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={onChangeFolder}>Switch folder…</DropdownMenuItem>
            <DropdownMenuItem onClick={() => setTheme(toggleTheme())}>
              {theme === "dark" ? "Light theme" : "Dark theme"}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        <button
          onClick={onNewTask}
          aria-label="New task"
          className="grid size-7 shrink-0 place-items-center rounded-md text-muted-foreground hover:bg-foreground/5 hover:text-foreground"
        >
          <PenSquare className="size-4" />
        </button>
        <NotificationBell path={path} actor={status.actor} onOpenTask={onOpenTask} />
      </div>

      <nav className="flex-1 space-y-px px-2">
        {NAV.map(({ key, label, icon: Icon }) => (
          <button
            key={key}
            onClick={() => onFilter(key)}
            aria-current={active === key ? "page" : undefined}
            className={cn(
              "flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-[13px] transition-colors",
              active === key
                ? "bg-foreground/[0.07] font-medium text-foreground"
                : "text-muted-foreground hover:bg-foreground/5 hover:text-foreground",
            )}
          >
            <Icon className="size-4 shrink-0" />
            {label}
          </button>
        ))}

        <div className="my-1.5 border-t" />
        <button
          onClick={onBoard}
          aria-current={boardActive ? "page" : undefined}
          className={cn(
            "flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-[13px] transition-colors",
            boardActive
              ? "bg-foreground/[0.07] font-medium text-foreground"
              : "text-muted-foreground hover:bg-foreground/5 hover:text-foreground",
          )}
        >
          <SquareKanban className="size-4 shrink-0" />
          Board
        </button>
        <button
          onClick={onGraph}
          aria-current={graphActive ? "page" : undefined}
          className={cn(
            "flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-[13px] transition-colors",
            graphActive
              ? "bg-foreground/[0.07] font-medium text-foreground"
              : "text-muted-foreground hover:bg-foreground/5 hover:text-foreground",
          )}
        >
          <Network className="size-4 shrink-0" />
          Graph
        </button>
      </nav>

      <div className="flex items-center justify-between px-3 py-2">
        <span className="truncate text-xs text-muted-foreground">{folderName}</span>
        <button
          onClick={() => setTheme(toggleTheme())}
          aria-label="Toggle theme"
          className="grid size-7 shrink-0 place-items-center rounded-md text-muted-foreground hover:bg-foreground/5 hover:text-foreground"
        >
          {theme === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
        </button>
      </div>
    </aside>
  );
}
