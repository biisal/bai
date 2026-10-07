import { useState } from 'react'
import { Check, Copy } from 'lucide-react'

const tokens = [
  'background',
  'foreground',
  'primary',
  'muted',
  'accent',
  'border',
  'destructive',
  'success',
  'warning',
] as const

type Token = (typeof tokens)[number]

const defaults: Record<Token, string> = {
  background: '#040f13',
  foreground: '#e7eee9',
  primary: '#81e6d9',
  muted: '#8ba6a5',
  accent: '#a7f3d0',
  border: '#1e3a3a',
  destructive: '#f87171',
  success: '#4ade80',
  warning: '#fbbf24',
}

const jsonFor = (colors: Record<Token, string>) => {
  const lines = tokens.map(name => `  "${name}": "${colors[name]}"`)
  return `{\n${lines.join(',\n')}\n}`
}

export function ThemeJsonEditor() {
  const [colors, setColors] = useState(defaults)
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    await navigator.clipboard.writeText(jsonFor(colors))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="overflow-hidden rounded-sm border border-border bg-muted/30">
      <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-2">
        <span className="font-['JetBrains_Mono'] text-xs text-muted-foreground">
          ~/.config/bai/themes/default.json
        </span>
        <button
          type="button"
          onClick={copy}
          aria-label="Copy theme JSON"
          className="flex items-center gap-1.5 font-['JetBrains_Mono'] text-xs text-muted-foreground transition-colors hover:text-foreground"
        >
          {copied ? <Check className="h-3.5 w-3.5 text-success" /> : <Copy className="h-3.5 w-3.5" />}
          {copied ? 'Copied' : 'Copy'}
        </button>
      </div>

      <div className="space-y-1 px-4 py-3 font-['JetBrains_Mono'] text-xs leading-6 text-foreground/90">
        <div>{'{'}</div>
        {tokens.map((name, i) => (
          <div key={name} className="flex items-center gap-2 pl-4">
            <span className="text-foreground/80">"{name}":</span>
            <label className="relative h-4 w-4 shrink-0 cursor-pointer overflow-hidden rounded-sm border border-border">
              <span className="absolute inset-0" style={{ backgroundColor: colors[name] }} />
              <input
                type="color"
                value={colors[name]}
                aria-label={`${name} color`}
                onChange={e => setColors(current => ({ ...current, [name]: e.target.value }))}
                className="absolute inset-0 cursor-pointer opacity-0"
              />
            </label>
            <span>"{colors[name]}"{i < tokens.length - 1 ? ',' : ''}</span>
          </div>
        ))}
        <div>{'}'}</div>
      </div>
    </div>
  )
}

export default function Themes() {
  return (
    <section id="themes" className="mx-auto w-full max-w-7xl px-6 py-24">
      <div className="grid items-start gap-12 lg:grid-cols-2 lg:gap-16">
        <div>
          <p className="mb-4 text-xs font-semibold uppercase tracking-[0.2em] text-primary">Themes</p>
          <h2 className="mb-5 font-['Playfair_Display'] text-4xl leading-[1.15] text-white md:text-5xl">
            Easy theming.
          </h2>
          <p className="max-w-lg text-muted-foreground">
            The keys are shadcn-inspired token names. Pick a color, copy the JSON into{' '}
            <span className="font-['JetBrains_Mono'] text-foreground">~/.config/bai/themes</span>, then switch with{' '}
            <span className="font-['JetBrains_Mono'] text-foreground">/themes</span>.
          </p>
        </div>

        <ThemeJsonEditor />
      </div>
    </section>
  )
}
