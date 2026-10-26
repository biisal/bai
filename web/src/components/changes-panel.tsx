import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { api, errorText } from "@/lib/api";
import { cn } from "@/lib/utils";
import { CopyIcon, RefreshCwIcon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

type ChangesPanelProps = {
  open: boolean;
  conversationId: string | null;
  onOpenChange: (open: boolean) => void;
};

type FileDiff = {
  path: string;
  lines: string[];
};

function parseDiff(diff: string): FileDiff[] {
  if (!diff.trim()) {
    return [];
  }
  const files: FileDiff[] = [];
  let current: FileDiff | null = null;
  for (const line of diff.split("\n")) {
    if (line.startsWith("diff --git ")) {
      const match = /^diff --git a\/(.+) b\/(.+)$/.exec(line);
      current = { path: match?.[2] ?? "unknown", lines: [] };
      files.push(current);
      continue;
    }
    if (
      line.startsWith("index ") ||
      line.startsWith("new file mode ") ||
      line.startsWith("deleted file mode ")
    ) {
      continue;
    }
    current?.lines.push(line);
  }
  return files;
}

export function ChangesPanel({
  open,
  conversationId,
  onOpenChange,
}: ChangesPanelProps) {
  const [diff, setDiff] = useState("");
  const [truncated, setTruncated] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reload, setReload] = useState(0);
  const files = useMemo(() => parseDiff(diff), [diff]);

  useEffect(() => {
    if (!open || !conversationId) {
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);
    setDiff("");
    setTruncated(false);
    api
      .diff(conversationId)
      .then((result) => {
        if (!cancelled) {
          setDiff(result.diff);
          setTruncated(result.truncated);
        }
      })
      .catch((cause: unknown) => {
        if (!cancelled) {
          setError(errorText(cause));
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [open, conversationId, reload]);

  const copy = () => {
    void navigator.clipboard
      .writeText(diff)
      .then(() => toast.success("Diff copied"))
      .catch(() => toast.error("Could not copy diff"));
  };

  return (
    <Sheet onOpenChange={onOpenChange} open={open}>
      <SheetContent
        className="w-full gap-0 sm:max-w-2xl"
        showCloseButton
        side="right"
      >
        <SheetHeader className="border-b border-border/60 pr-12">
          <SheetTitle>Changes</SheetTitle>
          <SheetDescription>
            Read-only diff against the shadow repository.
          </SheetDescription>
        </SheetHeader>

        <div className="flex items-center gap-2 border-b border-border/60 px-4 py-2">
          <Button
            disabled={loading}
            onClick={() => setReload((value) => value + 1)}
            size="sm"
            type="button"
            variant="ghost"
          >
            <RefreshCwIcon />
            Refresh
          </Button>
          <Button
            disabled={!diff}
            onClick={copy}
            size="sm"
            type="button"
            variant="ghost"
          >
            <CopyIcon />
            Copy
          </Button>
          <span className="ml-auto text-xs text-muted-foreground">
            {files.length} {files.length === 1 ? "file" : "files"}
          </span>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto">
          {!conversationId ? (
            <p className="p-5 text-sm text-muted-foreground">
              Open a conversation to see its changes.
            </p>
          ) : loading ? (
            <p className="p-5 text-sm text-muted-foreground">
              Loading changes...
            </p>
          ) : error ? (
            <p className="p-5 text-sm text-destructive">{error}</p>
          ) : files.length === 0 ? (
            <p className="p-5 text-sm text-muted-foreground">
              No file changes yet.
            </p>
          ) : (
            <>
              <div className="flex flex-wrap gap-1.5 border-b border-border/60 p-3">
                {files.map((file, index) => (
                  <button
                    className="max-w-full truncate rounded-md bg-muted/60 px-2 py-1 font-mono text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
                    key={`${file.path}-${index}`}
                    onClick={() =>
                      document
                        .getElementById(`diff-${index}`)
                        ?.scrollIntoView({ block: "start" })
                    }
                    type="button"
                  >
                    {file.path}
                  </button>
                ))}
              </div>
              {truncated && (
                <p className="border-b border-border/60 px-4 py-2 text-xs text-muted-foreground">
                  Diff truncated at 1 MB.
                </p>
              )}
              <div className="divide-y divide-border/60">
                {files.map((file, index) => (
                  <section id={`diff-${index}`} key={`${file.path}-${index}`}>
                    <h3 className="sticky top-0 z-10 truncate bg-popover px-4 py-2 font-mono text-xs font-medium">
                      {file.path}
                    </h3>
                    <pre className="overflow-x-auto px-0 pb-3 text-[11px] leading-5">
                      {file.lines.map((line, lineIndex) => (
                        <div
                          className={cn(
                            "whitespace-pre px-4",
                            line.startsWith("+") &&
                              !line.startsWith("+++") &&
                              "bg-success/10 text-success",
                            line.startsWith("-") &&
                              !line.startsWith("---") &&
                              "bg-destructive/10 text-destructive",
                            line.startsWith("@@") && "text-muted-foreground"
                          )}
                          key={`${lineIndex}-${line}`}
                        >
                          {line || " "}
                        </div>
                      ))}
                    </pre>
                  </section>
                ))}
              </div>
            </>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}
