import { useState } from "react";
import {
  ArrowLeft,
  AlertTriangle,
  Bot,
  Check as CheckMark,
  ChevronRight,
  Circle,
  CircleCheck,
  CircleX,
  CornerLeftUp,
  FileText,
  GitBranch,
  GitCommit,
  Link2,
  Loader2,
  MoreHorizontal,
  Pencil,
  Play,
  Plus,
  Trash2,
  UserPlus,
  X,
} from "lucide-react";
import { toast } from "sonner";
import { agentPromptForTask, taskDeepLink } from "@/lib/connect";
import { currentActor } from "@/lib/identity";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Assignee } from "@/components/Assignee";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ConfirmDeleteDialog } from "@/components/ConfirmDeleteDialog";
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
import { SessionStatus } from "@/components/SessionStatus";
import { SessionTimeline } from "@/components/SessionTimeline";
import {
  useAddNote,
  useAttest,
  useClaim,
  useDeleteNote,
  useDeleteTask,
  useEditNote,
  useRunChecks,
  useRuns,
  useTask,
  useTaskGitContext,
  useTaskSessions,
  useTasks,
  useTransition,
  useUpdateTask,
} from "@/lib/queries";
import { cn, initials, statusLabel, timeAgo } from "@/lib/utils";
import type { ChangedFile, Check, GitCommit as GitCommitData, Run, SessionGitContext, Status, Task } from "@/lib/api";

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
  const { data: sessions, isLoading: sessionsLoading } = useTaskSessions(path, id);
  const { data: gitContexts, isLoading: gitContextLoading } = useTaskGitContext(path, id);
  const claim = useClaim(path);
  const transition = useTransition(path);
  const runChecks = useRunChecks(path);
  const attest = useAttest(path);
  const addNote = useAddNote(path);
  const update = useUpdateTask(path);
  const deleteTask = useDeleteTask(path);
  const editNote = useEditNote(path);
  const deleteNote = useDeleteNote(path);
  const [note, setNote] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);

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
        {task && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="ml-auto" aria-label="Task actions">
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onSelect={() => {
                  navigator.clipboard.writeText(taskDeepLink(path, task.id));
                  toast.success("Link copied");
                }}
              >
                <Link2 /> Copy link
              </DropdownMenuItem>
              <DropdownMenuItem
                onSelect={() => {
                  navigator.clipboard.writeText(
                    agentPromptForTask(task, path, currentActor(), window.location.origin),
                  );
                  toast.success("Agent prompt copied");
                }}
              >
                <Bot /> Copy as agent prompt
              </DropdownMenuItem>
              <DropdownMenuItem variant="destructive" onSelect={() => setConfirmDelete(true)}>
                <Trash2 /> Delete task
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </header>

      {task && (
        <ConfirmDeleteDialog
          open={confirmDelete}
          onOpenChange={setConfirmDelete}
          title={`Delete ${task.id}?`}
          description={
            <>
              This permanently deletes <span className="font-medium">{task.title}</span>. Tasks with
              sub-tasks or dependents can't be deleted until those are removed.
            </>
          }
          confirmLabel="Delete task"
          pending={deleteTask.isPending}
          onConfirm={() => deleteTask.mutate(task.id, { onSuccess: onBack })}
        />
      )}

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
              <EditableTitle
                title={task.title}
                saving={update.isPending}
                onSave={(title) => update.mutate({ id: task.id, fields: { title } })}
              />

              <SubTasks
                path={path}
                parentId={task.id}
                all={allTasks ?? []}
                status={status}
                onOpenTask={onOpenTask}
                onAddSubtask={() => onAddSubtask(task.id)}
              />

              <EditableBody
                body={task.body ?? ""}
                saving={update.isPending}
                onSave={(body) => update.mutate({ id: task.id, fields: { body } })}
              />

              <SessionTimeline
                sessions={sessions ?? []}
                executionState={task.executionState}
                loading={sessionsLoading}
              />

              <CodeContextPanel sessions={gitContexts ?? []} loading={gitContextLoading} />

              {/* Activity */}
              <section className="mt-10">
                <h2 className="text-xs font-medium text-muted-foreground">
                  Activity
                </h2>
                <ol className="mt-3 space-y-3.5">
                  {(task.provenance ?? []).map((p, i) => (
                    <ActivityEntry
                      key={p.id || i}
                      entry={p}
                      onEdit={(text) =>
                        editNote.mutate({ id: task.id, text, note: p.id, index: i })
                      }
                      onDelete={() => deleteNote.mutate({ id: task.id, note: p.id, index: i })}
                      saving={editNote.isPending || deleteNote.isPending}
                    />
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
                      <Loader2 className="size-3 animate-spin" />
                      {transition.variables &&
                      ((status.closed ?? []).includes(transition.variables.to) ||
                        status.review === transition.variables.to)
                        ? "Running checks…"
                        : "Updating…"}
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

            {task.executionState && (
              <Prop label="Execution">
                <SessionStatus state={task.executionState} />
              </Prop>
            )}

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

            <ChecksSection
              checks={task.checks ?? []}
              runs={runs ?? []}
              running={runChecks.isPending}
              saving={update.isPending}
              onRun={() => runChecks.mutate({ id: task.id })}
              onAttest={(i, pass) => attest.mutate({ id: task.id, index: i, pass })}
              attesting={attest.isPending}
              onSave={(checks) => update.mutate({ id: task.id, fields: { checks } })}
            />
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
        <h3 className="text-xs font-medium text-muted-foreground">{label}</h3>
        {action}
      </div>
      {children}
    </div>
  );
}

function CodeContextPanel({
  sessions,
  loading,
}: {
  sessions: SessionGitContext[];
  loading: boolean;
}) {
  if (loading) {
    return (
      <section className="mt-8">
        <h2 className="text-xs font-medium text-muted-foreground">Code context</h2>
        <div className="mt-3 space-y-2">
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-14 w-full" />
        </div>
      </section>
    );
  }
  if (sessions.length === 0) return null;

  return (
    <section className="mt-8">
      <h2 className="text-xs font-medium text-muted-foreground">Code context</h2>
      <div className="mt-3 space-y-2">
        {sessions.map(({ session, context }) => (
          <article key={session.id} className="rounded-lg border bg-background px-3.5 py-3">
            <div className="flex min-w-0 items-center gap-2">
              <GitBranch className="size-3.5 text-muted-foreground" />
              <span className="min-w-0 flex-1 truncate text-sm font-medium">
                {context.branch || session.branch || "Detached HEAD"}
              </span>
              <Badge variant="outline" className="h-5 px-1.5 text-[10px] font-normal">
                {session.status}
              </Badge>
            </div>

            <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
              {context.headStarted && <Ref label="start" value={context.headStarted} />}
              {context.headFinished && <Ref label="finish" value={context.headFinished} />}
              {!context.headFinished && context.currentHead && <Ref label="current" value={context.currentHead} />}
            </div>

            {!context.available ? (
              <WarningList messages={[context.error || "Git context is unavailable."]} />
            ) : (
              <>
                {(context.warnings ?? []).length > 0 && (
                  <WarningList messages={(context.warnings ?? []).map((w) => w.message)} />
                )}
                <FileList title="Files changed" files={context.filesChanged ?? []} />
                <CommitList commits={context.commits ?? []} />
                <FileList title="Uncommitted" files={context.uncommitted ?? []} mutedEmpty />
              </>
            )}
          </article>
        ))}
      </div>
    </section>
  );
}

function Ref({ label, value }: { label: string; value: string }) {
  return (
    <span className="flex items-center gap-1">
      <span>{label}</span>
      <span className="font-mono text-foreground">{shortSha(value)}</span>
    </span>
  );
}

function WarningList({ messages }: { messages: string[] }) {
  return (
    <div className="mt-2 space-y-1.5">
      {messages.map((message) => (
        <div
          key={message}
          className="flex gap-2 rounded-md border bg-muted/40 px-2.5 py-2 text-xs text-muted-foreground"
        >
          <AlertTriangle className="mt-0.5 size-3.5 shrink-0 text-muted-foreground" />
          <span>{message}</span>
        </div>
      ))}
    </div>
  );
}

function FileList({
  title,
  files,
  mutedEmpty = false,
}: {
  title: string;
  files: ChangedFile[];
  mutedEmpty?: boolean;
}) {
  if (files.length === 0) {
    if (mutedEmpty) return null;
    return (
      <div className="mt-3 text-xs text-muted-foreground">
        <span className="font-medium">{title}</span>: none
      </div>
    );
  }
  return (
    <div className="mt-3">
      <div className="mb-1.5 flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
        <FileText className="size-3.5" />
        {title}
      </div>
      <div className="space-y-1">
        {files.slice(0, 8).map((file) => (
          <div
            key={`${file.status}:${file.oldPath ?? ""}:${file.path}`}
            className="flex min-w-0 items-center gap-2 text-xs"
          >
            <Badge variant="outline" className="h-5 w-8 justify-center px-0 font-mono text-[10px]">
              {file.status}
            </Badge>
            <span className="min-w-0 truncate font-mono text-muted-foreground">
              {file.oldPath ? `${file.oldPath} -> ${file.path}` : file.path}
            </span>
          </div>
        ))}
        {files.length > 8 && (
          <div className="text-xs text-muted-foreground">+{files.length - 8} more</div>
        )}
      </div>
    </div>
  );
}

function CommitList({ commits }: { commits: GitCommitData[] }) {
  if (commits.length === 0) return null;
  return (
    <div className="mt-3">
      <div className="mb-1.5 flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
        <GitCommit className="size-3.5" />
        Commits
      </div>
      <div className="space-y-1">
        {commits.slice(0, 6).map((commit) => (
          <div key={commit.hash} className="flex min-w-0 items-center gap-2 text-xs">
            <span className="font-mono text-muted-foreground">{shortSha(commit.hash)}</span>
            <span className="min-w-0 truncate">{commit.subject}</span>
          </div>
        ))}
        {commits.length > 6 && (
          <div className="text-xs text-muted-foreground">+{commits.length - 6} more</div>
        )}
      </div>
    </div>
  );
}

function shortSha(value: string) {
  return value.length > 7 ? value.slice(0, 7) : value;
}

// checkStatus maps a check's result (and whether a run is in flight) to its icon and pill
// styling. The left-edge icon makes the column scannable; the pill labels the state.
function checkStatus(result: string | undefined, running: boolean) {
  if (running) {
    return { Icon: Loader2, label: "running", icon: "animate-spin text-muted-foreground", pill: "bg-muted text-muted-foreground" };
  }
  switch (result) {
    case "pass":
      return { Icon: CircleCheck, label: "pass", icon: "text-success", pill: "bg-success/10 text-success" };
    case "fail":
      return { Icon: CircleX, label: "fail", icon: "text-destructive", pill: "bg-destructive/10 text-destructive" };
    default:
      return { Icon: Circle, label: "pending", icon: "text-muted-foreground/50", pill: "bg-muted text-muted-foreground" };
  }
}

function StatusPill({ className, children }: { className: string; children: React.ReactNode }) {
  return (
    <span className={cn("shrink-0 rounded-full px-1.5 py-0.5 text-[11px] font-medium", className)}>
      {children}
    </span>
  );
}

function CheckRow({
  check,
  run,
  running,
  onAttest,
  attesting,
}: {
  check: Check;
  run?: Run;
  running: boolean;
  onAttest: (pass: boolean) => void;
  attesting: boolean;
}) {
  const isManual = !check.cmd;
  const pending = (check.result ?? "pending") === "pending";
  const meta = checkStatus(check.result, running && !isManual);
  const expandable = !isManual && !!run;

  const lead = (
    <>
      <meta.Icon className={cn("size-4 shrink-0", meta.icon)} />
      <span className="min-w-0 flex-1 truncate text-sm">{check.desc}</span>
    </>
  );

  // Manual + pending: offer inline attest pass/fail (no command to run).
  if (isManual && pending) {
    return (
      <div className="flex items-center gap-2 px-3 py-2">
        {lead}
        <span className="flex shrink-0 items-center gap-0.5">
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
      </div>
    );
  }

  // Command check with captured output: the whole row toggles an inline log panel.
  if (expandable) {
    return (
      <Collapsible>
        <CollapsibleTrigger className="group flex w-full items-center gap-2 px-3 py-2 text-left hover:bg-foreground/[0.03]">
          {lead}
          <StatusPill className={meta.pill}>{meta.label}</StatusPill>
          <ChevronRight className="size-3.5 shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-90" />
        </CollapsibleTrigger>
        <CollapsibleContent>
          <div className="border-t bg-muted/30 px-3 py-2">
            <LogView run={run} />
          </div>
        </CollapsibleContent>
      </Collapsible>
    );
  }

  // Manual-passed/failed, or a command check not yet run: a plain status row.
  return (
    <div className="flex items-center gap-2 px-3 py-2">
      {lead}
      <StatusPill className={meta.pill}>{meta.label}</StatusPill>
    </div>
  );
}

function SubTasks({
  path,
  parentId,
  all,
  status,
  onOpenTask,
  onAddSubtask,
}: {
  path: string;
  parentId: string;
  all: Task[];
  status: Status;
  onOpenTask: (id: string) => void;
  onAddSubtask: () => void;
}) {
  const children = all.filter((t) => t.parent === parentId);
  const closed = new Set(status.closed ?? []);
  const done = children.filter((c) => closed.has(c.status)).length;
  const deleteTask = useDeleteTask(path);
  const [pendingDelete, setPendingDelete] = useState<Task | null>(null);

  return (
    <section className="mt-6">
      <div className="flex items-center justify-between">
        <h2 className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
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
            <div
              key={c.id}
              className="group/sub flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-foreground/[0.04]"
            >
              <button
                onClick={() => onOpenTask(c.id)}
                className="flex min-w-0 flex-1 items-center gap-2 text-left"
              >
                <StatusIcon status={c.status} closed={status.closed} initial={status.initial} className="size-3.5" />
                <span className="w-16 shrink-0 font-mono text-xs text-muted-foreground">{c.id}</span>
                <span className="truncate">{c.title}</span>
              </button>
              <Button
                variant="ghost"
                size="icon"
                className="size-6 shrink-0 text-destructive opacity-0 transition-opacity group-hover/sub:opacity-100"
                aria-label={`Delete ${c.id}`}
                onClick={() => setPendingDelete(c)}
              >
                <Trash2 className="size-3.5" />
              </Button>
            </div>
          ))}
        </div>
      )}
      <ConfirmDeleteDialog
        open={!!pendingDelete}
        onOpenChange={(o) => !o && setPendingDelete(null)}
        title={pendingDelete ? `Delete ${pendingDelete.id}?` : ""}
        description={
          <>
            This permanently deletes{" "}
            <span className="font-medium">{pendingDelete?.title}</span>. Sub-tasks with their own
            children or dependents can't be deleted until those are removed.
          </>
        }
        confirmLabel="Delete sub-task"
        pending={deleteTask.isPending}
        onConfirm={() =>
          pendingDelete &&
          deleteTask.mutate(pendingDelete.id, { onSuccess: () => setPendingDelete(null) })
        }
      />
    </section>
  );
}

