export default function TerminalHeader({ title }: { title?: string }) {
  return (
    <div className="flex items-center gap-2 px-4 py-3 bg-muted border-b border-border">
      <span className="w-3 h-3 rounded-full bg-[#ff5f57]" />
      <span className="w-3 h-3 rounded-full bg-[#febc2e]" />
      <span className="w-3 h-3 rounded-full bg-[#28c840]" />
      {title && (
        <span className="text-muted-foreground text-xs font-['JetBrains_Mono'] ml-2">
          {title}
        </span>
      )}
    </div>
  );
}
