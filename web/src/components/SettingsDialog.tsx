import { useEffect, useState } from "react";
import { RefreshCw } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  autostartEnabled,
  isTauri,
  osNotifEnabled,
  setAutostart,
  setOsNotifEnabled,
} from "@/lib/desktop";
import { useSetCheckShell } from "@/lib/queries";

// SettingsDialog holds desktop preferences. In the browser it shows only a short note,
// since every toggle here drives a native capability.
export function SettingsDialog({
  open,
  onOpenChange,
  onCheckUpdates,
  path,
  checkShell,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCheckUpdates: () => void;
  path: string;
  checkShell?: string;
}) {
  const desktop = isTauri();
  const [notifs, setNotifs] = useState(true);
  const [autostart, setAutostartState] = useState(false);
  const [checking, setChecking] = useState(false);
  const [shell, setShell] = useState(checkShell ?? "");
  const saveShell = useSetCheckShell(path);
  const shellDirty = shell.trim() !== (checkShell ?? "");

  useEffect(() => {
    if (open) setShell(checkShell ?? "");
  }, [open, checkShell]);

  useEffect(() => {
    if (!open || !desktop) return;
    setNotifs(osNotifEnabled());
    void autostartEnabled().then(setAutostartState);
  }, [open, desktop]);

  const toggleNotifs = (on: boolean) => {
    setNotifs(on);
    setOsNotifEnabled(on);
  };

  const toggleAutostart = async (on: boolean) => {
    setAutostartState(on); // optimistic
    try {
      await setAutostart(on);
    } catch {
      setAutostartState(!on); // revert on failure
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Settings</DialogTitle>
          <DialogDescription>Preferences for this workspace and machine.</DialogDescription>
        </DialogHeader>

        <div className="space-y-1">
          <Row
            title="Check shell"
            desc="Shell that runs command checks. Empty uses sh; on Windows use Git Bash/WSL or a path to a shell. (The CAIRN_SHELL env var overrides this.)"
          >
            <div className="flex items-center gap-1.5">
              <Input
                value={shell}
                onChange={(e) => setShell(e.target.value)}
                placeholder="sh"
                spellCheck={false}
                className="h-8 w-40 font-mono text-sm"
              />
              <Button
                variant="outline"
                size="sm"
                disabled={!shellDirty || saveShell.isPending}
                onClick={() => saveShell.mutate(shell.trim())}
              >
                Save
              </Button>
            </div>
          </Row>

          {desktop ? (
            <>
              <Row
                title="System notifications"
                desc="Alert me when a task becomes ready, a check fails, or work is awaiting review."
              >
                <Switch checked={notifs} onCheckedChange={toggleNotifs} />
              </Row>
              <Row
                title="Launch at login"
                desc="Start Cairn in the background so agents can always reach the MCP endpoint."
              >
                <Switch checked={autostart} onCheckedChange={(v) => void toggleAutostart(v)} />
              </Row>
              <Row title="Updates" desc="Check GitHub for a newer version.">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={checking}
                  onClick={() => {
                    setChecking(true);
                    onCheckUpdates();
                    setTimeout(() => setChecking(false), 1500);
                  }}
                >
                  <RefreshCw className={checking ? "animate-spin" : undefined} />
                  Check
                </Button>
              </Row>
            </>
          ) : (
            <p className="pt-2 text-xs text-muted-foreground">
              Notifications, launch-at-login, and updates are available in the Cairn desktop app.
            </p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

function Row({
  title,
  desc,
  children,
}: {
  title: string;
  desc: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-4 py-2.5">
      <div className="min-w-0">
        <p className="text-sm font-medium">{title}</p>
        <p className="text-xs text-muted-foreground">{desc}</p>
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  );
}