// EditableTitle shows the task title as a heading; clicking it (or the pencil) swaps in an
// input. Enter/blur saves a non-empty change; Escape cancels.
function EditableTitle({
  title,
  saving,
  onSave,
}: {
  title: string;
  saving: boolean;
  onSave: (title: string) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState(title);

  if (editing) {
    const commit = () => {
      const v = value.trim();
      if (v && v !== title) onSave(v);
      setEditing(false);
    };
    return (
      <Input
        autoFocus
        value={value}
        disabled={saving}
        onChange={(e) => setValue(e.target.value)}
        onBlur={commit}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            commit();
          } else if (e.key === "Escape") {
            setValue(title);
            setEditing(false);
          }
        }}
        className="!text-2xl h-auto py-1 font-semibold tracking-tight"
      />
    );
  }

  return (
    <h1
      className="group/title -mx-1 flex cursor-text items-start gap-2 rounded px-1 text-2xl font-semibold tracking-tight hover:bg-foreground/[0.03]"
      onClick={() => {
        setValue(title);
        setEditing(true);
      }}
    >
      <span className="min-w-0 flex-1">{title}</span>
      <Pencil className="mt-1.5 size-3.5 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover/title:opacity-100" />
    </h1>
  );
}

// EditableBody renders the markdown body with a hover "Edit" affordance; an empty body shows
// an "Add description" button. Editing swaps in the shared MarkdownEditor with save/cancel.
function EditableBody({
  body,
  saving,
  onSave,
}: {
  body: string;
  saving: boolean;
  onSave: (body: string) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState("");
  const trimmed = body.trim();

  if (editing) {
    return (
      <div className="mt-5 space-y-2">
        <MarkdownEditor value={draft} onChange={setDraft} placeholder="Describe the task…" />
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="sm" onClick={() => setEditing(false)}>
            Cancel
          </Button>
          <Button
            size="sm"
            variant="secondary"
            disabled={saving}
            onClick={() => {
              onSave(draft.trim());
              setEditing(false);
            }}
          >
            {saving && <Loader2 className="animate-spin" />}
            Save
          </Button>
        </div>
      </div>
    );
  }

  const startEdit = () => {
    setDraft(body);
    setEditing(true);
  };

  if (!trimmed) {
    return (
      <Button variant="ghost" size="sm" className="mt-4 text-muted-foreground" onClick={startEdit}>
        <Plus className="size-3.5" /> Add description
      </Button>
    );
  }

  return (
    <div className="group/body relative mt-5">
      <Button
        variant="ghost"
        size="icon"
        aria-label="Edit description"
        className="absolute right-0 top-0 size-6 opacity-0 transition-opacity group-hover/body:opacity-100"
        onClick={startEdit}
      >
        <Pencil className="size-3.5" />
      </Button>
      <Markdown>{trimmed}</Markdown>
    </div>
  );
}

// ChecksSection shows the task's checks (with run/attest) and toggles into a ChecksEditor for
// adding, removing, or modifying them. Editing replaces the whole list in one update; retained
// checks carry their result forward, new checks default to pending server-side.
function ChecksSection({
  checks,
  runs,
  running,
  saving,
  onRun,
  onAttest,
  attesting,
  onSave,
}: {
  checks: Check[];
  runs: Run[];
  running: boolean;
  saving: boolean;
  onRun: () => void;
  onAttest: (index: number, pass: boolean) => void;
  attesting: boolean;
  onSave: (checks: Check[]) => void;
}) {
  const [editing, setEditing] = useState(false);

  if (editing) {
    return (
      <ChecksEditor
        checks={checks}
        saving={saving}
        onCancel={() => setEditing(false)}
        onSave={(next) => {
          onSave(next);
          setEditing(false);
        }}
      />
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h3 className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
          Checks
          {checks.length > 0 && (
            <span className={cn("tabular-nums", checks.every((c) => c.result === "pass") && "text-success")}>
              {checks.filter((c) => c.result === "pass").length}/{checks.length}
            </span>
          )}
        </h3>
        <div className="flex items-center gap-1">
          {checks.some((c) => c.cmd) && (
            <Button
              variant="outline"
              size="sm"
              className="h-6 gap-1 px-2 text-xs"
              disabled={running}
              onClick={onRun}
            >
              {running ? <Loader2 className="size-3 animate-spin" /> : <Play className="size-3" />}
              Run
            </Button>
          )}
          <Button
            variant="ghost"
            size="sm"
            className="h-6 gap-1 px-2 text-xs"
            onClick={() => setEditing(true)}
          >
            <Pencil className="size-3" />
            Edit
          </Button>
        </div>
      </div>
      {checks.length > 0 ? (
        <div className="divide-y overflow-hidden rounded-lg border">
          {checks.map((c, i) => (
            <CheckRow
              key={i}
              check={c}
              run={c.cmd ? runs.find((r) => r.cmd === c.cmd) : undefined}
              running={running}
              onAttest={(pass) => onAttest(i, pass)}
              attesting={attesting}
            />
          ))}
        </div>
      ) : (
        <p className="text-xs text-muted-foreground">No checks.</p>
      )}
    </div>
  );
}

// ChecksEditor edits the checks list in a local draft: each row has a description, an optional
// command (blank = a manual/attested check), and a remove button. Save emits the whole list.
function ChecksEditor({
  checks,
  saving,
  onCancel,
  onSave,
}: {
  checks: Check[];
  saving: boolean;
  onCancel: () => void;
  onSave: (checks: Check[]) => void;
}) {
  const [draft, setDraft] = useState<Check[]>(checks.map((c) => ({ ...c })));

  const setRow = (i: number, patch: Partial<Check>) =>
    setDraft((d) => d.map((c, j) => (j === i ? { ...c, ...patch } : c)));
  const removeRow = (i: number) => setDraft((d) => d.filter((_, j) => j !== i));
  const addRow = () => setDraft((d) => [...d, { desc: "", cmd: "" }]);

  const save = () => {
    // Drop blank-description rows; a command implies a non-manual check, else it's manual.
    const cleaned = draft
      .map((c) => ({ ...c, desc: c.desc.trim(), cmd: (c.cmd ?? "").trim() }))
      .filter((c) => c.desc)
      .map((c) => ({ ...c, type: c.cmd ? "" : "manual" }));
    onSave(cleaned);
  };

  return (
    <div className="space-y-2">
      <h3 className="text-xs font-medium text-muted-foreground">Checks</h3>
      <div className="space-y-2 rounded-lg border p-2">
        {draft.length === 0 && (
          <p className="px-1 py-2 text-xs text-muted-foreground">No checks. Add one below.</p>
        )}
        {draft.map((c, i) => (
          <div key={i} className="space-y-1.5 rounded-md border bg-muted/30 p-2">
            <div className="flex items-center gap-1.5">
              <Input
                value={c.desc}
                placeholder="What it verifies…"
                onChange={(e) => setRow(i, { desc: e.target.value })}
                className="h-7 text-xs"
              />
              <Button
                variant="ghost"
                size="icon"
                className="size-7 shrink-0 text-destructive"
                aria-label="Remove check"
                onClick={() => removeRow(i)}
              >
                <X className="size-3.5" />
              </Button>
            </div>
            <Input
              value={c.cmd ?? ""}
              placeholder="Command (blank = manual check)"
              onChange={(e) => setRow(i, { cmd: e.target.value })}
              className="h-7 font-mono text-xs"
            />
          </div>
        ))}
        <Button variant="ghost" size="sm" className="h-7 w-full justify-start text-xs" onClick={addRow}>
          <Plus className="size-3" /> Add check
        </Button>
      </div>
      <div className="flex justify-end gap-2">
        <Button variant="ghost" size="sm" onClick={onCancel}>
          Cancel
        </Button>
        <Button size="sm" variant="secondary" disabled={saving} onClick={save}>
          {saving && <Loader2 className="animate-spin" />}
          Save
        </Button>
      </div>
    </div>
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

type ProvEntry = { id?: string; who: string; did: string; at: string; text?: string; editedAt?: string };

// ActivityEntry renders one provenance row. Entries that carry a note body are collapsible
// (collapsed by default, with a one-line preview) so long notes don't flood the log. Note
// entries (did === "note") can be edited or deleted inline; system entries are read-only.
function ActivityEntry({
  entry,
  onEdit,
  onDelete,
  saving,
}: {
  entry: ProvEntry;
  onEdit: (text: string) => void;
  onDelete: () => void;
  saving: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);
  const note = entry.text?.trim();
  const preview = note ? (note.split("\n").find((l) => l.trim()) ?? "") : "";
  const isNote = entry.did === "note";

  const startEdit = () => {
    setDraft(entry.text ?? "");
    setEditing(true);
    setOpen(true);
  };
  const save = () => {
    const v = draft.trim();
    if (v) onEdit(v);
    setEditing(false);
  };

  return (
    <li className="group flex gap-2">
      {/* chevron gutter — reserved on every row so avatars/names stay aligned */}
      <button
        aria-label={open ? "Collapse note" : "Expand note"}
        onClick={note ? () => setOpen((o) => !o) : undefined}
        className={cn(
          "flex h-6 w-3.5 shrink-0 items-center justify-center",
          !note && "pointer-events-none",
        )}
      >
        {note && (
          <ChevronRight
            className={cn(
              "size-3.5 text-muted-foreground transition-transform",
              open && "rotate-90",
            )}
          />
        )}
      </button>
      <Avatar className="size-6 shrink-0">
        <AvatarFallback className="text-[9px]">{initials(entry.who)}</AvatarFallback>
      </Avatar>
      <div className="min-w-0 flex-1">
        <div
          className={cn(
            "flex min-h-6 items-center gap-1.5",
            note && !editing && "cursor-pointer select-none",
          )}
          onClick={note && !editing ? () => setOpen((o) => !o) : undefined}
        >
          <span className="shrink-0 text-sm font-medium">{entry.who}</span>
          <span className="shrink-0 text-xs text-muted-foreground">{entry.did}</span>
          {entry.editedAt && <span className="shrink-0 text-xs text-muted-foreground/70">(edited)</span>}
          {note && !open && (
            <span className="min-w-0 flex-1 truncate text-xs text-muted-foreground/80">{preview}</span>
          )}
          {isNote && !editing && (
            <span className="ml-auto flex shrink-0 items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
              <Button
                variant="ghost"
                size="icon"
                className="size-6"
                aria-label="Edit note"
                disabled={saving}
                onClick={(e) => {
                  e.stopPropagation();
                  startEdit();
                }}
              >
                <Pencil className="size-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="size-6 text-destructive"
                aria-label="Delete note"
                disabled={saving}
                onClick={(e) => {
                  e.stopPropagation();
                  setConfirmDelete(true);
                }}
              >
                <Trash2 className="size-3.5" />
              </Button>
            </span>
          )}
          <span
            className={cn(
              "shrink-0 text-xs text-muted-foreground tabular-nums",
              isNote && !editing ? "ml-1.5 group-hover:ml-0" : "ml-auto",
            )}
            title={new Date(entry.at).toLocaleString()}
          >
            {timeAgo(entry.at)}
          </span>
        </div>
        {editing ? (
          <div className="mt-1.5 space-y-2">
            <MarkdownEditor value={draft} onChange={setDraft} placeholder="Edit note…" />
            <div className="flex justify-end gap-2">
              <Button variant="ghost" size="sm" onClick={() => setEditing(false)}>
                Cancel
              </Button>
              <Button size="sm" variant="secondary" disabled={!draft.trim() || saving} onClick={save}>
                {saving && <Loader2 className="animate-spin" />}
                Save
              </Button>
            </div>
          </div>
        ) : (
          note &&
          open && (
            <div className="mt-1.5 rounded-lg border bg-muted/40 px-3 py-2">
              <Markdown>{entry.text!}</Markdown>
            </div>
          )
        )}
      </div>

      <ConfirmDeleteDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title="Delete note?"
        description="This permanently removes the note from the activity log."
        confirmLabel="Delete note"
        pending={saving}
        onConfirm={onDelete}
      />
    </li>
  );
}
