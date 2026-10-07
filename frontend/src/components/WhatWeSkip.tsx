const skips = [
  { no: 'No runtime to install', yes: 'One CGO-free binary. Drop it in and run.' },
  { no: 'No git pollution', yes: 'Every turn is committed to a private .bai_git, never your repo history.' },
  { no: 'No tool sprawl', yes: 'Four built-in tools: read, write, edit, bash. Add your own through a manifest.' },
  { no: 'No locked-in provider', yes: 'OpenAI, Anthropic, Groq, or any OpenAI-compatible endpoint from config.' },
  { no: 'No waiting around', yes: 'Keep typing while the agent works. Messages queue up in order.' },
  { no: 'No cloud account', yes: 'Bring an API key. Sessions live in a local SQLite file.' },
]

export default function WhatWeSkip() {
  return (
    <section className="mx-auto w-full max-w-7xl px-6 py-24">
      <p className="mb-4 text-xs font-semibold uppercase tracking-[0.2em] text-primary">Less bloat</p>
      <h2 className="mb-12 max-w-xl font-['Playfair_Display'] text-4xl leading-[1.15] text-white md:text-5xl">
        What bAI leaves out.
      </h2>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {skips.map(s => (
          <div key={s.no} className="rounded-sm border border-border bg-muted/30 p-5">
            <h3 className="mb-2 font-['JetBrains_Mono'] text-sm font-bold text-white">{s.no}</h3>
            <p className="text-sm leading-relaxed text-muted-foreground">{s.yes}</p>
          </div>
        ))}
      </div>
    </section>
  )
}
