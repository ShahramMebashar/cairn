import { useState } from "react";
import { Check, ChevronDown, ExternalLink, Loader2, Plug, Unplug } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { CodeBlock } from "@/components/CodeBlock";
import { cn } from "@/lib/utils";
import { useAgentManual, useConnectAgent, useDisconnectAgent, useIntegrations } from "@/lib/queries";
import type { AgentStatus, Status } from "@/lib/api";

// Connect is the integrations page: it detects which AI agents are installed and wires them
// to this project's MCP server in one click (the cairn process writes each agent's config),
// with a copy-paste manual guide for everything else.
export function Connect({ path }: { path: string; status: Status }) {
  const { data: agents, isLoading } = useIntegrations(path);

  const installed = (agents ?? []).filter((a) => a.installed);
  const others = (agents ?? []).filter((a) => !a.installed);

  return (
    <div className="flex h-full flex-col">
      <header className="flex h-11 shrink-0 items-center gap-2 border-b px-4">
        <Plug className="size-4 text-muted-foreground" />
        <h1 className="text-[13px] font-medium">Connect an agent</h1>
      </header>

      <div className="min-h-0 flex-1 overflow-y-auto">
        <div className="mx-auto max-w-3xl space-y-6 p-4 sm:p-6">
          <p className="text-sm text-muted-foreground">
            Give an AI agent the same task tools this UI uses. One click writes its MCP config for{" "}
            <span className="font-medium text-foreground">this project</span> — or open the manual
            setup to wire it up by hand. Each agent connects under its own identity (e.g.{" "}
            <span className="font-mono">agent:cursor</span>) so cairn attributes its work
            correctly; edit it on a card to run more than one instance.
          </p>

          {isLoading ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="size-4 animate-spin" /> Detecting agents…
            </div>
          ) : (
            <>
              {installed.length > 0 && (
                <Section title="Installed on this machine">
                  {installed.map((a) => (
                    <AgentCard key={a.id} path={path} agent={a} />
                  ))}
                </Section>
              )}
              <Section title={installed.length ? "All integrations" : "Integrations"}>
                {others.map((a) => (
                  <AgentCard key={a.id} path={path} agent={a} />
                ))}
              </Section>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="space-y-2">
      <h2 className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
        {title}
      </h2>
      <div className="grid gap-2">{children}</div>
    </section>
  );
}

function AgentCard({ path, agent }: { path: string; agent: AgentStatus }) {
  const [open, setOpen] = useState(false);
  // Each agent connects as itself by default (agent:cursor); editable for multiple instances.
  const [actor, setActor] = useState(`agent:${agent.id}`);
  const connect = useConnectAgent(path);
  const disconnect = useDisconnectAgent(path);
  const manual = useAgentManual(path, agent.id, actor, open);
  const canAuto = agent.mode === "auto";
  const busy = connect.isPending || disconnect.isPending;

  return (
    <div className="flex flex-col rounded-lg border bg-card p-3">
      <div className="flex items-center gap-3">
        <span className="grid size-8 shrink-0 place-items-center rounded-md bg-foreground/[0.06] text-[13px] font-semibold">
          {agent.name.slice(0, 1)}
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="truncate text-[13px] font-medium">{agent.name}</span>
            {agent.connected && (
              <span className="flex items-center gap-0.5 text-[11px] font-medium text-success">
                <Check className="size-3" /> Connected
              </span>
            )}
          </div>
          {/* Identity is the agent's own (agent:<id>) and editable inline for extra instances. */}
          <input
            value={actor}
            onChange={(e) => setActor(e.target.value)}
            spellCheck={false}
            aria-label={`${agent.name} identity`}
            className="w-full truncate rounded bg-transparent font-mono text-[11px] text-muted-foreground outline-none hover:text-foreground focus:text-foreground"
          />
        </div>
        {canAuto ? (
          <div className="flex shrink-0 items-center gap-1">
            <Button
              size="sm"
              variant={agent.connected ? "outline" : "default"}
              className="h-7"
              disabled={busy}
              onClick={() => connect.mutate({ agent: agent.id, actor })}
            >
              {connect.isPending && <Loader2 className="size-3 animate-spin" />}
              {agent.connected ? "Reconnect" : "Connect"}
            </Button>
            {agent.connected && (
              <Button
                size="icon"
                variant="ghost"
                className="size-7 text-muted-foreground hover:text-destructive"
                title="Disconnect (removes cairn from this agent's config)"
                aria-label={`Disconnect ${agent.name}`}
                disabled={busy}
                onClick={() => disconnect.mutate(agent.id)}
              >
                <Unplug className="size-3.5" />
              </Button>
            )}
          </div>
        ) : (
          <Badge variant="outline" className="h-5 shrink-0 text-[10px] text-muted-foreground">
            Manual
          </Badge>
        )}
      </div>

      <Collapsible open={open} onOpenChange={setOpen}>
        <div className="mt-2 flex items-center gap-3">
          <CollapsibleTrigger className="flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground">
            <ChevronDown className={cn("size-3 transition-transform", open && "rotate-180")} />
            Manual setup
          </CollapsibleTrigger>
          {agent.docsURL && (
            <a
              href={agent.docsURL}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground"
            >
              <ExternalLink className="size-3" /> Docs
            </a>
          )}
        </div>
        <CollapsibleContent className="pt-2">
          {manual.isLoading ? (
            <p className="text-[11px] text-muted-foreground">Loading…</p>
          ) : manual.data ? (
            <div className="space-y-1.5">
              {manual.data.path && (
                <p className="font-mono text-[11px] text-muted-foreground">
                  Add to {prettyPath(manual.data.path)}
                </p>
              )}
              <CodeBlock label={manual.data.lang} text={manual.data.config} />
            </div>
          ) : (
            <p className="text-[11px] text-muted-foreground">
              No snippet — see this agent's docs for MCP setup.
            </p>
          )}
        </CollapsibleContent>
      </Collapsible>
    </div>
  );
}

// prettyPath collapses the user's home dir for compact display.
function prettyPath(p: string): string {
  return p.replace(/^\/(?:Users|home)\/[^/]+/, "~");
}
