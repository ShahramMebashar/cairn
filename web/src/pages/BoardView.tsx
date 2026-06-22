import { useEffect, useMemo, useState } from "react";
import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  pointerWithin,
  rectIntersection,
  useDroppable,
  useSensor,
  useSensors,
  type CollisionDetection,
  type DragEndEvent,
  type DragOverEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  arrayMove,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { CheckCircle2, GitBranch, Plus, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { SelectItem } from "@/components/ui/select";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { StatusIcon } from "@/components/StatusIcon";
import { PriorityIcon, priorityLabel } from "@/components/PriorityIcon";
import { Facet } from "@/components/Facet";
import { useReorder, useTasks, useTransition } from "@/lib/queries";
import { effectiveRank } from "@/lib/filter";
import { cn, initials, statusLabel } from "@/lib/utils";
import type { Status, Task } from "@/lib/api";

export function BoardView({
  path,
  status,
  onOpenTask,
  onNewTask,
}: {
  path: string;
  status: Status;
  onOpenTask: (id: string) => void;
  onNewTask: () => void;
}) {
  const { data: tasks } = useTasks(path);
  const transition = useTransition(path);
  const reorder = useReorder(path);

  const [query, setQuery] = useState("");
  const [priority, setPriority] = useState("");
  const [label, setLabel] = useState("");
  const [assignee, setAssignee] = useState("");
  const [activeId, setActiveId] = useState<string | null>(null);
  const [cols, setCols] = useState<Record<string, string[]>>({});

  const states = useMemo(() => status.states ?? [], [status.states]);
  const byId = useMemo(() => new Map((tasks ?? []).map((t) => [t.id, t])), [tasks]);

  const labelOpts = [...new Set((tasks ?? []).flatMap((t) => t.labels ?? []))].sort();
  const assigneeOpts = [...new Set((tasks ?? []).map((t) => t.assignee).filter(Boolean))].sort() as string[];
  const prioOpts = ["urgent", "high", "medium", "low"].filter((p) =>
    (tasks ?? []).some((t) => t.priority === p),
  );

  const q = query.trim().toLowerCase();
  const visible = useMemo(
    () =>
      (tasks ?? [])
        .filter((t) => !q || t.id.toLowerCase().includes(q) || t.title.toLowerCase().includes(q))
        .filter((t) => !label || (t.labels ?? []).includes(label))
        .filter((t) => !assignee || t.assignee === assignee)
        .filter((t) => !priority || t.priority === priority),
    [tasks, q, label, assignee, priority],
  );

  // Rebuild columns only when the server data (or filters) actually change — NOT when the
  // drag ends. Listing activeId here would rebuild from stale pre-mutation data the instant
  // a drop happens, flashing the card back to its old slot. The optimistic cols hold until
  // the mutation settles and a fresh `tasks` arrives.
  useEffect(() => {
    if (activeId) return; // a drag is in flight — keep the optimistic order
    const next: Record<string, string[]> = {};
    for (const s of states) next[s] = [];
    for (const t of visible) (next[t.status] ??= []).push(t.id);
    for (const s of Object.keys(next))
      next[s].sort((a, b) => effectiveRank(byId.get(a)!) - effectiveRank(byId.get(b)!));
    setCols(next);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visible, states, byId]);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  // pointerWithin resolves an empty column under the cursor (closestCorners misses it);
  // fall back to rectIntersection for the keyboard sensor (no pointer coordinates).
  const collisionDetection: CollisionDetection = (args) => {
    const pointer = pointerWithin(args);
    return pointer.length ? pointer : rectIntersection(args);
  };

  const containerOf = (id: string): string | undefined =>
    id in cols ? id : Object.keys(cols).find((k) => cols[k].includes(id));

  const onDragStart = (e: DragStartEvent) => setActiveId(String(e.active.id));

  const onDragOver = (e: DragOverEvent) => {
    const { active, over } = e;
    if (!over) return;
    const from = containerOf(String(active.id));
    const to = containerOf(String(over.id));
    if (!from || !to || from === to) return;
    setCols((prev) => {
      const src = [...prev[from]];
      const dst = [...prev[to]];
      const overIdx = dst.indexOf(String(over.id));
      src.splice(src.indexOf(String(active.id)), 1);
      dst.splice(overIdx >= 0 ? overIdx : dst.length, 0, String(active.id));
      return { ...prev, [from]: src, [to]: dst };
    });
  };

  const onDragEnd = (e: DragEndEvent) => {
    const { active, over } = e;
    const id = String(active.id);
    setActiveId(null);
    if (!over) return;
    const to = containerOf(String(over.id));
    if (!to) return;

    let list = cols[to];
    const overIdx = list.indexOf(String(over.id));
    const curIdx = list.indexOf(id);
    if (overIdx >= 0 && curIdx >= 0 && overIdx !== curIdx) {
      list = arrayMove(list, curIdx, overIdx);
      setCols((prev) => ({ ...prev, [to]: list }));
    }

    const idx = list.indexOf(id);
    const prevT = idx > 0 ? byId.get(list[idx - 1]) : undefined;
    const nextT = idx < list.length - 1 ? byId.get(list[idx + 1]) : undefined;
    const lo = prevT ? effectiveRank(prevT) : undefined;
    const hi = nextT ? effectiveRank(nextT) : undefined;
    const rank =
      lo !== undefined && hi !== undefined ? (lo + hi) / 2 : lo !== undefined ? lo + 1 : hi !== undefined ? hi - 1 : 1;

    const task = byId.get(id);
    if (task && task.status !== to) {
      transition.mutate(
        { id, to },
        { onSuccess: () => reorder.mutate({ id, rank }) },
      );
    } else {
      reorder.mutate({ id, rank });
    }
  };

  const activeTask = activeId ? byId.get(activeId) : undefined;

  return (
    <div className="flex h-full flex-col">
      <header className="flex h-11 shrink-0 items-center justify-between border-b px-4">
        <div className="flex items-center gap-2 text-[13px]">
          <span className="font-medium">{status.prefix}</span>
          <span className="text-muted-foreground">Board</span>
        </div>
        <div className="flex items-center gap-2">
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
          <div className="relative">
            <Search className="pointer-events-none absolute top-1/2 left-2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Filter…"
              className="h-7 w-40 pl-7 text-xs"
            />
          </div>
          <Button size="sm" className="h-7 gap-1 px-2.5 text-xs" onClick={onNewTask}>
            <Plus className="size-3.5" /> New task
          </Button>
        </div>
      </header>

      <DndContext
        sensors={sensors}
        collisionDetection={collisionDetection}
        onDragStart={onDragStart}
        onDragOver={onDragOver}
        onDragEnd={onDragEnd}
        onDragCancel={() => setActiveId(null)}
      >
        <div className="flex flex-1 gap-3 overflow-x-auto p-3">
          {states.map((s) => (
            <Column
              key={s}
              status={s}
              info={status}
              cardIds={cols[s] ?? []}
              byId={byId}
              onOpenTask={onOpenTask}
            />
          ))}
        </div>
        <DragOverlay dropAnimation={null}>
          {activeTask ? <Card task={activeTask} dragging /> : null}
        </DragOverlay>
      </DndContext>
    </div>
  );
}

function Column({
  status,
  info,
  cardIds,
  byId,
  onOpenTask,
}: {
  status: string;
  info: Status;
  cardIds: string[];
  byId: Map<string, Task>;
  onOpenTask: (id: string) => void;
}) {
  const { setNodeRef, isOver } = useDroppable({ id: status });
  return (
    <div className="flex w-72 shrink-0 flex-col">
      <div className="mb-2 flex items-center gap-2 px-1 text-[13px]">
        <StatusIcon status={status} closed={info.closed} initial={info.initial} className="size-3.5" />
        <span className="font-medium">{statusLabel(status)}</span>
        <span className="text-xs text-muted-foreground">{cardIds.length}</span>
      </div>
      <SortableContext items={cardIds} strategy={verticalListSortingStrategy}>
        <div
          ref={setNodeRef}
          className={cn(
            "flex min-h-24 flex-1 flex-col gap-2 rounded-lg p-1 transition-colors",
            isOver && "bg-foreground/[0.04]",
          )}
        >
          {cardIds.map((id) => {
            const t = byId.get(id);
            return t ? <SortableCard key={id} task={t} onOpenTask={onOpenTask} /> : null;
          })}
        </div>
      </SortableContext>
    </div>
  );
}

function SortableCard({
  task,
  onOpenTask,
}: {
  task: Task;
  onOpenTask: (id: string) => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: task.id,
  });
  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Translate.toString(transform), transition }}
      {...attributes}
      {...listeners}
      onClick={() => !isDragging && onOpenTask(task.id)}
      role="button"
      tabIndex={0}
    >
      {isDragging ? <DropIndicator /> : <Card task={task} />}
    </div>
  );
}

