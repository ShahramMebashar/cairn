import { useState } from "react";
import {
  ArrowLeft,
  Check as CheckMark,
  ChevronRight,
  CornerLeftUp,
  Loader2,
  Play,
  Plus,
  UserPlus,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Assignee } from "@/components/Assignee";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useDefaultLayout } from "react-resizable-panels";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { StatusIcon } from "@/components/StatusIcon";
import { Markdown } from "@/components/Markdown";
import { MarkdownEditor } from "@/components/MarkdownEditor";
import { LogView } from "@/components/LogView";
import { Input } from "@/components/ui/input";
import { PriorityIcon, PRIORITIES, priorityLabel } from "@/components/PriorityIcon";
import {
  useAddNote,
  useAttest,
  useClaim,
  useRunChecks,
  useRuns,
  useTask,
  useTasks,
  useTransition,
  useUpdateTask,
} from "@/lib/queries";
import { cn, initials, statusLabel, timeAgo } from "@/lib/utils";
import type { Check, Run, Status, Task } from "@/lib/api";

export function TaskDetail({
  path,
  id,
  status,
  onBack,
  onOpenTask,
  onAddSubtask,
}: {
  path: string;
  id: string;
  status: Status;
  onBack: () => void;
  onOpenTask: (id: string) => void;
  onAddSubtask: (parentId: string) => void;
}) {
  const { data: task, isLoading } = useTask(path, id);
  const { data: runs } = useRuns(path, id);
  const { data: allTasks } = useTasks(path);
  const claim = useClaim(path);
  const transition = useTransition(path);
  const runChecks = useRunChecks(path);
  const attest = useAttest(path);
  const addNote = useAddNote(path);
  const update = useUpdateTask(path);
  const [note, setNote] = useState("");

  // Persist the resizable split (main / properties) to localStorage.
  const { defaultLayout, onLayoutChanged } = useDefaultLayout({
    id: "cairn-detail-layout",
    storage: localStorage,
    panelIds: ["detail-main", "detail-props"],
  });

  return (
    <div className="flex h-full flex-col">
      <header className="flex h-11 shrink-0 items-center gap-2 border-b px-3">
        <Button variant="ghost" size="icon" aria-label="Back" onClick={onBack}>
          <ArrowLeft />
        </Button>
        <div className="flex items-center gap-1.5 text-sm">
          <button onClick={onBack} className="font-medium hover:underline">
            {status.prefix}
          </button>
          <ChevronRight className="size-3.5 text-muted-foreground" />
          <span className="font-mono text-muted-foreground">{id}</span>
        </div>
      </header>

      {isLoading ? (
        <div className="mx-auto w-full max-w-2xl space-y-4 p-8">
          <Skeleton className="h-8 w-3/4" />
          <Skeleton className="h-24 w-full" />
        </div>
      ) : !task ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 text-center">
          <p className="text-sm text-muted-foreground">Task {id} not found.</p>
          <Button variant="outline" size="sm" onClick={onBack}>
            Back to tasks
          </Button>
        </div>
      ) : (
        <ResizablePanelGroup
          orientation="horizontal"
          defaultLayout={defaultLayout}
          onLayoutChanged={onLayoutChanged}
          className="min-h-0 flex-1"
        >
          <ResizablePanel id="detail-main" defaultSize="68%" className="min-w-0">
            {/* Main column */}
            <main className="h-full overflow-y-auto">
            <div className="mx-auto max-w-2xl px-8 py-8">
              {task.parent && (
                <button
                  onClick={() => onOpenTask(task.parent!)}
                  className="mb-2 flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
                >
                  <CornerLeftUp className="size-3.5" />
                  <span className="font-mono">{task.parent}</span>
                  <span className="truncate">
                    {(allTasks ?? []).find((t) => t.id === task.parent)?.title}
                  </span>
                </button>
              )}
              <div className="mb-2 flex items-center gap-2">
                <StatusIcon status={task.status} closed={status.closed} initial={status.initial} />
                <span className="font-mono text-xs text-muted-foreground">{task.id}</span>
              </div>
              <h1 className="text-2xl font-semibold tracking-tight">{task.title}</h1>

              <SubTasks
                parentId={task.id}
                all={allTasks ?? []}
                status={status}
                onOpenTask={onOpenTask}
                onAddSubtask={() => onAddSubtask(task.id)}
              />

              {task.body?.trim() && <Markdown className="mt-5">{task.body.trim()}</Markdown>}

              {/* Activity */}
              <section className="mt-10">
                <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Activity
                </h2>
                <ol className="mt-3 space-y-3.5">
                  {(task.provenance ?? []).map((p, i) => (
                    <li key={i} className="flex gap-2.5">
                      <Avatar className="mt-0.5 size-6 shrink-0">
                        <AvatarFallback className="text-[9px]">{initials(p.who)}</AvatarFallback>
                      </Avatar>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-baseline gap-1.5">
                          <span className="text-sm font-medium">{p.who}</span>
                          <span className="truncate text-xs text-muted-foreground">{p.did}</span>
                          <span
                            className="ml-auto shrink-0 text-xs text-muted-foreground tabular-nums"
                            title={new Date(p.at).toLocaleString()}
                          >
                            {timeAgo(p.at)}
                          </span>
                        </div>
                        {p.text && (
                          <div className="mt-1.5 rounded-lg border bg-muted/40 px-3 py-2">
                            <Markdown>{p.text}</Markdown>
                          </div>
                        )}
                      </div>
                    </li>
                  ))}
                </ol>

                <div className="mt-4 space-y-2">
                  <MarkdownEditor value={note} onChange={setNote} placeholder="Leave a note…" />
                  <div className="flex justify-end">
                    <Button
                      size="sm"
                      variant="secondary"
                      disabled={!note.trim() || addNote.isPending}
                      onClick={() =>
                        addNote.mutate(
                          { id: task.id, text: note.trim() },
                          { onSuccess: () => setNote("") },
                        )
                      }
                    >
                      {addNote.isPending && <Loader2 className="animate-spin" />}
                      Add note
                    </Button>
                  </div>
                </div>
              </section>
            </div>
            </main>
          </ResizablePanel>

          <ResizableHandle />

          {/* Properties panel */}
          <ResizablePanel
            id="detail-props"
            defaultSize="32%"
            minSize="24%"
            maxSize="50%"
            className="min-w-[260px]"
          >
            <aside className="h-full space-y-5 overflow-y-auto p-5">
            <Prop label="Status">
              <Select
                value={task.status}
                onValueChange={(to) => to !== task.status && transition.mutate({ id: task.id, to })}
              >
                <SelectTrigger className="h-8 w-full">
                  {transition.isPending ? (
                    <span className="flex items-center gap-2 text-sm">
                      <Loader2 className="size-3 animate-spin" /> Updating…
                    </span>
                  ) : (
                    <SelectValue />
                  )}
                </SelectTrigger>
                <SelectContent>
                  {(status.states ?? []).map((s) => (
                    <SelectItem key={s} value={s}>
                      <span className="flex items-center gap-2">
                        <StatusIcon status={s} closed={status.closed} initial={status.initial} />
                        {statusLabel(s)}
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Prop>

            <Prop label="Assignee">
              {task.assignee ? (
                <div className="flex items-center gap-2 text-sm">
                  <Assignee actor={task.assignee} />
                  {task.assignee}
                </div>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  className="w-full justify-start"
                  disabled={claim.isPending}
                  onClick={() => claim.mutate(task.id)}
                >
                  {claim.isPending ? <Loader2 className="animate-spin" /> : <UserPlus />}
                  Claim
                </Button>
              )}
            </Prop>

            <Prop label="Ready">
              {task.ready ? (
                <Badge className="bg-brand text-brand-foreground">Ready</Badge>
              ) : (
                <Badge variant="outline">Blocked by deps</Badge>
              )}
            </Prop>

            <Prop label="Priority">
              <Select
                value={task.priority || "none"}
                onValueChange={(v) =>
                  update.mutate({ id: task.id, fields: { priority: v === "none" ? "" : v } })
                }
              >
                <SelectTrigger className="h-8 w-full">
                  <span className="flex items-center gap-2">
                    <PriorityIcon priority={task.priority} />
                    {priorityLabel(task.priority)}
                  </span>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">
                    <span className="flex items-center gap-2">
                      <PriorityIcon /> No priority
                    </span>
                  </SelectItem>
                  {PRIORITIES.map((p) => (
                    <SelectItem key={p} value={p}>
                      <span className="flex items-center gap-2">
                        <PriorityIcon priority={p} /> {priorityLabel(p)}
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Prop>

            <Prop label="Labels">
              <LabelsEditor
                labels={task.labels ?? []}
                onChange={(labels) => update.mutate({ id: task.id, fields: { labels } })}
              />
            </Prop>

            <Prop label="Parent">
              <Select
                value={task.parent || "none"}
                onValueChange={(v) =>
                  update.mutate({ id: task.id, fields: { parent: v === "none" ? "" : v } })
                }
              >
                <SelectTrigger className="h-8 w-full">
                  <SelectValue placeholder="No parent" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">No parent</SelectItem>
                  {(allTasks ?? [])
                    .filter((t) => t.id !== task.id)
                    .map((t) => (
                      <SelectItem key={t.id} value={t.id}>
                        <span className="font-mono text-xs">{t.id}</span> {t.title}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </Prop>

            {task.deps && task.deps.length > 0 && (
              <Prop label="Depends on">
                <div className="flex flex-wrap gap-1.5">
                  {task.deps.map((d) => (
                    <Badge key={d} variant="outline" className="font-mono text-xs">
                      {d}
                    </Badge>
                  ))}
                </div>
              </Prop>
            )}

            {task.checks && task.checks.length > 0 && (
              <Prop
                label="Checks"
                action={
                  (task.checks ?? []).some((c) => c.cmd) ? (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 px-2 text-xs"
                      disabled={runChecks.isPending}
                      onClick={() => runChecks.mutate({ id: task.id })}
                    >
                      {runChecks.isPending ? (
                        <Loader2 className="animate-spin" />
                      ) : (
                        <Play className="size-3" />
                      )}
                      Run
                    </Button>
                  ) : null
                }
              >
                <div className="space-y-1.5">
                  {task.checks.map((c, i) => (
                    <CheckRow
                      key={i}
                      check={c}
                      // Latest run for this check, matched by command (runs are newest-first).
                      run={c.cmd ? (runs ?? []).find((r) => r.cmd === c.cmd) : undefined}
                      onAttest={(pass) => attest.mutate({ id: task.id, index: i, pass })}
                      attesting={attest.isPending}
                    />
                  ))}
                </div>
              </Prop>
            )}
            </aside>
          </ResizablePanel>
        </ResizablePanelGroup>
      )}
    </div>
  );
}

function Prop({
  label,
  action,
  children,
}: {
  label: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between">
        <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{label}</h3>
        {action}
      </div>
      {children}
    </div>
  );
}

function CheckRow({
  check,
  run,
  onAttest,
  attesting,
}: {
  check: Check;
  run?: Run;
  onAttest: (pass: boolean) => void;
  attesting: boolean;
}) {
  const color =
    check.result === "pass"
      ? "text-success"
      : check.result === "fail"
        ? "text-destructive"
        : "text-muted-foreground";
  const isManual = !check.cmd;
  const pending = (check.result ?? "pending") === "pending";

  const head = (
    <>
      <span className="min-w-0 truncate">{check.desc}</span>
      <span className={cn("shrink-0 text-xs font-medium", color)}>{check.result ?? "pending"}</span>
    </>
  );

  // Manual + pending: offer attest pass/fail (no command to run).
  if (isManual) {
    return (
      <div className="flex items-center justify-between gap-2 text-sm">
        {head}
        {pending && (
          <span className="flex shrink-0 items-center gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="size-6 text-success"
              aria-label="Attest pass"
              disabled={attesting}
              onClick={() => onAttest(true)}
            >
              {attesting ? <Loader2 className="animate-spin" /> : <CheckMark className="size-3.5" />}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="size-6 text-destructive"
              aria-label="Attest fail"
              disabled={attesting}
              onClick={() => onAttest(false)}
            >
              <X className="size-3.5" />
            </Button>
          </span>
        )}
      </div>
    );
  }

  // Command check with no run yet: plain row.
  if (!run) {
    return <div className="flex items-center justify-between gap-2 text-sm">{head}</div>;
  }

  // Command check with captured output: expandable log panel.
  return (
    <Collapsible>
      <CollapsibleTrigger className="group flex w-full items-center justify-between gap-2 text-left text-sm">
        <span className="flex min-w-0 items-center gap-1">
          <ChevronRight className="size-3.5 shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-90" />
          <span className="min-w-0 truncate">{check.desc}</span>
        </span>
        <span className={cn("shrink-0 text-xs font-medium", color)}>
          {check.result ?? "pending"}
        </span>
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-1.5">
        <LogView run={run} />
      </CollapsibleContent>
    </Collapsible>
  );
}

function SubTasks({
  parentId,
  all,
  status,
  onOpenTask,
  onAddSubtask,
}: {
  parentId: string;
  all: Task[];
  status: Status;
  onOpenTask: (id: string) => void;
  onAddSubtask: () => void;
}) {
  const children = all.filter((t) => t.parent === parentId);
  const closed = new Set(status.closed ?? []);
  const done = children.filter((c) => closed.has(c.status)).length;

  return (
    <section className="mt-6">
      <div className="flex items-center justify-between">
        <h2 className="flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          Sub-tasks
          {children.length > 0 && (
            <span className="tabular-nums">
              {done}/{children.length}
            </span>
          )}
        </h2>
        <Button variant="ghost" size="sm" className="h-6 px-2 text-xs" onClick={onAddSubtask}>
          <Plus className="size-3" /> Add
        </Button>
      </div>
      {children.length > 0 && (
        <div className="mt-2 divide-y rounded-lg border">
          {children.map((c) => (
            <button
              key={c.id}
              onClick={() => onOpenTask(c.id)}
              className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm hover:bg-foreground/[0.04]"
            >
              <StatusIcon status={c.status} closed={status.closed} initial={status.initial} className="size-3.5" />
              <span className="w-16 shrink-0 font-mono text-xs text-muted-foreground">{c.id}</span>
              <span className="truncate">{c.title}</span>
            </button>
          ))}
        </div>
      )}
    </section>
  );
}

function LabelsEditor({
  labels,
  onChange,
}: {
  labels: string[];
  onChange: (labels: string[]) => void;
}) {
  const [input, setInput] = useState("");
  const add = () => {
    const v = input.trim();
    if (v && !labels.includes(v)) onChange([...labels, v]);
    setInput("");
  };
  return (
    <div className="space-y-1.5">
      {labels.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {labels.map((l) => (
            <Badge key={l} variant="secondary" className="gap-1 pr-1 text-xs font-normal">
              {l}
              <button
                aria-label={`Remove ${l}`}
                onClick={() => onChange(labels.filter((x) => x !== l))}
                className="grid size-3.5 place-items-center rounded hover:bg-foreground/10"
              >
                <X className="size-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
      <Input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            add();
          }
        }}
        placeholder="Add label…"
        className="h-7 text-xs"
      />
    </div>
  );
}
