import { useState } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'

// Statements describing how bAI actually behaves — not invented user testimonials.
// Swap in real user quotes here once you have them.
const quotes = [
  { text: 'Every turn is a commit in a private git dir. Your own history stays clean.', from: 'Shadow Git', detail: '.bai_git' },
  { text: 'Keep typing while the agent works. Messages queue up and run in order.', from: 'Message queue', detail: 'TUI' },
  { text: 'One binary, no CGO, no runtime to install. Drop it in and run.', from: 'Single binary', detail: 'Linux · macOS · Windows' },
]

export default function Testimonial() {
  const [i, setI] = useState(0)
  const q = quotes[i]
  const go = (d: number) => setI((i + d + quotes.length) % quotes.length)

  return (
    <section className="relative w-full overflow-hidden py-28">
      <div className="relative z-10 mx-auto max-w-3xl px-6 text-center">
        <div>
          <p className="mb-6 text-xs font-semibold uppercase tracking-[0.2em] text-primary">
            Small steps. Big changes.
          </p>
          <h2 className="mb-10 font-['Playfair_Display'] text-4xl leading-[1.15] text-white md:text-5xl">
            A cleaner repo<br />builds better code.
          </h2>

          <blockquote className="mb-8 min-h-[5.5rem] text-lg leading-relaxed text-foreground" aria-live="polite">
            “{q.text}”
          </blockquote>

          <div className="flex items-center justify-center gap-6">
            <button
              onClick={() => go(-1)}
              aria-label="Previous"
              className="flex h-10 w-10 items-center justify-center rounded-sm border border-border text-muted-foreground transition-colors hover:border-primary/50 hover:text-foreground"
            >
              <ChevronLeft className="h-5 w-5" />
            </button>
            <div className="min-w-40 text-sm">
              <div className="font-medium text-white">{q.from}</div>
              <div className="font-['JetBrains_Mono'] text-xs text-muted-foreground">{q.detail}</div>
            </div>
            <button
              onClick={() => go(1)}
              aria-label="Next"
              className="flex h-10 w-10 items-center justify-center rounded-sm border border-border text-muted-foreground transition-colors hover:border-primary/50 hover:text-foreground"
            >
              <ChevronRight className="h-5 w-5" />
            </button>
          </div>
        </div>
      </div>
    </section>
  )
}
