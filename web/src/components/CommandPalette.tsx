import { useEffect, useState } from "react";
import {
  Activity,
  CircleDashed,
  ClockAlert,
  FolderOpen,
  ListTodo,
  Moon,
  Network,
  Plus,
  ScanEye,
  Sparkles,
  SquareKanban,
} from "lucide-react";
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import { StatusIcon } from "@/components/StatusIcon";
import { useTasks } from "@/lib/queries";
import { toggleTheme } from "@/lib/theme";
import type { Filter } from "@/components/AppSidebar";
import type { Status } from "@/lib/api";

const VIEWS: { key: Filter; label: string; icon: typeof ListTodo }[] = [
  { key: "all", label: "All tasks", icon: ListTodo },
  { key: "backlog", label: "Backlog", icon: CircleDashed },
  { key: "ready", label: "Ready", icon: Sparkles },
  { key: "active", label: "Active agent work", icon: Activity },
  { key: "stalled", label: "Stalled agent work", icon: ClockAlert },
  { key: "review", label: "Awaiting review", icon: ScanEye },
];

export function CommandPalette({
  path,
  status,
  onView,
  onOpenTask,
  onNewTask,
  onChangeFolder,
  onGraph,
  onBoard,
}: {
  path: string;
  status: Status;
  onView: (f: Filter) => void;
  onOpenTask: (id: string) => void;
  onNewTask: () => void;
  onChangeFolder: () => void;
  onGraph: () => void;
  onBoard: () => void;
}) {
  const [open, setOpen] = useState(false);
  const { data: tasks } = useTasks(path);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setOpen((o) => !o);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  const run = (fn: () => void) => {
    setOpen(false);
    fn();
  };

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <Command>
        <CommandInput placeholder="Search tasks or run a command…" />
        <CommandList>
        <CommandEmpty>No results.</CommandEmpty>

        <CommandGroup heading="Actions">
          <CommandItem onSelect={() => run(onNewTask)}>
            <Plus /> New task
          </CommandItem>
          <CommandItem onSelect={() => run(onBoard)}>
            <SquareKanban /> Board
          </CommandItem>
          <CommandItem onSelect={() => run(onGraph)}>
            <Network /> Dependency graph
          </CommandItem>
          <CommandItem onSelect={() => run(() => toggleTheme())}>
            <Moon /> Toggle theme
          </CommandItem>
          <CommandItem onSelect={() => run(onChangeFolder)}>
            <FolderOpen /> Switch folder…
          </CommandItem>
        </CommandGroup>

        <CommandGroup heading="Views">
          {VIEWS.map(({ key, label, icon: Icon }) => (
            <CommandItem key={key} onSelect={() => run(() => onView(key))}>
              <Icon /> {label}
            </CommandItem>
          ))}
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Tasks">
          {(tasks ?? []).map((t) => (
            <CommandItem
              key={t.id}
              value={`${t.id} ${t.title}`}
              onSelect={() => run(() => onOpenTask(t.id))}
            >
              <StatusIcon status={t.status} closed={status.closed} initial={status.initial} />
              <span className="font-mono text-xs text-muted-foreground">{t.id}</span>
              <span className="truncate">{t.title}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        </CommandList>
      </Command>
    </CommandDialog>
  );
}
