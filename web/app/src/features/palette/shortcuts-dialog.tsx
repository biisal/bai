import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

type ShortcutsDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

const shortcuts = [
  ["Cmd/Ctrl + K", "Command palette"],
  ["Cmd/Ctrl + B", "Toggle sidebar"],
  ["Cmd/Ctrl + ,", "Open settings"],
  ["@", "Pick a file to mention"],
  ["/", "Open slash commands"],
  ["?", "Show this cheat sheet"],
  ["Esc", "Stop a run when the prompt is focused and empty"],
];

export function ShortcutsDialog({ open, onOpenChange }: ShortcutsDialogProps) {
  return (
    <Dialog onOpenChange={onOpenChange} open={open}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Keyboard shortcuts</DialogTitle>
          <DialogDescription>Small keys, fewer clicks.</DialogDescription>
        </DialogHeader>
        <dl className="divide-y divide-border/60">
          {shortcuts.map(([keys, action]) => (
            <div
              className="flex items-center justify-between gap-4 py-2.5 text-sm"
              key={keys}
            >
              <dt className="text-muted-foreground">{action}</dt>
              <dd className="shrink-0 rounded bg-muted px-2 py-0.5 font-mono text-[11px]">
                {keys}
              </dd>
            </div>
          ))}
        </dl>
      </DialogContent>
    </Dialog>
  );
}
