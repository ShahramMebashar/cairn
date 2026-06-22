import {
  Activity,
  ChevronsUpDown,
  CircleDashed,
  ClockAlert,
  HelpCircle,
  ListTodo,
  Moon,
  Network,
  Pencil,
  PenSquare,
  ScanEye,
  Sparkles,
  SquareKanban,
  Sun,
} from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
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
import { HelpDialog } from "@/components/HelpDialog";
import { Assignee } from "@/components/Assignee";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useIdentity, displayName } from "@/lib/identity";
import { useTasks } from "@/lib/queries";
import type { Status } from "@/lib/api";

export type Filter = "all" | "active" | "stalled" | "review" | "backlog" | "ready";

const TASK_NAV: { key: Filter; label: string; icon: typeof ListTodo }[] = [
  { key: "all", label: "All tasks", icon: ListTodo },
  { key: "backlog", label: "Backlog", icon: CircleDashed },
  { key: "ready", label: "Ready", icon: Sparkles },
];

const SESSION_NAV: { key: Filter; label: string; icon: typeof ListTodo }[] = [
  { key: "active", label: "Active", icon: Activity },
  { key: "stalled", label: "Stalled", icon: ClockAlert },
  { key: "review", label: "Awaiting review", icon: ScanEye },
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
  const [helpOpen, setHelpOpen] = useState(false);
  const { actor, setName } = useIdentity(status.suggestedActor);
  const { data: tasks } = useTasks(path);
  const folderName = status.root.split("/").filter(Boolean).pop() ?? status.root;
  const counts: Partial<Record<Filter, number>> = {
    active: tasks?.filter((task) => task.executionState === "active").length,
    stalled: tasks?.filter((task) => task.executionState === "stalled").length,
    review: tasks?.filter((task) => task.executionState === "awaiting_review").length,
  };

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
        <NotificationBell path={path} actor={actor} onOpenTask={onOpenTask} />
      </div>

      <nav className="flex-1 space-y-px px-2">
        {TASK_NAV.map(({ key, label, icon: Icon }) => (
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
        <p className="px-2 pb-1 pt-1.5 text-[11px] font-medium text-muted-foreground">Agent work</p>
        {SESSION_NAV.map(({ key, label, icon: Icon }) => (
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
            <span>{label}</span>
            {!!counts[key] && (
              <span className="ml-auto text-xs tabular-nums text-muted-foreground">{counts[key]}</span>
            )}
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

      <div className="flex flex-col gap-1 px-3 py-2">
        <IdentityChip key={actor} actor={actor} onRename={setName} />
        <div className="flex items-center justify-between">
          <span className="truncate text-xs text-muted-foreground">{folderName}</span>
          <div className="flex items-center gap-0.5">
            <button
              onClick={() => setHelpOpen(true)}
              aria-label="How cairn works"
              className="grid size-7 shrink-0 place-items-center rounded-md text-muted-foreground hover:bg-foreground/5 hover:text-foreground"
            >
              <HelpCircle className="size-4" />
            </button>
            <button
              onClick={() => setTheme(toggleTheme())}
              aria-label="Toggle theme"
              className="grid size-7 shrink-0 place-items-center rounded-md text-muted-foreground hover:bg-foreground/5 hover:text-foreground"
            >
              {theme === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
            </button>
          </div>
        </div>
      </div>
      <HelpDialog open={helpOpen} onOpenChange={setHelpOpen} />
    </aside>
  );
}

function IdentityChip({ actor, onRename }: { actor: string; onRename: (name: string) => void }) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState(displayName(actor));

  const save = () => {
    const n = name.trim();
    if (n) onRename(n);
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button className="group flex items-center gap-2 rounded-md px-1 py-1 text-left hover:bg-foreground/5">
          <Assignee actor={actor || "human:?"} />
          <span className="min-w-0 flex-1 truncate text-[13px]">
            {displayName(actor) || "Set your name"}
          </span>
          <Pencil className="size-3 text-muted-foreground opacity-0 group-hover:opacity-100" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-60">
        <p className="mb-2 text-xs text-muted-foreground">
          Your name — stamped on everything you do here.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            save();
          }}
          className="flex gap-2"
        >
          <Input
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. shahram"
            className="h-8 text-sm"
          />
          <Button type="submit" size="sm" className="h-8">
            Save
          </Button>
        </form>
      </PopoverContent>
    </Popover>
  );
}
