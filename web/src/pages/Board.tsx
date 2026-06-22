import { useEffect, useRef, useState } from "react";
import { Bookmark, ChevronRight, Inbox, Plus, Search, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { SelectItem } from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { PriorityIcon, priorityLabel } from "@/components/PriorityIcon";
import { Facet } from "@/components/Facet";
import { EmptyState } from "@/components/EmptyState";
import { Onboarding } from "@/components/Onboarding";
import { addView, loadViews, removeView, type SavedView } from "@/lib/views";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { TaskRow } from "@/components/TaskRow";
import { StatusIcon } from "@/components/StatusIcon";
import { useTasks } from "@/lib/queries";
import { cn, statusLabel } from "@/lib/utils";
import { FILTER_LABEL, matches } from "@/lib/filter";
import type { Filter } from "@/components/AppSidebar";
import type { Status, Task } from "@/lib/api";

export function Board({
  path,
  status,
  filter,
  onOpenTask,
  onNewTask,
  onPickFilter,
}: {
  path: string;
  status: Status;
  filter: Filter;
  onOpenTask: (id: string) => void;
  onNewTask: () => void;
  onPickFilter: (f: Filter) => void;
}) {
  const { data: tasks, isLoading } = useTasks(path);
  const [query, setQuery] = useState("");
  const [label, setLabel] = useState("");
  const [assignee, setAssignee] = useState("");
  const [priority, setPriority] = useState("");
  const [views, setViews] = useState<SavedView[]>(() => loadViews(path));

  const states = status.states ?? [];
  const closed = new Set(status.closed ?? []);
  const q = query.trim().toLowerCase();

  // Facet option lists, derived from what's present.
  const labelOpts = [...new Set((tasks ?? []).flatMap((t) => t.labels ?? []))].sort();
  const assigneeOpts = [...new Set((tasks ?? []).map((t) => t.assignee).filter(Boolean))].sort() as string[];
  const prioOpts = ["urgent", "high", "medium", "low"].filter((p) =>
    (tasks ?? []).some((t) => t.priority === p),
  );
  const facetsActive = !!(label || assignee || priority || q);

  const visible = (tasks ?? [])
    .filter((t) => matches(t, filter, status))
    .filter((t) => !q || t.id.toLowerCase().includes(q) || t.title.toLowerCase().includes(q))
    .filter((t) => !label || (t.labels ?? []).includes(label))
    .filter((t) => !assignee || t.assignee === assignee)
    .filter((t) => !priority || t.priority === priority);

  const applyView = (v: SavedView) => {
    setQuery(v.query ?? "");
    setLabel(v.label ?? "");
    setAssignee(v.assignee ?? "");
    setPriority(v.priority ?? "");
    onPickFilter((v.filter as Filter) ?? "all");
  };
  const saveCurrentView = () => {
    const name = window.prompt("Save view as");
    if (!name?.trim()) return;
    setViews(addView(path, { name: name.trim(), filter, query, label, assignee, priority }));
  };
  const clearFacets = () => {
    setQuery("");
    setLabel("");
    setAssignee("");
    setPriority("");
  };

  const byStatus = new Map<string, Task[]>(states.map((s) => [s, []]));
  for (const t of visible) {
    if (!byStatus.has(t.status)) byStatus.set(t.status, []);
    byStatus.get(t.status)!.push(t);
  }
  const groups = [...byStatus.entries()].filter(([, list]) => list.length > 0);
  const isEmpty = !isLoading && visible.length === 0;

  // Keyboard navigation over the flat, display-ordered list (j/k/o/enter/c//).
  const flat = groups.flatMap(([, list]) => list);
  const [sel, setSel] = useState(0);
  const searchRef = useRef<HTMLInputElement>(null);
  const selId = flat[Math.min(sel, flat.length - 1)]?.id;

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement | null;
      if (t && (t.tagName === "INPUT" || t.tagName === "TEXTAREA" || t.isContentEditable)) return;
      // bail when any modal / palette / Select-listbox / menu popup is open
      if (document.querySelector('[role="dialog"],[role="listbox"],[role="menu"]')) return;
      if (e.key === "j") {
        e.preventDefault();
        setSel((s) => Math.min(flat.length - 1, s + 1));
      } else if (e.key === "k") {
        e.preventDefault();
        setSel((s) => Math.max(0, s - 1));
      } else if (e.key === "Enter" || e.key === "o") {
        if (selId) {
          e.preventDefault();
          onOpenTask(selId);
        }
      } else if (e.key === "c") {
        e.preventDefault();
        onNewTask();
      } else if (e.key === "/") {
        e.preventDefault();
        searchRef.current?.focus();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [flat.length, selId, onOpenTask, onNewTask]);

  useEffect(() => {
    if (selId) document.getElementById(`row-${selId}`)?.scrollIntoView({ block: "nearest" });
  }, [selId]);

  // Reset the keyboard cursor when the base view changes (avoids stale out-of-range index).
  useEffect(() => setSel(0), [filter]);

  return (
    <div className="flex h-full flex-col">
      <header className="flex h-11 shrink-0 items-center justify-between border-b px-4">
        <div className="flex items-center gap-1.5 text-[13px]">
          <span className="font-medium">{status.prefix}</span>
          <ChevronRight className="size-3.5 text-muted-foreground" />
          <span className="text-muted-foreground">{FILTER_LABEL[filter]}</span>
          {!isLoading && (
            <span className="ml-1 text-xs text-muted-foreground">{visible.length}</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="pointer-events-none absolute top-1/2 left-2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              ref={searchRef}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Filter…  ( / )"
              className="h-7 w-44 pl-7 text-xs"
            />
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm" className="h-7 gap-1 px-2 text-xs">
                <Bookmark className="size-3.5" /> Views
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel>Saved views</DropdownMenuLabel>
              {views.length === 0 && (
                <div className="px-2 py-1.5 text-xs text-muted-foreground">None yet</div>
              )}
              {views.map((v) => (
                <DropdownMenuItem key={v.name} onSelect={() => applyView(v)} className="justify-between">
                  <span className="truncate">{v.name}</span>
                  <button
                    aria-label={`Delete ${v.name}`}
                    onClick={(e) => {
                      e.stopPropagation();
                      setViews(removeView(path, v.name));
                    }}
                    className="text-muted-foreground hover:text-destructive"
                  >
                    <Trash2 className="size-3.5" />
                  </button>
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator />
              <DropdownMenuItem onSelect={saveCurrentView}>Save current view…</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <Button size="sm" className="h-7 gap-1 px-2.5 text-xs" onClick={onNewTask}>
            <Plus className="size-3.5" /> New task
          </Button>
        </div>
      </header>

      {(labelOpts.length > 0 || assigneeOpts.length > 0 || prioOpts.length > 0) && (
        <div className="flex shrink-0 items-center gap-2 border-b px-4 py-1.5">
          {prioOpts.length > 0 && (
            <Facet value={priority} onChange={setPriority} placeholder="Priority">
              {prioOpts.map((p) => (
                <SelectItem key={p} value={p}>
                  <span className="flex items-center gap-2">
                    <PriorityIcon priority={p} /> {priorityLabel(p)}
                  </span>
                </SelectItem>
              ))}
            </Facet>
          )}
          {labelOpts.length > 0 && (
            <Facet value={label} onChange={setLabel} placeholder="Label">
              {labelOpts.map((l) => (
                <SelectItem key={l} value={l}>
                  {l}
                </SelectItem>
              ))}
            </Facet>
          )}
          {assigneeOpts.length > 0 && (
            <Facet value={assignee} onChange={setAssignee} placeholder="Assignee">
              {assigneeOpts.map((a) => (
                <SelectItem key={a} value={a}>
                  {a}
                </SelectItem>
              ))}
            </Facet>
          )}
          {facetsActive && (
            <Button variant="ghost" size="sm" className="h-6 gap-1 px-2 text-xs" onClick={clearFacets}>
              <X className="size-3" /> Clear
            </Button>
          )}
        </div>
      )}

      <div className="flex-1 overflow-y-auto py-1">
        {isLoading ? (
          <div className="space-y-1.5 px-3 py-2">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-7 w-full" />
            ))}
          </div>
        ) : (tasks?.length ?? 0) === 0 ? (
          <Onboarding status={status} onNewTask={onNewTask} />
        ) : isEmpty ? (
          <EmptyState
            icon={Inbox}
            title={`Nothing in ${FILTER_LABEL[filter].toLowerCase()}`}
            message="Try another view, or create a task."
            action={{ label: "New task", icon: Plus, onClick: onNewTask }}
          />
        ) : (
          groups.map(([state, list]) => (
            <StatusSection
              key={state}
              path={path}
              state={state}
              tasks={list}
              status={status}
              defaultOpen={!closed.has(state)}
              onOpen={onOpenTask}
              selectedId={selId}
            />
          ))
        )}
      </div>
    </div>
  );
}

function StatusSection({
  path,
  state,
  tasks,
  status,
  defaultOpen,
  onOpen,
  selectedId,
}: {
  path: string;
  state: string;
  tasks: Task[];
  status: Status;
  defaultOpen: boolean;
  onOpen: (id: string) => void;
  selectedId?: string;
}) {
  const [open, setOpen] = useState(defaultOpen);
  useEffect(() => setOpen(defaultOpen), [defaultOpen]);
  return (
    <Collapsible open={open} onOpenChange={setOpen} className="mb-1">
      <CollapsibleTrigger asChild>
        <button className="flex h-9 w-full items-center gap-2 px-3 text-[13px] hover:bg-foreground/[0.02]">
          <ChevronRight
            className={cn(
              "size-3.5 shrink-0 text-muted-foreground transition-transform",
              open && "rotate-90",
            )}
          />
          <StatusIcon
            status={state}
            closed={status.closed}
            initial={status.initial}
            className="size-3.5"
          />
          <span className="font-medium">{statusLabel(state)}</span>
          <span className="text-xs text-muted-foreground">{tasks.length}</span>
        </button>
      </CollapsibleTrigger>
      <CollapsibleContent>
        {tasks.map((t) => (
          <TaskRow
            key={t.id}
            path={path}
            task={t}
            status={status}
            onOpen={onOpen}
            selected={t.id === selectedId}
          />
        ))}
      </CollapsibleContent>
    </Collapsible>
  );
}

