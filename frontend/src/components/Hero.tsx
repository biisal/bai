import { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { installCmd, repoUrl } from '@/config'
import { cn } from '@/lib/utils'
import TerminalHeader from './ui/terminal-header'

const installTabs = [
  { id: 'curl', label: 'curl', cmd: installCmd },
  { id: 'source', label: 'source', cmd: `git clone ${repoUrl} && cd bai && make build` },
]

export default function Hero() {
  const [copied, setCopied] = useState(false)
  const [tab, setTab] = useState(0)
  const cmd = installTabs[tab].cmd

  const copy = async () => {
    await navigator.clipboard.writeText(cmd)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <section className="relative w-full min-h-[90vh] pt-24 pb-12 overflow-hidden flex items-center">
      {/* Background Image Layer */}
      <div className="absolute inset-0 z-0">
        <div className="absolute inset-0 bg-linear-to-r w-2/3 from-background  via-background/80 to-transparent z-10" />
        <div className="absolute inset-0 bg-linear-to-t  from-background via-transparent to-background/50 z-10" />
        <img
          src="cat-hero-2.png"
          alt="Serene mountain lake with glowing ring"
          className="w-full h-full object-cover object-right md:object-center"
        />
      </div>

      {/* Content Layer */}
      <div className="relative z-20 max-w-7xl mx-auto px-6 w-full flex flex-col md:flex-row items-center justify-between gap-12">
        
        {/* Left Copy */}
        <div className="w-full max-w-2xl">
          <p className="text-primary text-xs font-semibold tracking-[0.2em] uppercase mb-6">
            LightWeight. Less Bloat.
          </p>
          <h1 className="text-6xl sm:text-7xl lg:text-8xl font-['Playfair_Display'] text-white leading-[1.1] tracking-tight mb-6">
            Build Things.<br />
            <span className="text-foreground">Don't Break.</span>
          </h1>
          <p className="text-xl text-neutral-300 mb-10 max-w-lg">
            Invisible undo. Queue while it works. Zero git pollution.
          </p>

          {/* Install command / terminal */}
          <div className="w-full  rounded-sm border border-border bg-muted/20 backdrop-blur-sm overflow-hidden">
            <TerminalHeader title="install" />
            <div className="flex gap-1 border-b border-border px-3 pt-2 font-['JetBrains_Mono'] text-xs">
              {installTabs.map((t, i) => (
                <button
                  key={t.id}
                  onClick={() => setTab(i)}
                  aria-pressed={i === tab}
                  className={cn(
                    'border-b-2 px-3 py-1.5 transition-colors',
                    i === tab ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground',
                  )}
                >
                  {t.label}
                </button>
              ))}
            </div>
            <div className="flex items-center justify-between gap-4 px-4 py-4 font-['JetBrains_Mono'] text-sm">
              <p className="text-white/90 break-all">
                <span className="text-primary select-none">$ </span>{cmd}
              </p>
              <button
                onClick={copy}
                aria-label="Copy install command"
                className="shrink-0 flex items-center gap-1.5 rounded-sm border border-border bg-muted px-3 py-2 text-xs text-muted-foreground hover:text-foreground hover:border-primary/50 transition-colors"
              >
                {copied ? (
                  <>
                    <Check className="w-4 h-4 text-success" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4" />
                    Copy
                  </>
                )}
              </button>
            </div>
          </div>
          <div className="mt-6 flex items-center gap-3 font-['JetBrains_Mono'] text-xs">
            <span className="rounded-sm border border-primary/50 bg-primary/10 px-2.5 py-1.5 font-semibold text-primary">
              14 MB
            </span>
            <span className="text-muted-foreground">one static binary. no runtime, no model downloads.</span>
          </div>
        </div>

        
        
      </div>
    </section>
  )
}
