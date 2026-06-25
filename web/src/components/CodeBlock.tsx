import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";

// CodeBlock is a labeled, copy-to-clipboard snippet box — shared by the Connect page and the
// quick Connect-agent dialog. `label` is a short uppercase tag (e.g. the language or target).
export function CodeBlock({ label, text }: { label: string; text: string }) {
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
