import { Bot, GitBranch, History, ListChecks, Terminal } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

const SECTIONS = [
  {
    icon: ListChecks,
    title: "Tasks & states",
    body: "Every task is a file in your repo. It moves through your configured states (e.g. backlog → in progress → in review → done). The id is assigned for you and never reused.",
  },
  {
    icon: GitBranch,
    title: "Two gates keep work honest",
    body: "Dependencies: a task can't leave the backlog until everything it depends on is closed. Checks: a task can't be marked done until its checks pass (commands run automatically; manual checks you attest).",
  },
  {
    icon: Terminal,
    title: "Checks run a shell",
    body: "A command check runs in a POSIX shell (sh) — go test ./..., pytest -q && ruff check ., ./scripts/verify.sh. On Windows install Git Bash or WSL, or point CAIRN_SHELL at a shell on your PATH.",
  },
  {
    icon: Bot,
    title: "You + your AI agent",
    body: "Humans and agents share one task list and one rule-set. Claim a task to take it; hand off by leaving it for the other. Connect Claude Code over MCP and it sees exactly what you see.",
  },
  {
    icon: History,
    title: "Provenance",
    body: "Each task keeps an append-only log of what happened and who did it — created, claimed, transitioned, notes, attestations — so the decision trail lives with the work.",
  },
];

export function HelpDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (o: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>How cairn works</DialogTitle>
          <DialogDescription>A repo-native task tracker you run with your AI agent.</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          {SECTIONS.map(({ icon: Icon, title, body }) => (
            <div key={title} className="flex items-start gap-3">
              <span className="mt-0.5 grid size-7 shrink-0 place-items-center rounded-md bg-muted text-muted-foreground">
                <Icon className="size-4" />
              </span>
              <div>
                <h3 className="text-sm font-medium">{title}</h3>
                <p className="mt-0.5 text-[13px] leading-relaxed text-muted-foreground">{body}</p>
              </div>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
