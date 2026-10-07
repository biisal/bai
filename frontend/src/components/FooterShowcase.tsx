import { useState } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from './ui/button'
import { installUrl, repoUrl, siteName } from '@/config'

const cards = [
  { title: 'Safe Experimentation', desc: 'Let the agent try things. Every turn lands in a private git dir, so your repo stays clean.' },
  { title: 'Atomic Edits', desc: 'Exact-match replacements, serialized per file, so concurrent writes never collide.' },
  { title: 'Easy Rollbacks', desc: 'Each tool call carries a purpose that becomes the commit message. Find any step and revert it.' },
]

const stats = [
  { value: '1', label: 'Binary' },
  { value: '6', label: 'Providers' },
  { value: 'SQLite', label: 'Sessions' },
  { value: '4', label: 'Built-in tools' },
]

export default function FooterShowcase() {
  const [i, setI] = useState(0)
  const go = (d: number) => setI((i + d + cards.length) % cards.length)

  return (
    <section id="docs" className="relative z-10 mx-auto w-full max-w-7xl px-6 pt-24 pb-10">
      <div className="flex flex-col gap-16 lg:flex-row lg:items-center">
        {/* Carousel */}
        <div className="w-full lg:w-3/5">
          <div className="mb-6 flex items-center justify-between">
            <Button variant="outline" asChild>
              <a href={repoUrl}>Explore Shadow Git &rarr;</a>
            </Button>
            <div className="flex gap-2">
              <button
                onClick={() => go(-1)}
                aria-label="Previous"
                className="flex h-10 w-10 items-center justify-center rounded-sm border border-border text-muted-foreground transition-colors hover:border-primary/50 hover:text-foreground"
              >
                <ChevronLeft className="h-5 w-5" />
              </button>
              <button
                onClick={() => go(1)}
                aria-label="Next"
                className="flex h-10 w-10 items-center justify-center rounded-sm border border-border text-muted-foreground transition-colors hover:border-primary/50 hover:text-foreground"
              >
                <ChevronRight className="h-5 w-5" />
              </button>
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-3">
            {cards.map((c, idx) => (
              <button
                key={c.title}
                onClick={() => setI(idx)}
                className={`rounded-sm border p-5 text-left transition-colors ${
                  idx === i ? 'border-primary/60 bg-primary/10' : 'border-border bg-muted/30 hover:border-primary/30'
                }`}
              >
                <div className="mb-2 font-['Playfair_Display'] text-xl text-white">{c.title}</div>
                <p className="text-sm leading-relaxed text-muted-foreground">{c.desc}</p>
              </button>
            ))}
          </div>
        </div>

        {/* Stats */}
        <div className="w-full lg:w-2/5">
          <h3 className="mb-6 font-['Playfair_Display'] text-3xl text-white">Built for developer productivity.</h3>
          <div className="grid grid-cols-2 gap-4">
            {stats.map(s => (
              <div key={s.label} className="rounded-sm border border-border bg-muted/30 p-4">
                <div className="font-['JetBrains_Mono'] text-2xl font-bold text-primary">{s.value}</div>
                <div className="text-xs text-muted-foreground">{s.label}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <footer className="mt-24 flex flex-col items-center justify-between gap-4 border-t border-border pt-6 text-sm text-muted-foreground sm:flex-row">
        <span>{siteName} — a terminal AI coding agent.</span>
        <div className="flex gap-6">
          <a href={repoUrl} className="transition-colors hover:text-foreground">GitHub</a>
          <a href={`${repoUrl}#readme`} className="transition-colors hover:text-foreground">Docs</a>
          <a href={installUrl} className="transition-colors hover:text-foreground">Install script</a>
        </div>
      </footer>
    </section>
  )
}
