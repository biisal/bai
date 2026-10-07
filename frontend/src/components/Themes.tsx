const theme = `{
  "background": "",
  "foreground": "7",
  "primary": "6",
  "muted": "236",
  "success": "2",
  "warning": "3",
  "destructive": "1"
}`

export default function Themes() {
  return (
    <section id="themes" className="mx-auto w-full max-w-7xl px-6 py-24">
      <div className="grid items-start gap-12 lg:grid-cols-2 lg:gap-16">
        <div>
          <p className="mb-4 text-xs font-semibold uppercase tracking-[0.2em] text-primary">Themes</p>
          <h2 className="mb-5 font-['Playfair_Display'] text-4xl leading-[1.15] text-white md:text-5xl">
            Colors live in a JSON file.
          </h2>
          <p className="max-w-lg text-muted-foreground">
            Drop a theme in <span className="font-['JetBrains_Mono'] text-foreground">~/.config/bai/themes</span> and switch with{' '}
            <span className="font-['JetBrains_Mono'] text-foreground">/themes</span>. The choice is saved and reloaded on start. Missing keys fall back to the built-in theme.
          </p>
        </div>

        <div className="overflow-hidden rounded-sm border border-border bg-muted/30">
          <div className="border-b border-border px-4 py-2 font-['JetBrains_Mono'] text-xs text-muted-foreground">
            ~/.config/bai/themes/default.json
          </div>
          <pre className="overflow-x-auto px-4 py-3 font-['JetBrains_Mono'] text-xs leading-5 text-foreground/90">{theme}</pre>
          <p className="border-t border-border px-4 py-3 text-xs text-muted-foreground">
            Values are ANSI color numbers. A theme that disappears from disk falls back the same way.
          </p>
        </div>
      </div>
    </section>
  )
}
