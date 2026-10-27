function greeting(): string {
  const hour = new Date().getHours();
  if (hour >= 6 && hour < 12) {
    return "Good morning";
  }
  if (hour >= 12 && hour < 17) {
    return "Good afternoon";
  }
  if (hour >= 17 && hour < 21) {
    return "Good evening";
  }
  return "Hello, night owl";
}

export function EmptyState({ directory }: { directory: string }) {
  const folder = directory.split(/[\\/]/).filter(Boolean).at(-1) ?? directory;

  return (
    <div className="text-center">
      <h1 className="text-3xl font-semibold tracking-tight">{greeting()}</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        What should we explore in{" "}
        <span className="font-medium text-foreground">{folder}</span>?
      </p>
    </div>
  );
}
