import { useEffect, useRef, useState, type ReactNode } from 'react'
import TerminalHeader from './ui/terminal-header'

type Line = {
  id: number
  kind: 'user' | 'reasoning' | 'cmd' | 'assistant' | 'notice'
  text: string
}

type Palette = { mode: 'commands' } | { mode: 'models'; sel: number }

const hints = [
  { key: 'escape', action: 'interrupt' },
  { key: 'ctrl+c', action: 'exit' },
  { key: '/', action: 'commands' },
  { key: '@', action: 'find files' },
]

const COMMANDS = [
  { name: 'new', desc: 'create a new conversation' },
  { name: 'themes', desc: 'list available themes' },
  { name: 'git_init', desc: 'initialize a git repository in this directory' },
  { name: 'exit', desc: 'exit the application' },
  { name: 'models', desc: 'show available models' },
  { name: 'sessions', desc: 'show list of conversations' },
]

const MODELS = [
  'anthropic/claude-fable-5-1',
  'anthropic/claude-sonnet-5-5',
  'openai/gpt-6-astra',
  'openai/gpt-5.6-sol',
  'deepseek/deepseek-v4-pro',
  'deepseek/deepseek-v4-flash',
]

const START_MODEL = 'Anthropic - claude-fable-5-1'
const PICKED_MODEL = 'DeepSeek - deepseek-v4-pro'
const PICKED_INDEX = 4

// Fuzzy match: the query's letters must appear in order, not necessarily next to each other.
const fuzzy = (query: string, text: string) => {
  const q = query.toLowerCase()
  let i = 0
  for (const c of text.toLowerCase()) {
    if (c === q[i]) i++
    if (i === q.length) return true
  }
  return q.length === 0
}

const PROMPT = 'fix the rendering issue'

type Step = { kind: Exclude<Line['kind'], 'user' | 'notice'>; text: string }
const STEPS: Step[] = [
  {
    kind: 'reasoning',
    text: 'The user says "fix the rendering issue" with no context. I should explore the repo to find what project we\'re in.',
  },
  { kind: 'cmd', text: 'pwd && ls -la' },
  { kind: 'cmd', text: 'git status 2>/dev/null | head -30' },
  {
    kind: 'reasoning',
    text: 'A Go TUI typing app. The rendering issue is probably in the game view. Let me look at View().',
  },
  { kind: 'cmd', text: 'grep -rn "View()" --include=*.go .' },
  {
    kind: 'assistant',
    text: 'Found it: View() redraws the board before the frame buffer is cleared, which causes the flicker. Clearing it first fixes the rendering issue.',
  },
]

const MODEL_NOTICE = `Model changed to: ${MODELS[PICKED_INDEX]}`

const finalLines: Line[] = [
  { id: 1, kind: 'notice', text: MODEL_NOTICE },
  { id: 2, kind: 'user', text: PROMPT },
  ...STEPS.map((s, i) => ({ id: 3 + i, ...s })),
]

const sleep = (ms: number) => new Promise<void>(resolve => setTimeout(resolve, ms))

