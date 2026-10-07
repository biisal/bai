import { useEffect, useState, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { installCmd, repoUrl, siteName } from '@/config'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { cn } from '@/lib/utils'

type Doc = { id: string; title: string; body: ReactNode }

const docs: { group: string; items: Doc[] }[] = [
  {
    group: 'Getting Started',
    items: [
      {
        id: 'introduction',
        title: 'Introduction',
        body: (
          <>
            <P>
              {siteName} is a terminal-based AI coding agent. Single binary, TUI interface. Status:
              early development — the core chat loop and coding agent tools work.
            </P>
            <H2>Features</H2>
            <Ul>
              <li>TUI with provider and model selection</li>
              <li>Streaming chat against OpenAI, Anthropic, or Groq (any OpenAI-compatible endpoint)</li>
              <li>SQLite-backed sessions (conversations)</li>
              <li>File read / write / edit tools, bash command execution</li>
              <li>Diff view of file changes</li>
              <li>OpenCode variant support</li>
            </Ul>
            <H2>Requirements</H2>
            <Ul>
              <li>Go 1.26+ (to build from source)</li>
              <li>An API key for at least one supported provider</li>
            </Ul>
          </>
        ),
      },
      {
        id: 'installation',
        title: 'Installation',
        body: (
          <>
            <P>
              One static binary, ~14 MB. No runtime, no model downloads.
            </P>
            <H2>With curl</H2>
            <Code>{installCmd}</Code>
            <H2>From source</H2>
            <Code>{`git clone ${repoUrl} && cd bai
make build      # builds bin/bai
make run        # build and run`}</Code>
            <H2>Make targets</H2>
            <Code>{`make build      build the binary into bin/bai
make run        build and run
make dev        run in dev mode
make watch      hot reload on go file changes
make test       run the test suite
make clean      remove build artifacts`}</Code>
          </>
        ),
      },
    ],
  },
  {
    group: 'Guides',
    items: [
      {
        id: 'configuration',
        title: 'Configuration',
        body: (
          <>
            <P>
              Config lives at <Code inline>~/.config/bai/config.json</Code>. Copy the example and
              fill in your API keys:
            </P>
            <Code>{`mkdir -p ~/.config/bai
cp config.example.json ~/.config/bai/config.json`}</Code>
            <H2>Key fields</H2>
            <Code>{`{
  "database_path": "~/.config/bai/bai.db",
  "log_file_path": "~/.config/bai/bai.log",
  "providers": [
    {
      "name": "Anthropic",
      "format": "anthropic",
      "api_key": "sk-ant-...",
      "base_url": "https://api.anthropic.com",
      "models": [{ "id": "claude-sonnet-4-20250514", "context": 200000 }]
    }
  ]
}`}</Code>
            <Ul>
              <li>
                <Code inline>format</Code> — <Code inline>anthropic</Code> or{' '}
                <Code inline>openai-compatible</Code>
              </li>
              <li>
                <Code inline>variant</Code> — optional, e.g. <Code inline>opencode</Code>
              </li>
              <li>
                Point elsewhere with <Code inline>-config path/to.json</Code>
              </li>
            </Ul>
          </>
        ),
      },
      {
        id: 'usage',
        title: 'Usage',
        body: (
          <>
            <Code>{`bai                        use the default config
bai -config path/to.json   use a specific config`}</Code>
            <P>
              The TUI opens with provider and model selection. From there you chat with streaming
              responses, and the agent can read, write, and edit files, run bash commands, and show
              diffs of what it changed. Conversations are persisted to SQLite, so sessions are
              resumable.
            </P>
          </>
        ),
      },
    ],
  },
  {
    group: 'Reference',
    items: [
      {
        id: 'providers',
        title: 'Providers',
        body: (
          <>
            <P>Any of these can be configured via config.json:</P>
            <Code>{`OpenAI      openai-compatible  https://api.openai.com/v1
Anthropic   anthropic          https://api.anthropic.com
DeepSeek    openai-compatible  https://api.deepseek.com/v1
Qwen        openai-compatible  https://dashscope.aliyuncs.com/compatible-mode/v1
Groq        openai-compatible  https://api.groq.com/openai/v1
OpenCode    openai-compatible  https://opencode.ai/zen/v1   (variant: opencode)`}</Code>
            <P>
              Any other OpenAI-compatible endpoint works too — point <Code inline>base_url</Code> at
              it and list your models.
            </P>
          </>
        ),
      },
      {
        id: 'project-layout',
        title: 'Project Layout',
        body: (
          <Code>{`cmd/bai/         entry point and start command
internal/agent/  provider gateway and message handling
internal/tui/    bubble tea model, view, update
internal/db/     sqlite connection and queries
internal/config/ config loading and defaults`}</Code>
        ),
      },
    ],
  },
]

function P({ children }: { children: ReactNode }) {
  return <p className="mb-4 leading-relaxed text-neutral-300">{children}</p>
}

function H2({ children }: { children: ReactNode }) {
  return <h2 className="mt-8 mb-3 text-xl font-semibold text-white">{children}</h2>
}

function Ul({ children }: { children: ReactNode }) {
  return (
    <ul className="mb-4 list-disc space-y-1.5 pl-5 leading-relaxed text-neutral-300">{children}</ul>
  )
}

function Code({ children, inline }: { children: ReactNode; inline?: boolean }) {
  const cls =
    "font-['JetBrains_Mono'] text-xs whitespace-pre text-white/90"
  return inline ? (
    <code className={cn('rounded-sm border border-border bg-muted/40 px-1.5 py-0.5', cls)}>
      {children}
    </code>
  ) : (
    <pre
      className={cn(
        'my-4 overflow-x-auto rounded-sm border border-border bg-muted/20 p-4 leading-relaxed',
        cls,
      )}
    >
      {children}
    </pre>
  )
}

export default function Docs() {
  const [active, setActive] = useState(docs[0].items[0].id)

  useEffect(() => {
    window.scrollTo(0, 0)
  }, [active])

  const current = docs.flatMap(g => g.items).find(d => d.id === active)!

  return (
    <div className="min-h-screen bg-background text-foreground antialiased selection:bg-primary/20 selection:text-primary">
      <SidebarProvider>
        <Sidebar>
          <SidebarHeader>
            <div className="px-2 py-1.5 font-['JetBrains_Mono'] text-xs uppercase tracking-[0.2em] text-muted-foreground">
              {siteName} docs
            </div>
          </SidebarHeader>
          <SidebarContent>
            {docs.map(group => (
              <SidebarGroup key={group.group}>
                <SidebarGroupLabel>{group.group}</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    {group.items.map(doc => (
                      <SidebarMenuItem key={doc.id}>
                        <SidebarMenuButton
                          isActive={doc.id === active}
                          onClick={() => setActive(doc.id)}
                        >
                          {doc.title}
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    ))}
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            ))}
          </SidebarContent>
        </Sidebar>

        <SidebarInset>
          <header className="sticky top-0 z-40 flex h-14 items-center justify-between border-b border-border bg-background/80 px-4 backdrop-blur-sm">
            <div className="flex items-center gap-3">
              <SidebarTrigger />
              <Link
                to="/"
                className="flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                <ArrowLeft className="w-4 h-4" />
                {siteName}
              </Link>
            </div>
            <span className="font-['JetBrains_Mono'] text-xs text-muted-foreground">docs</span>
          </header>

          <div className="mx-auto w-full max-w-3xl flex-1 px-6 py-10">
            <h1 className="mb-6 text-3xl font-['Playfair_Display'] text-white sm:text-4xl">
              {current.title}
            </h1>
            {current.body}
          </div>
        </SidebarInset>
      </SidebarProvider>
    </div>
  )
}
