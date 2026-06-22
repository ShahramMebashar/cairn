import { useState } from "react";
import Anser from "anser";
import { Check, Copy, WrapText } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import type { Run } from "@/lib/api";

// LogView renders a check run's captured output: ANSI colors → HTML (escaped, so safe),
// with a copy button and a soft-wrap toggle. The header shows exit code / duration / timeout.
export function LogView({ run }: { run: Run }) {
  const [wrap, setWrap] = useState(true);
  const [copied, setCopied] = useState(false);
  const output = run.output?.trimEnd() ?? "";
  const html = output ? Anser.ansiToHtml(output, { use_classes: false }) : "";

  const copy = () => {
    navigator.clipboard.writeText(output).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1200);
    });
  };

  return (
    <div className="overflow-hidden rounded-md border bg-muted/40">
      <div className="flex items-center gap-3 border-b px-2.5 py-1 font-mono text-[11px] text-muted-foreground">
        <span className={cn(run.exit === 0 ? "text-success" : "text-destructive")}>
          exit {run.exit}
        </span>
        {run.duration && <span>{run.duration}</span>}
        {run.timedout && <span className="text-destructive">timed out</span>}
        <span className="ml-auto flex items-center gap-0.5">
          <button
            onClick={() => setWrap((w) => !w)}
            aria-label="Toggle wrap"
            className={cn(
              "grid size-5 place-items-center rounded hover:bg-foreground/10",
              wrap && "text-foreground",
            )}
          >
            <WrapText className="size-3" />
          </button>
          <button
            onClick={copy}
            aria-label="Copy output"
            className="grid size-5 place-items-center rounded hover:bg-foreground/10"
          >
            {copied ? <Check className="size-3 text-success" /> : <Copy className="size-3" />}
          </button>
        </span>
      </div>
      <ScrollArea className="max-h-64">
        {output ? (
          <pre
            className={cn(
              "px-2.5 py-2 font-mono text-[11px] leading-relaxed",
              wrap ? "break-words whitespace-pre-wrap" : "whitespace-pre",
            )}
            dangerouslySetInnerHTML={{ __html: html }}
          />
        ) : (
          <pre className="px-2.5 py-2 font-mono text-[11px] text-muted-foreground">(no output)</pre>
        )}
      </ScrollArea>
    </div>
  );
}