export default function TuiPreview() {
  const sectionRef = useRef<HTMLElement>(null)
  const chatRef = useRef<HTMLDivElement>(null)
  const [visible, setVisible] = useState(false)
  const [reduced, setReduced] = useState(false)
  const [input, setInput] = useState('')
  const [lines, setLines] = useState<Line[]>([])
  const [working, setWorking] = useState(false)
  const [palette, setPalette] = useState<Palette | null>(null)
  const [model, setModel] = useState(START_MODEL)

  // Like a web chat: new content flows top to bottom, and the view follows once it overflows.
  useEffect(() => {
    const el = chatRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [lines, working])

  // Only animate while the section is on screen.
  useEffect(() => {
    setReduced(window.matchMedia('(prefers-reduced-motion: reduce)').matches)
    const el = sectionRef.current
    if (!el) return
    const io = new IntersectionObserver(([entry]) => setVisible(entry.isIntersecting), { threshold: 0.3 })
    io.observe(el)
    return () => io.disconnect()
  }, [])

  useEffect(() => {
    if (reduced) {
      setLines(finalLines)
      setInput('')
      setWorking(false)
      setPalette(null)
      setModel(PICKED_MODEL)
      return
    }
    if (!visible) return

    let cancelled = false
    let nextId = 1
    const wait = async (ms: number) => {
      await sleep(ms)
      if (cancelled) throw new Error('cancelled')
    }
    const push = (kind: Line['kind'], text: string) => {
      const id = nextId++
      setLines(l => [...l, { id, kind, text }])
      return id
    }
    const stream = async (kind: Line['kind'], text: string) => {
      const id = push(kind, '')
      for (let i = 3; i < text.length + 3; i += 3) {
        setLines(l => l.map(x => (x.id === id ? { ...x, text: text.slice(0, i) } : x)))
        await wait(20)
      }
    }

    const run = async () => {
      while (!cancelled) {
        setLines([])
        setInput('')
        setWorking(false)
        setPalette(null)
        setModel(START_MODEL)
        nextId = 1
        await wait(1400)

        // Scene 1: "/" opens the command list, typing narrows it with fuzzy matching
        setInput('/')
        setPalette({ mode: 'commands' })
        await wait(1500)
        for (const typed of ['/m', '/md', '/mdl']) {
          setInput(typed)
          await wait(1000)
        }
        await wait(400)

        // Enter runs /models and shows the model picker
        setInput('/models')
        setPalette({ mode: 'models', sel: 0 })
        await wait(1300)
        for (let sel = 1; sel <= PICKED_INDEX; sel++) {
          setPalette({ mode: 'models', sel })
          await wait(550)
        }
        await wait(500)

        // Enter picks the model, and the change is posted to the chat
        setModel(PICKED_MODEL)
        setPalette(null)
        setInput('')
        push('notice', MODEL_NOTICE)
        await wait(1500)

        // Scene 2: a normal prompt
        for (let i = 1; i <= PROMPT.length; i++) {
          setInput(PROMPT.slice(0, i))
          await wait(45 + Math.random() * 45)
        }
        await wait(500)

        setInput('')
        push('user', PROMPT)
        setWorking(true)
        await wait(900)

        for (const step of STEPS) {
          if (step.kind === 'assistant') setWorking(false)
          if (step.kind === 'cmd') {
            push('cmd', step.text)
            await wait(1000)
          } else {
            await stream(step.kind, step.text)
            await wait(500)
          }
        }

        await wait(5000)
      }
    }

    run().catch(() => {})
    return () => {
      cancelled = true
    }
  }, [visible, reduced])

  const matches = COMMANDS.filter(c => fuzzy(input.slice(1), c.name))

  return (
    <section ref={sectionRef} className="mx-auto w-full max-w-5xl px-6 py-24">
      <h2 className="mb-10 text-center font-['Playfair_Display'] text-4xl leading-[1.15] text-white md:text-5xl">
        Minimal. Clean. Simple.
      </h2>

      <div className="overflow-hidden rounded-sm border border-border bg-[#1c1713] shadow-2xl">
        <TerminalHeader title="bai" />

        <p className="sr-only">
          Animated preview: typing a slash opens a fuzzy-searchable command list, the models command lets you pick a model, and then a prompt is sent. bAI reasons about it, runs shell commands in the repo and replies with what it found.
        </p>

        <div aria-hidden className="relative flex h-[36rem] flex-col font-['JetBrains_Mono'] text-sm">
          <div className="absolute bottom-0 left-0 top-40 w-0.5 bg-[#e8742a]" />

          {/* main screen: hints while empty, chat once a message is sent */}
          <div className="flex min-h-0 flex-1 flex-col overflow-hidden pt-5">
            <p className="shrink-0 px-5 text-[#8a5a5a]">~/projects/my-app (main)</p>

            {lines.length === 0 && !working ? (
              <div className="mt-6 space-y-1 px-5">
                <p>
                  <span className="font-bold text-[#e8742a]">bai</span>{' '}
                  <span className="text-[#8a5a5a]">v0.0.1</span>
                </p>
                {hints.map(h => (
                  <p key={h.key}>
                    <span className="text-[#8a5a5a]">{h.key}: </span>
                    <span className="font-bold text-[#e8dc9a]">{h.action}</span>
                  </p>
                ))}
              </div>
            ) : (
              <div
                ref={chatRef}
                className="pointer-events-none mt-4 min-h-0 flex-1 space-y-4 overflow-hidden pb-3"
              >
                {lines.map(line => (
                  <TranscriptLine key={line.id} line={line} />
                ))}
                {working && (
                  <p className="px-5 text-[#8a5a5a]">
                    <span className="mr-2 inline-block animate-pulse">⠶</span>working...
                  </p>
                )}
              </div>
            )}
          </div>

          {/* input */}
          <div className="shrink-0 border-t border-border/80 px-5 py-3">
            <div className="flex items-center border-l border-white/70 pl-3">
              {input ? (
                <>
                  <span className="text-white">{input}</span>
                  <span className="ml-0.5 inline-block h-4 w-2 animate-pulse bg-white motion-reduce:animate-none" />
                </>
              ) : (
                <span className="text-muted-foreground/70">
                  <span className="animate-pulse bg-white text-[#1c1713] motion-reduce:animate-none">S</span>
                  end a message...
                </span>
              )}
            </div>
          </div>

          {/* command list / model picker, opens below the input */}
          {palette && (
            <div className="h-[13.5rem] shrink-0 space-y-1 overflow-hidden border-t border-border/80 px-5 py-4">
              {palette.mode === 'commands'
                ? matches.map((c, i) => (
                    <PaletteRow key={c.name} selected={i === 0}>
                      /{c.name} - {c.desc}
                    </PaletteRow>
                  ))
                : MODELS.map((m, i) => (
                    <PaletteRow key={m} selected={i === palette.sel}>
                      {m}
                    </PaletteRow>
                  ))}
              {palette.mode === 'models' && (
                <p className="pl-4 pt-3 text-xs tracking-[0.3em] text-[#5a4a4a]">
                  <span className="text-muted-foreground">●</span>○○○○
                </p>
              )}
            </div>
          )}

          <div className="shrink-0 border-t border-border/80 px-5 py-3 text-[#8a5a5a]">{model}</div>
        </div>
      </div>

      <p className="mt-4 text-center text-sm text-muted-foreground">
        Type <span className="font-['JetBrains_Mono'] text-foreground">/</span> for commands, or{' '}
        <span className="font-['JetBrains_Mono'] text-foreground">@</span> to find files.
      </p>
    </section>
  )
}

function PaletteRow({ selected, children }: { selected: boolean; children: ReactNode }) {
  return (
    <p className={`truncate whitespace-pre ${selected ? 'font-bold text-[#e060e0]' : 'text-foreground'}`}>
      {selected ? '> ' : '  '}
      {children}
    </p>
  )
}

function TranscriptLine({ line }: { line: Line }) {
  switch (line.kind) {
    case 'notice':
      return <p className="px-5 text-[#e0a64a]">{line.text}</p>
    case 'user':
      return <p className="bg-[#26221e] px-5 py-5 text-white">{line.text}</p>
    case 'reasoning':
      return <p className="px-5 leading-relaxed text-[#8a5a5a]">{line.text}</p>
    case 'cmd':
      return (
        <p className="px-5 font-bold text-[#e8742a]">
          $ {line.text}
        </p>
      )
    case 'assistant':
      return <p className="px-5 leading-relaxed text-foreground">{line.text}</p>
  }
}
