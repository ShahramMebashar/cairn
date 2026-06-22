import { useState } from "react";
import { ArrowLeft, Loader2, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useInitRepo } from "@/lib/queries";
import type { Status } from "@/lib/api";

export function InitProject({
  path,
  status,
  onChangeFolder,
}: {
  path: string;
  status: Status;
  onChangeFolder: () => void;
}) {
  const [prefix, setPrefix] = useState(status.suggestedPrefix);
  const init = useInitRepo(path);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-6 text-foreground">
      <div className="w-full max-w-md">
        <button
          onClick={onChangeFolder}
          className="mb-4 flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="size-4" /> Choose another folder
        </button>

        <div className="rounded-xl border bg-card p-6 text-card-foreground shadow-xs">
          <div className="flex items-center gap-3">
            <span className="grid size-9 place-items-center rounded-lg bg-brand/10 text-brand">
              <Sparkles className="size-5" />
            </span>
            <div>
              <h1 className="text-lg font-semibold tracking-tight">Initialize cairn</h1>
              <p className="text-sm text-muted-foreground">No workspace here yet.</p>
            </div>
          </div>

          <p className="mt-4 truncate font-mono text-xs text-muted-foreground">{status.root}</p>

          <div className="mt-5 grid gap-1.5">
            <Label htmlFor="prefix">Task id prefix</Label>
            <Input
              id="prefix"
              value={prefix}
              onChange={(e) => setPrefix(e.target.value.toUpperCase())}
              onKeyDown={(e) => e.key === "Enter" && init.mutate(prefix.trim())}
              placeholder="PROJ"
              className="font-mono"
            />
            <p className="text-xs text-muted-foreground">
              Tasks will be created as{" "}
              <span className="font-mono">{(prefix || "PROJ") + "-001"}</span>.
            </p>
          </div>

          <Button
            className="mt-5 w-full"
            disabled={init.isPending}
            onClick={() => init.mutate(prefix.trim())}
          >
            {init.isPending && <Loader2 className="animate-spin" />}
            {init.isPending ? "Initializing…" : "Initialize workspace"}
          </Button>
        </div>
      </div>
    </div>
  );
}
