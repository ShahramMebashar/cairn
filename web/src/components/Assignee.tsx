import { Bot, User } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { actorKind, cn, initials } from "@/lib/utils";

// Assignee shows who owns a task with a clear agent (Bot) vs human (User) cue.
export function Assignee({ actor, className }: { actor?: string; className?: string }) {
  if (!actor) return null;
  const kind = actorKind(actor);
  const Icon = kind === "agent" ? Bot : User;
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className={cn("relative inline-grid", className)}>
          <Avatar className="size-5">
            <AvatarFallback className="text-[9px]">{initials(actor)}</AvatarFallback>
          </Avatar>
          <span
            className={cn(
              "absolute -right-1 -bottom-1 grid size-3 place-items-center rounded-full border bg-panel",
              kind === "agent" ? "text-brand" : "text-muted-foreground",
            )}
          >
            <Icon className="size-2" />
          </span>
        </span>
      </TooltipTrigger>
      <TooltipContent>
        {kind === "agent" ? "AI agent" : kind === "human" ? "Human" : "Assignee"} · {actor}
      </TooltipContent>
    </Tooltip>
  );
}
