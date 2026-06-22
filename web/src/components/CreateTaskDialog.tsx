import { useState } from "react";
import { Loader2, Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { MarkdownEditor } from "@/components/MarkdownEditor";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { PriorityIcon, PRIORITIES, priorityLabel } from "@/components/PriorityIcon";
import { useCreateTask } from "@/lib/queries";
import type { Check } from "@/lib/api";

export function CreateTaskDialog({
  path,
  open,
  onOpenChange,
  defaultParent,
}: {
  path: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  defaultParent?: string;
}) {
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [deps, setDeps] = useState("");
  const [checks, setChecks] = useState<Check[]>([]);
  const [priority, setPriority] = useState("");
  const [labels, setLabels] = useState("");
  const create = useCreateTask(path);

  function reset() {
    setTitle("");
    setBody("");
    setDeps("");
    setChecks([]);
    setPriority("");
    setLabels("");
  }

  function submit() {
    const parsedDeps = deps
      .split(/[\s,]+/)
      .map((s) => s.trim())
      .filter(Boolean);
    const parsedLabels = labels
      .split(/[\s,]+/)
      .map((s) => s.trim())
      .filter(Boolean);
    const parsedChecks = checks.filter((c) => c.desc.trim());
    create.mutate(
      {
        title: title.trim(),
        body: body.trim() ? body.trim() + "\n" : undefined,
        deps: parsedDeps.length ? parsedDeps : undefined,
        checks: parsedChecks.length ? parsedChecks : undefined,
        labels: parsedLabels.length ? parsedLabels : undefined,
        priority: priority || undefined,
        parent: defaultParent || undefined,
      },
      {
        onSuccess: () => {
          reset();
          onOpenChange(false);
        },
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{defaultParent ? "New sub-task" : "New task"}</DialogTitle>
          <DialogDescription>
            {defaultParent ? (
              <>
                Sub-task of <span className="font-mono">{defaultParent}</span>. The engine assigns
                the id and initial status.
              </>
            ) : (
              "The engine assigns the id and initial status."
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="ct-title">Title</Label>
            <Input
              id="ct-title"
              autoFocus
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Add idempotency keys to the webhook"
            />
          </div>

          <div className="grid gap-1.5">
            <Label>Description</Label>
            <MarkdownEditor
              value={body}
              onChange={setBody}
              placeholder="Intent and constraints…"
              minHeight="8rem"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="grid gap-1.5">
              <Label>Priority</Label>
              <Select value={priority || "none"} onValueChange={(v) => setPriority(v === "none" ? "" : v)}>
                <SelectTrigger className="h-8 w-full">
                  <span className="flex items-center gap-2">
                    <PriorityIcon priority={priority} />
                    {priorityLabel(priority)}
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
            </div>
            <div className="grid gap-1.5">
              <Label htmlFor="ct-labels">Labels</Label>
              <Input
                id="ct-labels"
                value={labels}
                onChange={(e) => setLabels(e.target.value)}
                placeholder="backend, db"
                className="h-8 text-sm"
              />
            </div>
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="ct-deps">Dependencies</Label>
            <Input
              id="ct-deps"
              value={deps}
              onChange={(e) => setDeps(e.target.value)}
              placeholder="PROJ-001, PROJ-002"
              className="font-mono text-sm"
            />
            <p className="text-xs text-muted-foreground">Must be closed before this can start.</p>
          </div>

          <div className="grid gap-2">
            <div className="flex items-center justify-between">
              <Label>Checks</Label>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setChecks((cs) => [...cs, { desc: "", cmd: "", result: "pending" }])}
              >
                <Plus /> Add check
              </Button>
            </div>
            {checks.map((c, i) => (
              <div key={i} className="flex items-center gap-2">
                <Input
                  value={c.desc}
                  onChange={(e) =>
                    setChecks((cs) => cs.map((x, j) => (j === i ? { ...x, desc: e.target.value } : x)))
                  }
                  placeholder="what it verifies"
                  className="flex-1"
                />
                <Input
                  value={c.cmd ?? ""}
                  onChange={(e) =>
                    setChecks((cs) => cs.map((x, j) => (j === i ? { ...x, cmd: e.target.value } : x)))
                  }
                  placeholder="cmd (blank = manual)"
                  className="flex-1 font-mono text-xs"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  aria-label="Remove check"
                  onClick={() => setChecks((cs) => cs.filter((_, j) => j !== i))}
                >
                  <X />
                </Button>
              </div>
            ))}
          </div>
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={submit} disabled={!title.trim() || create.isPending}>
            {create.isPending && <Loader2 className="animate-spin" />}
            Create task
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
