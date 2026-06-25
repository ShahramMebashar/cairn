import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { mcpAddCommand } from "@/lib/connect";

// ConnectAgentDialog shows ready-to-paste MCP connection details for the open project.
// HTTP targets the running app's /mcp endpoint (this same server); stdio points an
// agent at the bundled `cairn` binary. The actor is prefilled from the current identity.
export function ConnectAgentDialog({
  open,
  onOpenChange,
  path,
  actor,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  path: string;
  actor: string;
}) {
  // The webview origin is the cairn server in the packaged app; falls back gracefully.
  const base = typeof window !== "undefined" ? window.location.origin : "http://127.0.0.1:7777";
  const who = actor || "agent:claude-1";
  const httpCmd = mcpAddCommand(base, path, who);
  const stdioJson = JSON.stringify(
    { mcpServers: { cairn: { command: "cairn", args: ["serve", "--actor", who, "--repo", path] } } },
    null,
    2,
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Connect an agent</DialogTitle>
          <DialogDescription>
            Give an AI agent access to this project over MCP — the same rule-set as the UI.
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="http" className="mt-1">
          <TabsList className="w-full">
            <TabsTrigger value="http">HTTP (running app)</TabsTrigger>
            <TabsTrigger value="stdio">Stdio (binary)</TabsTrigger>
          </TabsList>

          <TabsContent value="http" className="space-y-2">
            <p className="text-sm text-muted-foreground">
              For URL-based clients (Claude Code, newer Claude Desktop). Connects to this app while it's open.
            </p>
            <CodeBlock label="terminal" text={httpCmd} />
          </TabsContent>

          <TabsContent value="stdio" className="space-y-2">
            <p className="text-sm text-muted-foreground">
              For stdio clients. Requires the <code className="font-mono">cairn</code> binary on your PATH.
            </p>
            <CodeBlock label="mcp config" text={stdioJson} />
          </TabsContent>
        </Tabs>

        <p className="text-xs text-muted-foreground">
          Edit the <span className="font-mono">actor</span> to identify each agent (e.g.{" "}
          <span className="font-mono">agent:claude-2</span>).
        </p>
      </DialogContent>
    </Dialog>
  );
}

function CodeBlock({ label, text }: { label: string; text: string }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      toast.success("Copied to clipboard");
      setTimeout(() => setCopied(false), 1500);
    } catch {
      toast.error("Could not copy");
    }
  };

  return (
    <div className="relative rounded-lg border bg-muted/40">
      <div className="flex items-center justify-between border-b px-3 py-1.5">
        <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
          {label}
        </span>
        <Button variant="ghost" size="sm" className="h-6 gap-1 px-2 text-xs" onClick={copy}>
          {copied ? <Check className="size-3" /> : <Copy className="size-3" />}
          {copied ? "Copied" : "Copy"}
        </Button>
      </div>
      <pre className="overflow-x-auto px-3 py-2.5 text-xs leading-relaxed">
        <code className="font-mono">{text}</code>
      </pre>
    </div>
  );
}
