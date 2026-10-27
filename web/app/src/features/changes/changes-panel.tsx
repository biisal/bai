import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { api, errorText } from "@/lib/api";
import { parsePatchFiles } from "@pierre/diffs";
import { CodeView, type CodeViewHandle } from "@pierre/diffs/react";
import { CopyIcon, RefreshCwIcon, XIcon } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";

type ChangesPanelProps = {
  open: boolean;
  conversationId: string | null;
  inline?: boolean;
  onOpenChange: (open: boolean) => void;
  theme: "light" | "dark";
};

const diffOptions = {
  diffStyle: "unified",
  overflow: "wrap",
  hunkSeparators: "line-info-basic",
  stickyHeader: true,
} as const;

export function ChangesPanel({
  open,
  conversationId,
  inline = false,
  onOpenChange,
  theme,
}: ChangesPanelProps) {
  const [diff, setDiff] = useState("");
  const [truncated, setTruncated] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reload, setReload] = useState(0);
  const view = useRef<CodeViewHandle<undefined, undefined>>(null);
  const parsed = useMemo(
    () => (diff.trim() ? parsePatchFiles(diff, "bai-changes") : []),
    [diff]
  );
  const items = useMemo(
    () =>
      parsed.flatMap((patch, patchIndex) =>
        patch.files.map((fileDiff, fileIndex) => ({
          id: `${patchIndex}:${fileIndex}:${fileDiff.name}`,
          type: "diff" as const,
          fileDiff,
        }))
      ),
    [parsed]
  );
  const files = useMemo(
    () => parsed.flatMap((patch) => patch.files.map((file) => file.name)),
    [parsed]
  );
  const options = useMemo(
    () => ({
      ...diffOptions,
      theme: theme === "dark" ? "pierre-dark" : "pierre-light",
      themeType: theme,
    }),
    [theme]
  );

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

  const toolbar = (
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
  );

  const body = (
    <div className="min-h-0 flex-1">
      {!conversationId ? (
        <p className="p-5 text-sm text-muted-foreground">
          Open a conversation to see its changes.
        </p>
      ) : loading ? (
        <p className="p-5 text-sm text-muted-foreground">Loading changes...</p>
      ) : error ? (
        <p className="p-5 text-sm text-destructive">{error}</p>
      ) : items.length === 0 ? (
        <p className="p-5 text-sm text-muted-foreground">
          No file changes yet.
        </p>
      ) : (
        <div className="flex h-full flex-col">
          <div className="flex flex-wrap gap-1.5 border-b border-border/60 p-3">
            {items.map((item, index) => (
              <button
                className="max-w-full truncate rounded-md bg-muted/60 px-2 py-1 font-mono text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
                key={item.id}
                onClick={() =>
                  view.current?.scrollTo({
                    type: "item",
                    id: item.id,
                    align: "start",
                  })
                }
                type="button"
              >
                {files[index]}
              </button>
            ))}
          </div>
          {truncated && (
            <p className="border-b border-border/60 px-4 py-2 text-xs text-muted-foreground">
              Diff truncated at 1 MB.
            </p>
          )}
          <CodeView
            className="h-full min-h-0 flex-1 overflow-auto"
            items={items}
            options={options}
            ref={view}
          />
        </div>
      )}
    </div>
  );

  if (inline) {
    return (
      <aside className="flex h-full min-h-0 min-w-0 flex-col bg-background">
        <header className="flex items-start justify-between gap-4 border-b border-border/60 p-4">
          <div className="min-w-0">
            <h2 className="text-base font-medium">Changes</h2>
            <p className="text-sm text-muted-foreground">
              Read-only diff against the shadow repository.
            </p>
          </div>
          <Button
            aria-label="Close changes"
            onClick={() => onOpenChange(false)}
            size="icon-sm"
            type="button"
            variant="ghost"
          >
            <XIcon />
          </Button>
        </header>
        {toolbar}
        {body}
      </aside>
    );
  }

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
        {toolbar}
        {body}
      </SheetContent>
    </Sheet>
  );
}