// DropIndicator marks where the dragged card will land: a blue rule with a centered ring.
function DropIndicator() {
  return (
    <div className="relative my-1 h-[3px] rounded-full bg-brand">
      <div className="absolute top-1/2 left-1/2 size-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-brand bg-panel" />
    </div>
  );
}

function Card({ task, dragging }: { task: Task; dragging?: boolean }) {
  const checks = task.checks ?? [];
  const passed = checks.filter((c) => c.result === "pass").length;
  const allPass = checks.length > 0 && passed === checks.length;
  return (
    <div
      className={cn(
        "cursor-pointer rounded-lg border bg-panel p-2.5 text-left shadow-xs transition-shadow hover:border-foreground/20",
        dragging && "rotate-2 shadow-md",
      )}
    >
      <div className="mb-1 flex items-center gap-1.5">
        <PriorityIcon priority={task.priority} />
        <span className="font-mono text-[11px] text-muted-foreground">{task.id}</span>
      </div>
      <div className="text-[13px] leading-snug">{task.title}</div>
      {(task.labels?.length || checks.length || task.deps?.length || task.assignee) && (
        <div className="mt-2 flex items-center gap-2">
          {task.labels?.slice(0, 2).map((l) => (
            <span key={l} className="rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
              {l}
            </span>
          ))}
          <span className="flex-1" />
          {task.deps && task.deps.length > 0 && (
            <span className="flex items-center gap-0.5 text-[10px] text-muted-foreground">
              <GitBranch className="size-3" />
              {task.deps.length}
            </span>
          )}
          {checks.length > 0 && (
            <span className={cn("flex items-center gap-0.5 text-[10px]", allPass ? "text-success" : "text-muted-foreground")}>
              <CheckCircle2 className="size-3" />
              {passed}/{checks.length}
            </span>
          )}
          {task.assignee && (
            <Avatar className="size-5">
              <AvatarFallback className="text-[9px]">{initials(task.assignee)}</AvatarFallback>
            </Avatar>
          )}
        </div>
      )}
    </div>
  );
}
